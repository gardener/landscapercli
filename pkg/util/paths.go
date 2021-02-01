package util

import (
	"path/filepath"

	"github.com/gardener/landscaper/apis/core/v1alpha1"
)

const (
	BlueprintDirectoryName      = "blueprint"
	BlueprintFileName           = v1alpha1.BlueprintFileName
	ComponentDescriptorFileName = "component-descriptor.yaml"
	ResourcesFileName           = "resources.yaml"
)

func BlueprintDirectoryPath(componentPath string) string {
	return filepath.Join(componentPath, BlueprintDirectoryName)
}

func BlueprintFilePath(componentPath string) string {
	return filepath.Join(componentPath, BlueprintDirectoryName, BlueprintFileName)
}

func ComponentDescriptorFilePath(componentPath string) string {
	return filepath.Join(componentPath, ComponentDescriptorFileName)
}

func ResourcesFilePath(componentPath string) string {
	return filepath.Join(componentPath, ResourcesFileName)
}

func ExecutionFilePath(componentPath, executionName string) string {
	return filepath.Join(componentPath, BlueprintDirectoryName, ExecutionFileName(executionName))
}

func ExecutionFileName(executionName string) string {
	return "deploy-execution-" + executionName + ".yaml"
}
