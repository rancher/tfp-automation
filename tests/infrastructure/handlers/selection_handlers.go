package handlers

import "net/http"

// WelcomeHandler is a function that handles the welcome page
func WelcomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	Templates.ExecuteTemplate(w, "welcome", nil)
}

// ClusterOrRancherHandler is a function that handles the cluster or Rancher selection page
func ClusterOrRancherHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	Templates.ExecuteTemplate(w, "selection", nil)
}
