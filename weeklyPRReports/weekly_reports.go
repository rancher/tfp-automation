package main

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
)

type PR struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	State string `json:"state"`
}

func main() {
	if len(os.Args) != 2 {
		logrus.Error("Usage: weekly_reports <prs.json>")
		os.Exit(1)
	}

	prsPath := os.Args[1]

	data, err := os.ReadFile(prsPath)
	if err != nil {
		logrus.Errorf("Error reading %s: %v", prsPath, err)
		os.Exit(1)
	}

	var prs []PR
	if err := json.Unmarshal(data, &prs); err != nil {
		logrus.Errorf("Error parsing JSON: %v", err)
		os.Exit(1)
	}
}
