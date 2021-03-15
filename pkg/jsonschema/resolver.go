package jsonschema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	lsjsonschema "github.com/gardener/landscaper/pkg/landscaper/jsonschema"
	"github.com/xeipuuv/gojsonreference"
	"github.com/xeipuuv/gojsonschema"
)

// acts as a backup for preventing endless loops when resolving circular referenced jsonschemas
const defaultMaxRefDepth = 5

type JSONSchema struct {
	Ref    string                 `json:"ref"`
	Schema map[string]interface{} `json:"schema"`
}
type JSONSchemaList []JSONSchema

func (l JSONSchemaList) ToString() (string, error) {
	if len(l) == 0 {
		return "", nil
	}

	buf := bytes.Buffer{}

	_, err := buf.WriteString("JSON schema\n")
	if err != nil {
		return "", fmt.Errorf("cannot write to buffer: %w", err)
	}

	inlineSchema, err := json.MarshalIndent(l[0].Schema, "", "  ")
	if err != nil {
		return "", fmt.Errorf(`cannot marshal inline jsonschema: %w`, err)
	}

	_, err = buf.Write(inlineSchema)
	if err != nil {
		return "", fmt.Errorf("cannot write to buffer: %w", err)
	}

	if len(l) > 1 {
		_, err = buf.WriteString("\n \nReferenced JSON schemas\n")
		if err != nil {
			return "", fmt.Errorf("cannot write to buffer: %w", err)
		}

		marshaledReferencedSchemas, err := json.MarshalIndent(l[1:], "", "  ")
		if err != nil {
			return "", fmt.Errorf(`cannot marshal referenced jsonschemas: %w`, err)
		}

		_, err = buf.Write(marshaledReferencedSchemas)
		if err != nil {
			return "", fmt.Errorf("cannot write to buffer: %w", err)
		}
	}

	return buf.String(), nil
}

type JSONSchemaResolver struct {
	loaderConfig *lsjsonschema.LoaderConfig
	maxRefDepth  int
}

func NewJSONSchemaResolver(loaderConfig *lsjsonschema.LoaderConfig, maxRefDepth int) *JSONSchemaResolver {
	if maxRefDepth <= 0 {
		maxRefDepth = defaultMaxRefDepth
	}
	obj := JSONSchemaResolver{
		loaderConfig: loaderConfig,
		maxRefDepth:  maxRefDepth,
	}
	return &obj
}

// returns a list of jsonschemas. the first element in the list is the initial jsonschema itself.
// all subsequent schemas in the list are the resolved schemas that are referenced via "$ref".
func (r *JSONSchemaResolver) Resolve(schema *lsv1alpha1.JSONSchemaDefinition) (JSONSchemaList, error) {
	schemaLoader := gojsonschema.NewBytesLoader(schema.RawMessage)
	return r.resolveRecursive("root", schemaLoader, []string{})
}

func (r *JSONSchemaResolver) resolveRecursive(ref string, schemaLoader gojsonschema.JSONLoader, refHistory []string) (JSONSchemaList, error) {
	cyclicDependencyDetected := false
	for i := range refHistory {
		if refHistory[i] == ref {
			cyclicDependencyDetected = true
			break
		}
	}

	if cyclicDependencyDetected {
		return JSONSchemaList{}, nil
	}

	// backup for preventing infinite recursion
	if len(refHistory) > r.maxRefDepth {
		return nil, fmt.Errorf("maxCallDepth (%d) reached: %s", r.maxRefDepth, strings.Join(refHistory, " -> "))
	}

	refHistory = append(refHistory, ref)

	// wrap default loader if config is defined
	if r.loaderConfig != nil {
		schemaLoader = lsjsonschema.NewWrappedLoader(*r.loaderConfig, schemaLoader)
	}

	schema, err := schemaLoader.LoadJSON()
	if err != nil {
		return nil, fmt.Errorf("cannot load schema: %w", err)
	}

	schemamap, ok := schema.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert schema to map[string]interface{}")
	}

	allSchemas := JSONSchemaList{}
	allSchemas = append(allSchemas, JSONSchema{Ref: ref, Schema: schemamap})

	for key, value := range schemamap {
		if key == "$ref" {
			refStr, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("cannot convert $ref to string: $ref = %+v", value)
			}

			ref, err := gojsonreference.NewJsonReference(refStr)
			if err != nil {
				return nil, fmt.Errorf("cannot create json reference: %w", err)
			}
			if !ref.IsCanonical() {
				return nil, fmt.Errorf("ref %s must be canonical", ref.String())
			}

			newLoader := schemaLoader.LoaderFactory().New(refStr)
			subSchemas, err := r.resolveRecursive(ref.String(), newLoader, refHistory)
			if err != nil {
				return nil, fmt.Errorf("cannot resolve ref %s: %w", ref.String(), err)
			}

			allSchemas = append(allSchemas, subSchemas...)
		}
	}

	return allSchemas, nil
}
