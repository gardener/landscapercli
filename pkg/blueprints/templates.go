package blueprints

import (
	"fmt"
	"strings"
)

func GetImportExpression(paramName string) string {
	if strings.Contains(paramName, "-") {
		return fmt.Sprintf(`{{ index .imports "%s" }}`, paramName)
	}

	return fmt.Sprintf(`{{ .imports.%s }}`, paramName)
}

func GetTargetNameExpression(clusterParamName string) string {
	if strings.Contains(clusterParamName, "-") {
		return fmt.Sprintf(`{{ index .imports "%s" "metadata" "name" }}`, clusterParamName)
	}

	return fmt.Sprintf(`{{ .imports.%s.metadata.name }}`, clusterParamName)
}

func GetTargetNamespaceExpression(clusterParamName string) string {
	if strings.Contains(clusterParamName, "-") {
		return fmt.Sprintf(`{{ index .imports "%s" "metadata" "namespace" }}`, clusterParamName)
	}

	return fmt.Sprintf(`{{ .imports.%s.metadata.namespace }}`, clusterParamName)
}
