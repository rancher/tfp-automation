package handlers

import "net/http"

// ClusterTypeHandler is a function that handles the cluster type selection page
func ClusterTypeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	if r.Method == "POST" {
		selection := r.FormValue("selection")

		data := struct {
			Selection string
		}{
			Selection: selection,
		}

		Templates.ExecuteTemplate(w, "clustertype", data)

		return
	} else {
		Templates.ExecuteTemplate(w, "clustertype", nil)
	}
}
