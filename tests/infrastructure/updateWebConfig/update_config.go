package updateWebConfig

import "encoding/json"

// UpdateConfigFromForm updates a config struct from form values
func UpdateConfigFromForm(postForm map[string][]string, cfg any) {
	if cfg == nil {
		return
	}

	formMap := make(map[string]any)
	for key, values := range postForm {
		if len(values) > 0 {
			formMap[key] = values[0]
		}
	}

	formJSON, _ := json.Marshal(formMap)
	_ = json.Unmarshal(formJSON, cfg)
}
