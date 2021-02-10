package blueprints

import (
	"fmt"
)

func GetImportExpression(paramName string) string {
	return fmt.Sprintf(`{{ index .imports "%s" }}`, paramName)
}

func GetTargetNameExpression(clusterParamName string) string {
	return fmt.Sprintf(`{{ index .imports "%s" "metadata" "name" }}`, clusterParamName)
}

func GetTargetNamespaceExpression(clusterParamName string) string {
	return fmt.Sprintf(`{{ index .imports "%s" "metadata" "namespace" }}`, clusterParamName)
}
