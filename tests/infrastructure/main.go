package main

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pkg/browser"
	"github.com/rancher/tfp-automation/tests/infrastructure/cli"
	"github.com/rancher/tfp-automation/tests/infrastructure/handlers"
	"github.com/sirupsen/logrus"
)

func main() {
	if os.Args[1] == "--web" {
		http.HandleFunc("/", handlers.WelcomeHandler)
		http.HandleFunc("/selection", handlers.ClusterOrRancherHandler)
		http.HandleFunc("/clustertype", handlers.ClusterTypeHandler)
		http.HandleFunc("/ranchertype", handlers.RancherTypeHandler)
		http.HandleFunc("/installtype", handlers.InstallTypeHandler)
		http.HandleFunc("/provider", handlers.ProviderHandler)
		http.HandleFunc("/providerversion", handlers.ProviderVersionHandler)
		http.HandleFunc("/run", handlers.RunHandler)
		http.HandleFunc("/confirm", handlers.ConfirmHandler)
		http.HandleFunc("/status", handlers.StatusHandler)

		browser.OpenURL("http://localhost:8080")
		logrus.Fatal(http.ListenAndServe(":8080", nil))
	}

	os.Exit(cli.RunCLI())
}

func init() {
	_, filename, _, _ := runtime.Caller(0)
	staticDir := filepath.Join(filepath.Dir(filename), "static")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
}
