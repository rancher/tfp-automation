package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	headerPath   = "testReports/assets/workflow_header.html"
	templatePath = "testReports/assets/template.html"

	resultsDir          = "results"
	testSummaryFile     = "test_summary.html"
	historyOutputPath   = "results/history/test_history.json"
	defaultRetentionDay = 365
)

type TestResult struct {
	Date      string `json:"date"`
	Status    string `json:"status"`
	Workflow  string `json:"workflow"`
	Job       string `json:"job"`
	Timestamp string `json:"timestamp,omitempty"`
}

type datedResult struct {
	TestResult
	ObservedAt time.Time
}

func main() {
	if err := os.MkdirAll(resultsDir, 0o755); err != nil {
		logrus.Fatal(err)
	}

	htmlPath := filepath.Join(resultsDir, testSummaryFile)
	results := loadTestResults(resultsDir)
	results = dedupeResults(results)
	results = applyRetention(results, resolveRetentionDays())
	writeHistory(results, historyOutputPath)
	workflowMap := groupByWorkflow(results)

	var chartsHTML strings.Builder

	workflowNames := make([]string, 0, len(workflowMap))
	for workflow := range workflowMap {
		workflowNames = append(workflowNames, workflow)
	}

	sort.Strings(workflowNames)

	for _, workflow := range workflowNames {
		trs := workflowMap[workflow]
		sort.Slice(trs, func(i, j int) bool {
			return trs[i].ObservedAt.Before(trs[j].ObservedAt)
		})
		renderPassRateTrendChart(&chartsHTML, workflow, trs)
	}

	cssBytes, err := os.ReadFile("testReports/assets/style.css")
	if err != nil {
		logrus.Fatal(err)
	}

	jsBytes, err := os.ReadFile("testReports/assets/toggle.js")
	if err != nil {
		logrus.Fatal(err)
	}

	pngBytes, err := os.ReadFile("testReports/assets/rancher.png")
	if err != nil {
		logrus.Fatal(err)
	}

	pngBase64 := base64.StdEncoding.EncodeToString(pngBytes)

	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		logrus.Error("Error reading template.html:", err)
		os.Exit(1)
	}

	finalHTML := string(templateBytes)
	finalHTML = strings.ReplaceAll(finalHTML, "__INLINE_CSS__", string(cssBytes))
	finalHTML = strings.ReplaceAll(finalHTML, "__INLINE_JS__", string(jsBytes))
	finalHTML = strings.ReplaceAll(finalHTML, "__RANCHER_LOGO__", pngBase64)
	finalHTML = strings.Replace(finalHTML, "<!--CHARTS_PLACEHOLDER-->", chartsHTML.String(), 1)

	f, err := os.Create(htmlPath)
	if err != nil {
		logrus.Error("Error creating HTML file:", err)
		os.Exit(1)
	}
	defer f.Close()

	_, _ = f.WriteString(finalHTML)
}

// loadTestResults reads JSON test result files recursively from the results directory.
func loadTestResults(root string) []datedResult {
	var results []datedResult

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() || filepath.Ext(d.Name()) != ".json" {
			return nil
		}

		loaded, readErr := readResultsFromJSON(path)
		if readErr != nil {
			return nil
		}

		results = append(results, loaded...)
		return nil
	})
	if err != nil {
		logrus.Error("Error reading results directory:", err)
		os.Exit(1)
	}

	return results
}

func readResultsFromJSON(path string) ([]datedResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var single TestResult
	if err := json.Unmarshal(data, &single); err == nil && (single.Workflow != "" || single.Job != "") {
		return normalizeResults([]TestResult{single}, path), nil
	}

	var many []TestResult
	if err := json.Unmarshal(data, &many); err == nil {
		return normalizeResults(many, path), nil
	}

	return nil, fmt.Errorf("unsupported JSON payload in %s", path)
}

func normalizeResults(items []TestResult, sourcePath string) []datedResult {
	clean := make([]datedResult, 0, len(items))

	for _, tr := range items {
		if tr.Workflow == "" {
			base := strings.TrimSuffix(filepath.Base(sourcePath), ".json")
			parts := strings.Split(base, "-")

			if len(parts) > 2 {
				tr.Workflow = strings.Join(parts[2:len(parts)-1], "-")
			} else {
				tr.Workflow = base
			}
		}

		if tr.Workflow == "" || tr.Job == "" {
			continue
		}

		observedAt := parseObservedAt(tr)
		tr.Status = strings.ToLower(strings.TrimSpace(tr.Status))
		if tr.Status == "" {
			tr.Status = "unknown"
		}

		if tr.Timestamp == "" {
			tr.Timestamp = observedAt.UTC().Format(time.RFC3339)
		}

		clean = append(clean, datedResult{TestResult: tr, ObservedAt: observedAt})
	}

	return clean
}

func parseObservedAt(tr TestResult) time.Time {
	if tr.Timestamp != "" {
		if ts, err := time.Parse(time.RFC3339, tr.Timestamp); err == nil {
			return ts
		}
	}

	layouts := []string{
		"January 02, 2006 at 03:04 PM",
		time.RFC3339,
		"2006-01-02",
	}

	for _, layout := range layouts {
		if ts, err := time.Parse(layout, tr.Date); err == nil {
			return ts
		}
	}

	return time.Now().UTC()
}

func dedupeResults(items []datedResult) []datedResult {
	seen := make(map[string]struct{}, len(items))
	out := make([]datedResult, 0, len(items))

	for _, item := range items {
		key := fmt.Sprintf("%s|%s|%s|%s", item.Workflow, item.Job, item.Timestamp, item.Status)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}

	return out
}

func resolveRetentionDays() int {
	value := strings.TrimSpace(os.Getenv("HISTORICAL_RETENTION_DAYS"))
	if value == "" {
		return defaultRetentionDay
	}

	days, err := strconv.Atoi(value)
	if err != nil || days <= 0 {
		return defaultRetentionDay
	}

	return days
}

func applyRetention(items []datedResult, retentionDays int) []datedResult {
	cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays)
	out := make([]datedResult, 0, len(items))

	for _, item := range items {
		if item.ObservedAt.Before(cutoff) {
			continue
		}
		out = append(out, item)
	}

	return out
}

func writeHistory(items []datedResult, outputPath string) {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		logrus.Fatal(err)
	}

	output := make([]TestResult, 0, len(items))
	for _, item := range items {
		t := item.TestResult
		if t.Timestamp == "" {
			t.Timestamp = item.ObservedAt.UTC().Format(time.RFC3339)
		}

		output = append(output, t)
	}

	encoded, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		logrus.Fatal(err)
	}

	if err := os.WriteFile(outputPath, encoded, 0o644); err != nil {
		logrus.Fatal(err)
	}
}

func groupByWorkflow(items []datedResult) map[string][]datedResult {
	workflowMap := make(map[string][]datedResult)
	for _, r := range items {
		workflowMap[r.Workflow] = append(workflowMap[r.Workflow], r)
	}

	return workflowMap
}

// renderPassRateTrendChart renders a long-term pass-rate trend line chart for each workflow.
func renderPassRateTrendChart(w io.Writer, workflow string, trs []datedResult) {
	jobSet := make(map[string]struct{})
	dateSet := make(map[string]struct{})
	jobStats := make(map[string][2]int)
	totalByDate := make(map[string][2]int)
	jobByDate := make(map[string]map[string][2]int)

	for _, r := range trs {
		jobSet[r.Job] = struct{}{}
		day := r.ObservedAt.UTC().Format("2006-01-02")
		dateSet[day] = struct{}{}

		stat := jobStats[r.Job]
		dateStat := totalByDate[day]
		if _, ok := jobByDate[r.Job]; !ok {
			jobByDate[r.Job] = make(map[string][2]int)
		}
		jobDateStat := jobByDate[r.Job][day]

		if r.Status == "success" {
			stat[0]++
			dateStat[0]++
			jobDateStat[0]++
		}
		stat[1]++
		dateStat[1]++
		jobDateStat[1]++

		jobStats[r.Job] = stat
		totalByDate[day] = dateStat
		jobByDate[r.Job][day] = jobDateStat
	}

	jobs := make([]string, 0, len(jobSet))
	for job := range jobSet {
		jobs = append(jobs, job)
	}

	sort.Strings(jobs)

	jobLabels := make([]string, len(jobs))
	for i, job := range jobs {
		stat := jobStats[job]
		percent := 0
		if stat[1] > 0 {
			percent = int(float64(stat[0]) / float64(stat[1]) * 100)
		}
		jobLabels[i] = fmt.Sprintf("%s • %d%% pass", job, percent)
	}

	dates := make([]string, 0, len(dateSet))
	for date := range dateSet {
		dates = append(dates, date)
	}

	sort.Strings(dates)

	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "slider", Start: 0, End: 100}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "inside"}),
		charts.WithXAxisOpts(opts.XAxis{Type: "category", Data: dates, Name: "Date"}),
		charts.WithYAxisOpts(opts.YAxis{Type: "value", Name: "Pass Rate (%)", Min: 0, Max: 100}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
		charts.WithInitializationOpts(opts.Initialization{Width: "1200px", Height: "400px"}),
	)

	overallSeries := make([]opts.LineData, 0, len(dates))
	for _, date := range dates {
		stat := totalByDate[date]
		percent := 0.0
		if stat[1] > 0 {
			percent = float64(stat[0]) / float64(stat[1]) * 100
		}

		overallSeries = append(overallSeries, opts.LineData{Value: percent})
	}

	line.AddSeries(
		"Overall pass rate",
		overallSeries,
		charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}),
		charts.WithItemStyleOpts(opts.ItemStyle{Color: "#2563eb"}),
	)

	for i, job := range jobs {
		series := make([]opts.LineData, 0, len(dates))
		for _, date := range dates {
			stat := jobByDate[job][date]
			if stat[1] == 0 {
				series = append(series, opts.LineData{Value: nil})
				continue
			}
			percent := float64(stat[0]) / float64(stat[1]) * 100
			series = append(series, opts.LineData{Value: percent})
		}

		line.AddSeries(
			jobLabels[i],
			series,
			charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}),
		)
	}

	summary := formatWorkflow(workflow)
	divID := fmt.Sprintf("trend-%s", strings.ReplaceAll(workflow, " ", "-"))
	btnID := fmt.Sprintf("btn-%s", strings.ReplaceAll(workflow, " ", "-"))

	headertmpl, err := template.ParseFiles(headerPath)
	if err != nil {
		logrus.Error("Error reading workflow_header.html:", err)
		return
	}

	headertmpl.Execute(w, map[string]string{
		"WorkflowName": summary,
		"DivID":        divID,
		"BtnID":        btnID,
	})

	line.Render(w)
	_, _ = w.Write([]byte("</div></div>"))
}

// formatWorkflow is a helper function that takes a raw workflow name and formats it for better readability in the chart titles.
func formatWorkflow(raw string) string {
	parts := strings.Split(raw, "-")
	var version, name string

	for i, p := range parts {
		if strings.HasPrefix(p, "v") && i+1 < len(parts) && isNumeric(parts[i+1]) {
			version = p + "." + parts[i+1]
			name = strings.Join(parts[:i], " ")
			break
		}
	}

	if version == "" {
		return cases.Title(language.English).String(strings.ReplaceAll(raw, "-", " "))
	}

	name = cases.Title(language.English).String(strings.ReplaceAll(name, "-", " "))
	return fmt.Sprintf("[%s] %s", version, name)
}

// isNumeric is a helper function that checks if a string consists entirely of numeric characters.
func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}
