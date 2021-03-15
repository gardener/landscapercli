// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package blueprints_test

import (
	"path/filepath"
	"testing"

	testing2 "github.com/go-logr/logr/testing"
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
	a.NoError(err)
	renderOpts := &blueprints.RenderOptions{
		ComponentDescriptorPath: "./component-descriptor.yaml",
		ValueFiles:              []string{"./imports.yaml"},
		OutDir:                  "./out",
		OutputFormat:            blueprints.YAMLOut,
	}

	a.NoError(renderOpts.Complete([]string{"./blueprint"}, testdataFs))
	fail(a.Equal("./blueprint", renderOpts.BlueprintPath))
	a.NoError(renderOpts.Run(testing2.NullLogger{}, testdataFs))

	renderedFiles, err := vfs.ReadDir(testdataFs, filepath.Join(renderOpts.OutDir, blueprints.DeployItemOutputDir))
	fail(a.NoError(err))
	a.Len(renderedFiles, 1)

	data, err := vfs.ReadFile(testdataFs, filepath.Join(renderOpts.OutDir, blueprints.DeployItemOutputDir, renderedFiles[0].Name()))
	a.NoError(err)
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
	a.NoError(err)
	renderOpts := &blueprints.RenderOptions{
		ValueFiles:   []string{"./imports.yaml"},
		OutDir:       "./out",
		OutputFormat: blueprints.YAMLOut,
	}

	a.NoError(renderOpts.Complete([]string{"./blueprint"}, testdataFs))
	fail(a.Equal("./blueprint", renderOpts.BlueprintPath))
	a.NoError(renderOpts.Run(testing2.NullLogger{}, testdataFs))

	renderedFiles, err := vfs.ReadDir(testdataFs, filepath.Join(renderOpts.OutDir, blueprints.DeployItemOutputDir))
	fail(a.NoError(err))
	a.Len(renderedFiles, 1)

	data, err := vfs.ReadFile(testdataFs, filepath.Join(renderOpts.OutDir, blueprints.DeployItemOutputDir, renderedFiles[0].Name()))
	a.NoError(err)
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
