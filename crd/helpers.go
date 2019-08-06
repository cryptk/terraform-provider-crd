package crd

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func idParts(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		err := fmt.Errorf("Unexpected ID format (%q), expected %q.", id, "namespace/name")
		return "", "", err
	}

	return parts[0], parts[1], nil
}

func buildId(un *unstructured.Unstructured) string {
	return un.GetNamespace() + "/" + un.GetName()
}

// helper function to convert YAML default type to JSON default type
func cleanYAML(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = cleanYAML(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = cleanYAML(v)
		}
	}
	return i
}
