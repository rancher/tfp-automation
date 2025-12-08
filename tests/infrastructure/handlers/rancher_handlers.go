package handlers

import "net/http"

// RancherTypeHandler is a function that handles the Rancher type selection page
func RancherTypeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	if r.Method == "POST" {
		selection := r.FormValue("selection")

		data := struct {
			Selection string
		}{
			Selection: selection,
		}

		Templates.ExecuteTemplate(w, "ranchertype", data)

		return
	} else {
		Templates.ExecuteTemplate(w, "ranchertype", nil)
	}
}

// InstallTypeHandler is a function that handles the install type selection page
func InstallTypeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	if r.Method == "POST" {
		selection := r.FormValue("selection")
		ranchertype := r.FormValue("ranchertype")

		var installOptions []string

		data := struct {
			Selection      string
			RancherType    string
			InstallOptions []string
		}{
			Selection:      selection,
			RancherType:    ranchertype,
			InstallOptions: installOptions,
		}

		// Dual-stack, IPv6, and Registry Rancher only support fresh installs
		switch ranchertype {
		case "dual", "ipv6", "registry":
			installOptions = []string{"fresh"}
		default:
			installOptions = []string{"fresh", "upgrade"}
		}

		Templates.ExecuteTemplate(w, "installtype", data)
	} else {
		Templates.ExecuteTemplate(w, "installtype", nil)
	}
}
