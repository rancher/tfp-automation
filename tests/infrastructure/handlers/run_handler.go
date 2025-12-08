package handlers

import (
	"net/http"

	retrieve "github.com/rancher/tfp-automation/tests/infrastructure/formCookie"
)

// RunHandler is a function that handles the run action
func RunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		selection := retrieve.GetFormOrCookie(r, "selection")
		clustertype := retrieve.GetFormOrCookie(r, "clustertype")
		ranchertype := retrieve.GetFormOrCookie(r, "ranchertype")
		installtype := retrieve.GetFormOrCookie(r, "installtype")
		provider := retrieve.GetFormOrCookie(r, "provider")
		providerversion := retrieve.GetFormOrCookie(r, "providerversion")
		confirm := retrieve.GetFormOrCookie(r, "confirm")

		if selection == "" && clustertype == "" && ranchertype == "" && installtype == "" && provider == "" && providerversion == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		http.SetCookie(w, &http.Cookie{Name: "selection", Value: selection, Path: "/"})
		http.SetCookie(w, &http.Cookie{Name: "provider", Value: provider, Path: "/"})
		http.SetCookie(w, &http.Cookie{Name: "providerversion", Value: providerversion, Path: "/"})
		http.SetCookie(w, &http.Cookie{Name: "confirm", Value: confirm, Path: "/"})

		http.Redirect(w, r, "/status", http.StatusSeeOther)
	}
}
