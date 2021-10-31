package cmd

import (
	"embed"
	"encoding/json"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
)

//go:embed schema.json
var schema embed.FS

type jsonSchemaEntry struct {
	Ref         string                      `json:"$ref,omitempty"`
	Schema      string                      `json:"$schema,omitempty"`
	Definitions map[string]*jsonSchemaEntry `json:"definitions,omitempty"`
	Type        interface{}                 `json:"type,omitempty"`
	Required    []string                    `json:"required,omitempty"`
	Properties  map[string]*jsonSchemaEntry `json:"properties,omitempty"`
	Title       string                      `json:"title,omitempty"`
	Description string                      `json:"description,omitempty"`
	OneOf       []*jsonSchemaEntry          `json:"oneOf,omitempty"`
	Enum        []string                    `json:"enum,omitempty"`
}

func (schema *jsonSchemaEntry) Resolve(file *jsonSchemaEntry) *jsonSchemaEntry {
	if strings.HasPrefix(schema.Ref, "#/definitions/") {
		return file.Definitions[strings.TrimPrefix(schema.Ref, "#/definitions/")]
	}
	return schema
}

func (schema *jsonSchemaEntry) Validate(file *jsonSchemaEntry, obj map[string]interface{}) bool {
	if len(schema.OneOf) > 0 {
		for i := 0; i < len(schema.OneOf); i++ {
			entry := schema.OneOf[i].Resolve(file)
			if entry.Validate(file, obj) {
				return true
			}
		}
	} else if schema.Type == "object" {
		for _, k := range schema.Required {
			if _, ok := obj[k]; !ok {
				return false
			}
		}
		for k := range obj {
			if _, ok := schema.Properties[k]; !ok {
				return false
			}
		}
		return true
	}
	return false
}

// getEventFromFile returns proper event name based on content of event json file
func getEventFromFile(p string) string {
	event := make(map[string]interface{})
	eventJSONBytes, err := ioutil.ReadFile(p)
	if err != nil {
		log.Errorf("%v", err)
		return ""
	}

	err = json.Unmarshal(eventJSONBytes, &event)
	if err != nil {
		log.Errorf("%v", err)
		return ""
	}

	var b []byte
	if b, err = schema.ReadFile("schema.json"); err != nil {
		return ""
	}

	jschema := &jsonSchemaEntry{}
	err = json.Unmarshal(b, jschema)
	if err != nil {
		log.Errorf("%v", err)
		return ""
	}

	for i := 0; i < len(jschema.OneOf); i++ {
		entry := jschema.OneOf[i].Resolve(jschema)
		if entry.Validate(jschema, event) {
			return strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(jschema.OneOf[i].Ref, "#/definitions/"), "$event"), "_event")
		}
	}

	return ""
}
