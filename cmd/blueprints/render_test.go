// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package blueprints_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/layerfs"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	"github.com/gardener/landscapercli/cmd/blueprints"
)

func TestRenderCommandWithComponentDescriptor(t *testing.T) {
	a := assert.New(t)
	fail := failFunc(a)
	assertNestedString := assertNestedStringFunc(a)

	testdataFs, err := createTestdataFs("./testdata/00-render")
	fail(a.NoError(err))
	renderOpts := &blueprints.RenderOptions{
		ComponentDescriptorPath: "./component-descriptor.yaml",
		ValueFiles:              []string{"./imports.yaml"},
		OutDir:                  "./out",
		OutputFormat:            blueprints.YAMLOut,
	}

	a.NoError(renderOpts.Complete(logr.Discard(), []string{"./blueprint"}, testdataFs))
	fail(a.Equal("./blueprint", renderOpts.BlueprintPath))
	fail(a.NoError(renderOpts.Run(context.TODO(), logr.Discard(), testdataFs)))

	renderedFiles, err := vfs.ReadDir(testdataFs, filepath.Join(renderOpts.OutDir, blueprints.DeployItemOutputDir))
	fail(a.NoError(err))
	a.Len(renderedFiles, 1)

	data, err := vfs.ReadFile(testdataFs, filepath.Join(renderOpts.OutDir, blueprints.DeployItemOutputDir, renderedFiles[0].Name()))
	fail(a.NoError(err))
	var actual unstructured.Unstructured
	fail(a.NoError(yaml.Unmarshal(data, &actual)))
	a.Equal(actual.GetKind(), "DeployItem")
	assertNestedString(actual, "mock", "spec", "type")
	assertNestedString(actual, "my-target", "spec", "target", "name")
	assertNestedString(actual, "test", "spec", "target", "namespace")
	assertNestedString(actual, "first-value", "spec", "config", "imports", "imp1")

	assertNestedString(actual, "example-blueprint", "spec", "config", "blueprint", "ref", "resourceName")
	assertNestedString(actual, "example.com/render-cmd", "spec", "config", "componentDescriptorDef", "ref", "componentName")
	assertNestedString(actual, "0.1.0", "spec", "config", "componentDescriptorDef", "ref", "version")
}

func TestRenderCommandWithDefaults(t *testing.T) {
	a := assert.New(t)
	fail := failFunc(a)
	assertNestedString := assertNestedStringFunc(a)

	testdataFs, err := createTestdataFs("./testdata/00-render")
	fail(a.NoError(err))
	renderOpts := &blueprints.RenderOptions{
		ValueFiles:   []string{"./imports.yaml"},
		OutDir:       "./out",
		OutputFormat: blueprints.YAMLOut,
	}

	a.NoError(renderOpts.Complete(logr.Discard(), []string{"./blueprint"}, testdataFs))
	fail(a.Equal("./blueprint", renderOpts.BlueprintPath))
	fail(a.NoError(renderOpts.Run(context.TODO(), logr.Discard(), testdataFs)))

	renderedFiles, err := vfs.ReadDir(testdataFs, filepath.Join(renderOpts.OutDir, blueprints.DeployItemOutputDir))
	fail(a.NoError(err))
	a.Len(renderedFiles, 1)

	data, err := vfs.ReadFile(testdataFs, filepath.Join(renderOpts.OutDir, blueprints.DeployItemOutputDir, renderedFiles[0].Name()))
	fail(a.NoError(err))
	var actual unstructured.Unstructured
	fail(a.NoError(yaml.Unmarshal(data, &actual)))
	a.Equal(actual.GetKind(), "DeployItem")
	assertNestedString(actual, "mock", "spec", "type")
	assertNestedString(actual, "my-target", "spec", "target", "name")
	assertNestedString(actual, "test", "spec", "target", "namespace")
	assertNestedString(actual, "first-value", "spec", "config", "imports", "imp1")

	assertNestedString(actual, "example-blueprint", "spec", "config", "blueprint", "ref", "resourceName")
	assertNestedString(actual, "my-example-component", "spec", "config", "componentDescriptorDef", "ref", "componentName")
	assertNestedString(actual, "v0.0.0", "spec", "config", "componentDescriptorDef", "ref", "version")
}

func TestImportValidationOfRenderCommand(t *testing.T) {

	t.Run("should validate a import of an invalid schema", func(t *testing.T) {
		a := assert.New(t)
		fail := failFunc(a)

		testdataFs, err := createTestdataFs("./testdata/00-render")
		fail(a.NoError(err))
		renderOpts := &blueprints.RenderOptions{
			ComponentDescriptorPath: "./component-descriptor.yaml",
			ValueFiles:              []string{"./imports.yaml"},
			OutDir:                  "./out",
			OutputFormat:            blueprints.YAMLOut,
		}
		importData := `
imports:
  cluster:
    apiVersion: landscaper.gardener.cloud/v1alpha1
    kind: Target
    metadata:
      name: my-target
      namespace: test
    spec:
      type: example.com/my-type
  imp1: 10
`
		fail(a.NoError(vfs.WriteFile(testdataFs, renderOpts.ValueFiles[0], []byte(importData), os.ModePerm)))

		a.NoError(renderOpts.Complete(logr.Discard(), []string{"./blueprint"}, testdataFs))
		fail(a.Equal("./blueprint", renderOpts.BlueprintPath))
		fail(a.Error(renderOpts.Run(context.TODO(), logr.Discard(), testdataFs)))
	})

	t.Run("should validate a import of an invalid target type", func(t *testing.T) {
		a := assert.New(t)
		fail := failFunc(a)

		testdataFs, err := createTestdataFs("./testdata/00-render")
		fail(a.NoError(err))
		renderOpts := &blueprints.RenderOptions{
			ComponentDescriptorPath: "./component-descriptor.yaml",
			ValueFiles:              []string{"./imports.yaml"},
			OutDir:                  "./out",
			OutputFormat:            blueprints.YAMLOut,
		}
		importData := `
imports:
  cluster:
    apiVersion: landscaper.gardener.cloud/v1alpha1
    kind: Target
    metadata:
      name: my-target
      namespace: test
    spec:
      type: no-mock
  imp1: abc
`
		fail(a.NoError(vfs.WriteFile(testdataFs, renderOpts.ValueFiles[0], []byte(importData), os.ModePerm)))

		a.NoError(renderOpts.Complete(logr.Discard(), []string{"./blueprint"}, testdataFs))
		fail(a.Equal("./blueprint", renderOpts.BlueprintPath))
		fail(a.Error(renderOpts.Run(context.TODO(), logr.Discard(), testdataFs)))
	})

	t.Run("should validate a required import", func(t *testing.T) {
		a := assert.New(t)
		fail := failFunc(a)

		testdataFs, err := createTestdataFs("./testdata/00-render")
		fail(a.NoError(err))
		renderOpts := &blueprints.RenderOptions{
			ComponentDescriptorPath: "./component-descriptor.yaml",
			ValueFiles:              []string{"./imports.yaml"},
			OutDir:                  "./out",
			OutputFormat:            blueprints.YAMLOut,
		}
		importData := `
imports:
  cluster:
    apiVersion: landscaper.gardener.cloud/v1alpha1
    kind: Target
    metadata:
      name: my-target
      namespace: test
    spec:
      type: example.com/my-type
`
		fail(a.NoError(vfs.WriteFile(testdataFs, renderOpts.ValueFiles[0], []byte(importData), os.ModePerm)))

		a.NoError(renderOpts.Complete(logr.Discard(), []string{"./blueprint"}, testdataFs))
		fail(a.Equal("./blueprint", renderOpts.BlueprintPath))
		fail(a.Error(renderOpts.Run(context.TODO(), logr.Discard(), testdataFs)))
	})

}

func assertNestedStringFunc(a *assert.Assertions) func(obj unstructured.Unstructured, expected string, fields ...string) {
	return func(obj unstructured.Unstructured, expected string, fields ...string) {
		fail := failFunc(a)
		actual, _, err := unstructured.NestedString(obj.Object, fields...)
		fail(a.NoError(err))
		fail(a.Equal(expected, actual))
	}
}

func failFunc(a *assert.Assertions) func(failed bool) {
	return func(assertionSucceeded bool) {
		if assertionSucceeded {
			return
		}
		a.FailNow("")
	}
}

func createTestdataFs(path string) (vfs.FileSystem, error) {
	testdataFs, err := projectionfs.New(osfs.New(), path)
	if err != nil {
		return nil, err
	}
	return layerfs.New(memoryfs.New(), testdataFs), nil
}
