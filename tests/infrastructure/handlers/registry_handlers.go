package handlers

import "net/http"

// RegistryTypeHandler is a function that handles the Registry type selection page
func RegistryTypeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	if r.Method == post {
		selection := r.FormValue("selection")

		data := struct {
			Selection string
		}{
			Selection: selection,
		}

		Templates.ExecuteTemplate(w, "registrytype", data)

		return
	} else {
		Templates.ExecuteTemplate(w, "registrytype", nil)
	}
}
