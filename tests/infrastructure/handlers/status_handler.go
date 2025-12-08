package handlers

import (
	"html/template"
	"net/http"
	"strings"

	retrieve "github.com/rancher/tfp-automation/tests/infrastructure/formCookie"
	share "github.com/rancher/tfp-automation/tests/infrastructure/state"
)

// StatusHandler is a function that serves the status page
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	selection := retrieve.GetFormOrCookie(r, "selection")
	clustertype := retrieve.GetFormOrCookie(r, "clustertype")
	ranchertype := retrieve.GetFormOrCookie(r, "ranchertype")
	installtype := retrieve.GetFormOrCookie(r, "installtype")
	provider := retrieve.GetFormOrCookie(r, "provider")
	providerversion := retrieve.GetFormOrCookie(r, "providerversion")
	confirm := retrieve.GetFormOrCookie(r, "confirm")

	if selection == "" && clustertype == "" && ranchertype == "" && installtype == "" && provider == "" && providerversion == "" && confirm == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	share.State.Mutex.Lock()
	defer share.State.Mutex.Unlock()

	data := struct {
		StageMsg template.HTML
		Error    string
	}{
		StageMsg: template.HTML(strings.ReplaceAll(share.State.StageMsg, "\n", "<br>")),
		Error:    share.State.ErrorMsg,
	}

	Templates.ExecuteTemplate(w, "status", data)
}
