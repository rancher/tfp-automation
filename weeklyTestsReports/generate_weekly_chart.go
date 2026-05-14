package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	headerPath   = "weeklyReports/assets/workflow_header.html"
	templatePath = "weeklyReports/assets/template.html"

	assetsSrcDir = "weeklyReports/assets"
	assetsDstDir = "results/assets"
	resultsDir   = "results"

	weeklySummary = "weekly_summary.html"
)

type TestResult struct {
	Date     string `json:"date"`
	Status   string `json:"status"`
	Workflow string `json:"workflow"`
	Job      string `json:"job"`
}

func main() {
	err := copyAssets(assetsSrcDir, assetsDstDir)
	if err != nil {
		logrus.Error("Error copying assets:", err)
		os.Exit(1)
	}

	htmlPath := filepath.Join(resultsDir, weeklySummary)
	workflowMap := loadTestResults(resultsDir)

	var chartsHTML strings.Builder
	for workflow, trs := range workflowMap {
		sort.Slice(trs, func(i, j int) bool {
			return trs[i].Date < trs[j].Date
		})
		renderScatterPlotChart(&chartsHTML, workflow, trs)
	}

	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		logrus.Error("Error reading template.html:", err)
		os.Exit(1)
	}

	templateStr := string(templateBytes)
	finalHTML := strings.Replace(templateStr, "<!--CHARTS_PLACEHOLDER-->", chartsHTML.String(), 1)

	f, err := os.Create(htmlPath)
	if err != nil {
		logrus.Error("Error creating HTML file:", err)
		os.Exit(1)
	}

	defer f.Close()

	f.WriteString(finalHTML)
}

// copyAssets copies all files from srcDir to dstDir (non-recursive, for style.css, toggle.js, rancher.png)
func copyAssets(srcDir, dstDir string) error {
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	files := []string{"style.css", "toggle.js", "rancher.png"}
	for _, f := range files {
		src := filepath.Join(srcDir, f)
		dst := filepath.Join(dstDir, f)

		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}

		if err := os.WriteFile(dst, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

// loadTestResults is a helper function that reads all JSON test result files from the given directory and
// returns a workflow map.
func loadTestResults(resultsDir string) map[string][]TestResult {
	files, err := os.ReadDir(resultsDir)
	if err != nil {
		logrus.Error("Error reading results directory:", err)
		os.Exit(1)
	}

	var results []TestResult
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			data, err := os.ReadFile(filepath.Join(resultsDir, file.Name()))
			if err != nil {
				continue
			}

			var tr TestResult
			if err := json.Unmarshal(data, &tr); err != nil {
				continue
			}

			if tr.Workflow == "" {
				base := strings.TrimSuffix(file.Name(), ".json")
				parts := strings.Split(base, "-")

				if len(parts) > 2 {
					tr.Workflow = strings.Join(parts[2:len(parts)-1], "-")
				} else {
					tr.Workflow = base
				}
			}

			results = append(results, tr)
		}
	}

	workflowMap := make(map[string][]TestResult)
	for _, r := range results {
		workflowMap[r.Workflow] = append(workflowMap[r.Workflow], r)
	}

	return workflowMap
}

// renderScatterPlotChart is a helper function that takes a writer, workflow name, and test results to render a
// scatter plot chart for the given workflow.
func renderScatterPlotChart(w io.Writer, workflow string, trs []TestResult) {
	jobSet := make(map[string]struct{})
	dateSet := make(map[string]struct{})
	jobStats := make(map[string][2]int)

	for _, r := range trs {
		jobSet[r.Job] = struct{}{}
		dateSet[r.Date] = struct{}{}
		stat := jobStats[r.Job]

		if strings.ToLower(r.Status) == "success" {
			stat[0]++
		}

		stat[1]++
		jobStats[r.Job] = stat
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

	passData := []opts.ScatterData{}
	failData := []opts.ScatterData{}
	jobIdx := make(map[string]int)

	for i, job := range jobs {
		jobIdx[job] = i
	}

	dateIdx := make(map[string]int)
	for i, date := range dates {
		dateIdx[date] = i
	}

	for _, r := range trs {
		point := opts.ScatterData{
			Name:       fmt.Sprintf("%s", r.Date),
			Value:      []interface{}{dateIdx[r.Date], jobIdx[r.Job]},
			SymbolSize: 18,
		}
		if strings.ToLower(r.Status) == "success" {
			passData = append(passData, point)
		} else {
			failData = append(failData, point)
		}
	}

	scatter := charts.NewScatter()
	scatter.SetGlobalOptions(
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
		charts.WithXAxisOpts(opts.XAxis{Type: "category", Data: dates, Name: "Date"}),
		charts.WithYAxisOpts(opts.YAxis{Type: "category", Data: jobLabels, Name: "Job"}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
		charts.WithInitializationOpts(opts.Initialization{Width: "1200px", Height: "400px"}),
	)
	scatter.AddSeries("Pass", passData, charts.WithItemStyleOpts(opts.ItemStyle{Color: "#2ca02c"}))
	scatter.AddSeries("Fail", failData, charts.WithItemStyleOpts(opts.ItemStyle{Color: "#d62728"}))

	summary := formatWorkflow(workflow)
	divID := fmt.Sprintf("scatter-%s", strings.ReplaceAll(workflow, " ", "-"))
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

	scatter.Render(w)
	w.Write([]byte("</div></div>"))
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
