package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	retrieve "github.com/rancher/tfp-automation/tests/infrastructure/formCookie"
	mask "github.com/rancher/tfp-automation/tests/infrastructure/maskFields"
	share "github.com/rancher/tfp-automation/tests/infrastructure/state"
	webConfig "github.com/rancher/tfp-automation/tests/infrastructure/updateWebConfig"
	"github.com/rancher/tfp-automation/tests/infrastructure/web"
)

// ConfirmHandler displays config for review and allows edit/confirm
func ConfirmHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	selection := retrieve.GetFormOrCookie(r, "selection")
	clustertype := retrieve.GetFormOrCookie(r, "clustertype")
	ranchertype := retrieve.GetFormOrCookie(r, "ranchertype")
	installtype := retrieve.GetFormOrCookie(r, "installtype")
	provider := retrieve.GetFormOrCookie(r, "provider")
	providerversion := retrieve.GetFormOrCookie(r, "providerversion")

	if clustertype == "" && ranchertype == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if ranchertype != "" && installtype == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if selection == "" || provider == "" || providerversion == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	editMode := false
	if r.Method == "POST" && r.FormValue("action") == "edit" {
		editMode = true
	}

	if r.Method == "POST" && r.FormValue("action") == "save" {
		webConfig.UpdateConfigFromForm(r.PostForm, rancherConfig)
		webConfig.UpdateConfigFromForm(r.PostForm, terraformConfig)
		webConfig.UpdateConfigFromForm(r.PostForm, terratestConfig)
		webConfig.UpdateConfigFromForm(r.PostForm, standaloneConfig)

		editMode = false
	}

	rancherConfigStr, _ := json.MarshalIndent(rancherConfig, "", "  ")
	terraformConfigStr, _ := json.MarshalIndent(terraformConfig, "", "  ")
	terratestConfigStr, _ := json.MarshalIndent(terratestConfig, "", "  ")
	standaloneConfigStr, _ := json.MarshalIndent(standaloneConfig, "", "  ")

	if r.Method == "POST" && r.FormValue("action") == "confirm" {
		selection := r.FormValue("selection")
		clustertype := r.FormValue("clustertype")
		ranchertype := r.FormValue("ranchertype")
		installtype := r.FormValue("installtype")
		provider := r.FormValue("provider")
		providerversion := r.FormValue("providerversion")
		confirm := r.FormValue("confirm")

		webConfig.UpdateConfigFromForm(r.PostForm, rancherConfig)
		webConfig.UpdateConfigFromForm(r.PostForm, terraformConfig)
		webConfig.UpdateConfigFromForm(r.PostForm, terratestConfig)
		webConfig.UpdateConfigFromForm(r.PostForm, standaloneConfig)

		http.SetCookie(w, &http.Cookie{Name: "selection", Value: selection, Path: "/"})
		http.SetCookie(w, &http.Cookie{Name: "provider", Value: provider, Path: "/"})
		http.SetCookie(w, &http.Cookie{Name: "providerversion", Value: providerversion, Path: "/"})
		http.SetCookie(w, &http.Cookie{Name: "confirm", Value: confirm, Path: "/"})

		share.State.Mutex.Lock()

		if clustertype != "" {
			share.State.StageMsg = strings.Join(share.ClusterStageMessage, "\n")
		} else if ranchertype != "" {
			share.State.StageMsg = strings.Join(share.RancherStageMessage, "\n")
		}

		share.State.ErrorMsg = ""
		share.State.Mutex.Unlock()

		if clustertype != "" {
			go web.RunClusterSetupWeb(provider, providerversion, clustertype)
		} else if ranchertype != "" {
			go web.RunRancherSetupWeb(provider, providerversion, ranchertype, installtype)
		}

		http.Redirect(w, r, "/status", http.StatusSeeOther)

		return
	}

	// To allow for dynamic editing, marshal all configs to map[string]any
	var rancherMap, terraformMap, terratestMap, standaloneMap map[string]any

	rancherBytes, _ := json.Marshal(rancherConfig)
	_ = json.Unmarshal(rancherBytes, &rancherMap)

	terraformBytes, _ := json.Marshal(terraformConfig)
	_ = json.Unmarshal(terraformBytes, &terraformMap)

	terratestBytes, _ := json.Marshal(terratestConfig)
	_ = json.Unmarshal(terratestBytes, &terratestMap)

	standaloneBytes, _ := json.Marshal(standaloneConfig)
	_ = json.Unmarshal(standaloneBytes, &standaloneMap)

	data := struct {
		EditMode            bool
		RancherConfig       string
		TerraformConfig     string
		TerratestConfig     string
		StandaloneConfig    string
		RancherMap          map[string]any
		TerraformMap        map[string]any
		TerratestMap        map[string]any
		StandaloneMap       map[string]any
		MaskedRancherMap    map[string]any
		MaskedTerraformMap  map[string]any
		MaskedTerratestMap  map[string]any
		MaskedStandaloneMap map[string]any
		Selection           string
		ClusterType         string
		RancherType         string
		InstallType         string
		Provider            string
		ProviderVersion     string
	}{
		EditMode:            editMode,
		RancherConfig:       string(rancherConfigStr),
		TerraformConfig:     string(terraformConfigStr),
		TerratestConfig:     string(terratestConfigStr),
		StandaloneConfig:    string(standaloneConfigStr),
		RancherMap:          rancherMap,
		TerraformMap:        terraformMap,
		TerratestMap:        terratestMap,
		StandaloneMap:       standaloneMap,
		MaskedRancherMap:    mask.HideSensitiveFields(rancherMap),
		MaskedTerraformMap:  mask.HideSensitiveFields(terraformMap),
		MaskedTerratestMap:  mask.HideSensitiveFields(terratestMap),
		MaskedStandaloneMap: mask.HideSensitiveFields(standaloneMap),
		Selection:           selection,
		ClusterType:         clustertype,
		RancherType:         ranchertype,
		InstallType:         installtype,
		Provider:            provider,
		ProviderVersion:     providerversion,
	}

	Templates.ExecuteTemplate(w, "confirm", data)
}
