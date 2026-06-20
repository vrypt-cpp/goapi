package core

import (
	"reflect"
	"strings"
)

type SchemaRegistry struct {
	schemas map[string]Schema
}

func NewSchemaRegistry() *SchemaRegistry {
	return &SchemaRegistry{schemas: map[string]Schema{}}
}

func (r *SchemaRegistry) Schemas() map[string]Schema {
	return r.schemas
}

func (r *SchemaRegistry) Register(v any) Schema {
	if v == nil {
		return Schema{}
	}
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return r.registerType(t)
}

func (r *SchemaRegistry) registerType(t reflect.Type) Schema {
	switch t.Kind() {
	case reflect.String:
		return Schema{Type: "string"}
	case reflect.Bool:
		return Schema{Type: "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return Schema{Type: "number"}
	case reflect.Slice, reflect.Array:
		item := r.registerType(elemType(t.Elem()))
		return Schema{Type: "array", Items: &item}
	case reflect.Map:
		val := r.registerType(elemType(t.Elem()))
		return Schema{Type: "object", AdditionalProperties: &val}
	case reflect.Ptr:
		return r.registerType(t.Elem())
	case reflect.Struct:
		name := t.Name()
		if name == "" {
			return r.buildObjectSchema(t)
		}
		if _, ok := r.schemas[name]; !ok {
			r.schemas[name] = Schema{}
			r.schemas[name] = r.buildObjectSchema(t)
		}
		return Schema{Ref: "#/components/schemas/" + name}
	default:
		return Schema{Type: "object"}
	}
}

func elemType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func (r *SchemaRegistry) buildObjectSchema(t reflect.Type) Schema {
	props := map[string]Schema{}
	var required []string

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		jsonTag := f.Tag.Get("json")
		name := f.Name
		omitempty := false
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] == "-" {
				continue
			}
			if parts[0] != "" {
				name = parts[0]
			}
			for _, p := range parts[1:] {
				if p == "omitempty" {
					omitempty = true
				}
			}
		}

		if f.Anonymous && jsonTag == "" {
			embedded := r.buildObjectSchema(elemType(f.Type))
			for k, v := range embedded.Properties {
				props[k] = v
			}
			required = append(required, embedded.Required...)
			continue
		}

		fieldSchema := r.registerType(f.Type)
		if desc := f.Tag.Get("doc"); desc != "" {
			fieldSchema.Description = desc
		}
		if enumTag := f.Tag.Get("enum"); enumTag != "" {
			fieldSchema.Enum = strings.Split(enumTag, ",")
		}
		props[name] = fieldSchema

		if !omitempty && f.Tag.Get("required") != "false" {
			required = append(required, name)
		}
	}

	return Schema{Type: "object", Properties: props, Required: required}
}
