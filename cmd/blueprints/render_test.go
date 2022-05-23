// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package blueprints_test

import (
	"context"
	"io/fs"
	"os"
	"path"
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

	testdataFs, err := createTestdataFs("/")
	fail(a.NoError(err))
	renderOpts := &blueprints.RenderOptions{
		ComponentDescriptorPath: "./testdata/00-render/component-descriptor.yaml",
		ValueFiles:              []string{"./testdata/00-render/imports.yaml"},
		OutDir:                  "./out",
		OutputFormat:            blueprints.YAMLOut,
	}

	a.NoError(renderOpts.Complete(logr.Discard(), []string{"./testdata/00-render/blueprint"}, testdataFs))

	absBlueprintPath, _ := filepath.Abs("./testdata/00-render/blueprint")
	fail(a.Equal(absBlueprintPath, renderOpts.BlueprintPath))
	fail(a.NoError(renderOpts.Run(context.TODO(), logr.Discard(), testdataFs)))

	renderedFiles, err := vfs.ReadDir(testdataFs, filepath.Join(renderOpts.OutDir, blueprints.DeployItemOutputDir))
	fail(a.NoError(err))
	a.Len(renderedFiles, 2, "expect a deploy item and the state file")

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

	testdataFs, err := createTestdataFs("/")
	fail(a.NoError(err))
	renderOpts := &blueprints.RenderOptions{
		ValueFiles:   []string{"./testdata/00-render/imports.yaml"},
		OutDir:       "./out",
		OutputFormat: blueprints.YAMLOut,
	}

	a.NoError(renderOpts.Complete(logr.Discard(), []string{"./testdata/00-render/blueprint"}, testdataFs))

	absBlueprintFilePath, _ := filepath.Abs("./testdata/00-render/blueprint")
	fail(a.Equal(absBlueprintFilePath, renderOpts.BlueprintPath))
	fail(a.NoError(renderOpts.Run(context.TODO(), logr.Discard(), testdataFs)))

	renderedFiles, err := vfs.ReadDir(testdataFs, filepath.Join(renderOpts.OutDir, blueprints.DeployItemOutputDir))
	fail(a.NoError(err))
	a.Len(renderedFiles, 2, "expect a deploy item and the state file")

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

		testdataFs, err := createTestdataFs("/")
		fail(a.NoError(err))

		absValuePath, _ := filepath.Abs("./testdata/00-render/imports.yaml")
		renderOpts := &blueprints.RenderOptions{
			ComponentDescriptorPath: "./testdata/00-render/component-descriptor.yaml",
			ValueFiles:              []string{absValuePath},
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

		a.NoError(renderOpts.Complete(logr.Discard(), []string{"./testdata/00-render/blueprint"}, testdataFs))
		absBlueprintPath, _ := filepath.Abs("./testdata/00-render/blueprint")
		fail(a.Equal(absBlueprintPath, renderOpts.BlueprintPath))
		fail(a.Error(renderOpts.Run(context.TODO(), logr.Discard(), testdataFs)))
	})

	t.Run("should validate a import of an invalid target type", func(t *testing.T) {
		a := assert.New(t)
		fail := failFunc(a)

		testdataFs, err := createTestdataFs("/")
		fail(a.NoError(err))
		absValuePath, _ := filepath.Abs("./testdata/00-render/imports.yaml")
		renderOpts := &blueprints.RenderOptions{
			ComponentDescriptorPath: "./testdata/00-render/component-descriptor.yaml",
			ValueFiles:              []string{absValuePath},
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

		a.NoError(renderOpts.Complete(logr.Discard(), []string{"./testdata/00-render/blueprint"}, testdataFs))
		absBlueprintPath, _ := filepath.Abs("./testdata/00-render/blueprint")
		fail(a.Equal(absBlueprintPath, renderOpts.BlueprintPath))
		fail(a.Error(renderOpts.Run(context.TODO(), logr.Discard(), testdataFs)))
	})

	t.Run("should validate a required import", func(t *testing.T) {
		a := assert.New(t)
		fail := failFunc(a)

		testdataFs, err := createTestdataFs("/")
		fail(a.NoError(err))
		renderOpts := &blueprints.RenderOptions{
			ComponentDescriptorPath: "./testdata/00-render/component-descriptor.yaml",
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

		a.NoError(renderOpts.Complete(logr.Discard(), []string{"./testdata/00-render/blueprint"}, testdataFs))
		absBlueprintPath, _ := filepath.Abs("./testdata/00-render/blueprint")
		fail(a.Equal(absBlueprintPath, renderOpts.BlueprintPath))
		fail(a.Error(renderOpts.Run(context.TODO(), logr.Discard(), testdataFs)))
	})

}

func TestRenderCommandWithExportTemplates(t *testing.T) {
	a := assert.New(t)
	fail := failFunc(a)

	testDir := "../../docs/examples/render-blueprint/02-render-with-export-templates"
	blueprintPath := path.Join(testDir, "blueprint")
	componentDescriptorPath := path.Join(testDir, "component-descriptor")

	testdataFs, err := createTestdataFs("/")
	fail(a.NoError(err))

	renderOpts := &blueprints.RenderOptions{
		ComponentDescriptorPath: path.Join(componentDescriptorPath, "root/component-descriptor.yaml"),
		AdditionalComponentDescriptorPath: []string{
			path.Join(componentDescriptorPath, "comp-a/component-descriptor.yaml"),
			path.Join(componentDescriptorPath, "comp-b/component-descriptor.yaml"),
			path.Join(componentDescriptorPath, "comp-c/component-descriptor.yaml"),
		},
		ResourcesPath:       path.Join(testDir, "resources.yaml"),
		ExportTemplatesPath: path.Join(testDir, "export-templates.yaml"),
		ValueFiles: []string{
			path.Join(testDir, "values.yaml"),
		},
		OutputFormat: blueprints.YAMLOut,
		OutDir:       "./out",
	}

	a.NoError(renderOpts.Complete(logr.Discard(), []string{path.Join(blueprintPath, "root")}, testdataFs))
	ctx := context.Background()
	defer ctx.Done()
	fail(a.NoError(renderOpts.Run(ctx, logr.Discard(), testdataFs)))

	files := make(map[string]interface{})

	err = vfs.Walk(testdataFs, renderOpts.OutDir, func(path string, info fs.FileInfo, err error) error {
		fail(a.NoError(err))

		if info.IsDir() {
			return nil
		}

		data, err2 := vfs.ReadFile(testdataFs, path)
		fail(a.NoError(err2))

		var actual map[string]interface{}
		fail(a.NoError(yaml.Unmarshal(data, &actual)))

		files[path] = actual
		return nil
	})
	fail(a.NoError(err))

	fail(a.Contains(files, "out/root/installation"))
	fail(a.Contains(files, "out/root/imports"))
	fail(a.Contains(files, "out/root/exports"))

	fail(a.Contains(files, "out/root/subinst-a/installation"))
	fail(a.Contains(files, "out/root/subinst-a/imports"))
	fail(a.Contains(files, "out/root/subinst-a/exports"))
	fail(a.Contains(files, "out/root/subinst-a/deployitems/subinst-a-deploy"))

	fail(a.Contains(files, "out/root/subinst-b/installation"))
	fail(a.Contains(files, "out/root/subinst-b/imports"))
	fail(a.Contains(files, "out/root/subinst-b/exports"))
	fail(a.Contains(files, "out/root/subinst-b/deployitems/subinst-b-deploy"))

	fail(a.Contains(files, "out/root/subinst-c/installation"))
	fail(a.Contains(files, "out/root/subinst-c/imports"))
	fail(a.Contains(files, "out/root/subinst-c/exports"))
	fail(a.NotContains(files, "out/root/subinst-a/deployitems/subinst-c-deploy"))

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
