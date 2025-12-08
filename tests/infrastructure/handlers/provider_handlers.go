package handlers

import "net/http"

// ProviderHandler is a function that handles the provider selection page
func ProviderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	if r.Method == "POST" {
		selection := r.FormValue("selection")
		clustertype := r.FormValue("clustertype")
		ranchertype := r.FormValue("ranchertype")
		installtype := r.FormValue("installtype")

		var installOptions []string

		if clustertype != "" {
			switch clustertype {
			case "airgap-rke2", "dual-rke2", "dual-k3s", "ipv6-rke2", "ipv6-k3s":
				installOptions = []string{"aws"}
			default:
				installOptions = []string{"aws", "linode", "vsphere"}
			}
		} else if ranchertype != "" {
			switch ranchertype {
			case "airgap", "dual", "ipv6", "registry":
				installOptions = []string{"aws"}
			case "proxy":
				installOptions = []string{"aws", "linode"}
			default:
				installOptions = []string{"aws", "linode", "vsphere"}
			}
		} else {
			installOptions = []string{}
		}

		data := struct {
			Selection      string
			ClusterType    string
			RancherType    string
			InstallType    string
			InstallOptions []string
		}{
			Selection:      selection,
			ClusterType:    clustertype,
			RancherType:    ranchertype,
			InstallType:    installtype,
			InstallOptions: installOptions,
		}

		Templates.ExecuteTemplate(w, "provider", data)

		return
	}

	Templates.ExecuteTemplate(w, "provider", nil)
}

// ProviderVersionHandler is a function that handles the version selection page
func ProviderVersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	if r.Method == "POST" {
		selection := r.FormValue("selection")
		clustertype := r.FormValue("clustertype")
		ranchertype := r.FormValue("ranchertype")
		installtype := r.FormValue("installtype")
		provider := r.FormValue("provider")

		data := struct {
			Selection   string
			ClusterType string
			RancherType string
			InstallType string
			Provider    string
		}{
			Selection:   selection,
			ClusterType: clustertype,
			RancherType: ranchertype,
			InstallType: installtype,
			Provider:    provider,
		}

		Templates.ExecuteTemplate(w, "providerversion", data)
	} else {
		Templates.ExecuteTemplate(w, "providerversion", nil)
	}
}
