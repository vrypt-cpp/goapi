package core

import (
	"encoding/json"
	"net/http"
)

type Route struct {
	Method      string
	Path        string
	Summary     string
	Description string
	Tags        []string
	Request     any
	Response    any
	Responses   map[int]any
	Params      []Parameter
	Handler     http.HandlerFunc
}

type Plugin interface {
	Name() string
	Routes() []Route
}

type App struct {
	mux      *http.ServeMux
	registry *SchemaRegistry
	spec     *OpenAPI
	prefix   string
	cors     *CORSConfig
}

func NewApp(title, version, prefix string) *App {
	return &App{
		mux:      http.NewServeMux(),
		registry: NewSchemaRegistry(),
		prefix:   prefix,
		spec: &OpenAPI{
			OpenAPI: "3.0.3",
			Info:    Info{Title: title, Version: version},
			Paths:   map[string]PathItem{},
		},
	}
}

func (a *App) Use(plugins ...Plugin) {
	for _, p := range plugins {
		for _, route := range p.Routes() {
			a.register(p.Name(), route)
		}
	}
}

func (a *App) register(pluginName string, r Route) {
	fullPath := a.prefix + r.Path
	a.mux.HandleFunc(r.Method+" "+fullPath, r.Handler)

	op := Operation{
		Summary:     r.Summary,
		Description: r.Description,
		Tags:        r.Tags,
		OperationID: pluginName + "." + r.Method + r.Path,
		Parameters:  r.Params,
		Responses:   map[string]Response{},
	}

	if r.Request != nil {
		schema := a.registry.Register(r.Request)
		op.RequestBody = &RequestBody{
			Required: true,
			Content: map[string]MediaType{
				"application/json": {Schema: schema},
			},
		}
	}

	if r.Response != nil {
		schema := a.registry.Register(r.Response)
		op.Responses["200"] = Response{
			Description: "OK",
			Content: map[string]MediaType{
				"application/json": {Schema: schema},
			},
		}
	}

	for code, body := range r.Responses {
		resp := Response{Description: http.StatusText(code)}
		if body != nil {
			schema := a.registry.Register(body)
			resp.Content = map[string]MediaType{
				"application/json": {Schema: schema},
			}
		}
		op.Responses[itoa(code)] = resp
	}

	if len(op.Responses) == 0 {
		op.Responses["200"] = Response{Description: "OK"}
	}

	openapiPath := toOpenAPIPath(fullPath)
	if a.spec.Paths[openapiPath] == nil {
		a.spec.Paths[openapiPath] = PathItem{}
	}
	a.spec.Paths[openapiPath][toLower(r.Method)] = op
}

func (a *App) Spec() *OpenAPI {
	a.spec.Components.Schemas = a.registry.Schemas()
	return a.spec
}

func (a *App) ServeOpenAPI(path string) {
	a.mux.HandleFunc("GET "+path, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(a.Spec())
	})
}

func (a *App) EnableCORS(cfg CORSConfig) {
	a.cors = &cfg
}

func (a *App) Handler() http.Handler {
	if a.cors != nil {
		return corsMiddleware(*a.cors, a.mux)
	}
	return a.mux
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [12]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

func toLower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

func toOpenAPIPath(path string) string {
	b := []byte(path)
	out := make([]byte, 0, len(b))
	i := 0
	for i < len(b) {
		if b[i] == '{' {
			j := i
			for j < len(b) && b[j] != '}' {
				j++
			}
			name := b[i+1 : j]
			if n := len(name); n >= 3 && string(name[n-3:]) == "..." {
				name = name[:n-3]
			}
			out = append(out, '{')
			out = append(out, name...)
			out = append(out, '}')
			i = j + 1
			continue
		}
		out = append(out, b[i])
		i++
	}
	return string(out)
}
