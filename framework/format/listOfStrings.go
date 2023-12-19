package framework

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// ListOfStrings is a function that will format a list of strings into a HCL list.
func ListOfStrings(list []string) hclwrite.Tokens {
	stringOfValues := ``

	for _, value := range list {
		if value != list[len(list)-1] {
			stringOfValues += `"` + value + `"` + `, `
		}
		if value == list[len(list)-1] {
			stringOfValues += `"` + value + `"`
		}
	}

	formattedList := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(`[` + stringOfValues + `]`)},
	}

	return formattedList
}
