package main

import (
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/QuantumNous/new-api/router"
	"github.com/gin-gonic/gin"
)

type openAPISpec struct {
	OpenAPI        string                          `json:"openapi"`
	Info           infoObject                      `json:"info"`
	Servers        []serverObject                  `json:"servers"`
	Tags           []tagObject                     `json:"tags"`
	Paths          map[string]map[string]operation `json:"paths"`
	Components     componentsObject                `json:"components"`
	XGeneratedBy   string                          `json:"x-generated-by"`
	XGeneratedFrom []string                        `json:"x-generated-from"`
	XRouteCount    int                             `json:"x-route-count"`
}

type infoObject struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type serverObject struct {
	URL string `json:"url"`
}

type tagObject struct {
	Name string `json:"name"`
}

type operation struct {
	Tags        []string            `json:"tags,omitempty"`
	Summary     string              `json:"summary"`
	OperationID string              `json:"operationId"`
	Parameters  []parameter         `json:"parameters,omitempty"`
	RequestBody *requestBody        `json:"requestBody,omitempty"`
	Responses   map[string]response `json:"responses"`
	XGinPath    string              `json:"x-gin-path"`
	XGinHandler string              `json:"x-gin-handler"`
}

type parameter struct {
	Name        string       `json:"name"`
	In          string       `json:"in"`
	Required    bool         `json:"required"`
	Description string       `json:"description,omitempty"`
	Schema      schemaObject `json:"schema"`
}

type requestBody struct {
	Required bool                   `json:"required"`
	Content  map[string]mediaObject `json:"content"`
}

type mediaObject struct {
	Schema schemaObject `json:"schema"`
}

type response struct {
	Description string                 `json:"description"`
	Content     map[string]mediaObject `json:"content,omitempty"`
}

type componentsObject struct {
	SecuritySchemes map[string]securityScheme `json:"securitySchemes"`
	Schemas         map[string]schemaObject   `json:"schemas"`
}

type securityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
	In           string `json:"in,omitempty"`
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
}

type schemaObject struct {
	Description          string                  `json:"description,omitempty"`
	Type                 string                  `json:"type,omitempty"`
	Format               string                  `json:"format,omitempty"`
	Properties           map[string]schemaObject `json:"properties,omitempty"`
	Items                *schemaObject           `json:"items,omitempty"`
	AdditionalProperties any                     `json:"additionalProperties,omitempty"`
	Ref                  string                  `json:"$ref,omitempty"`
}

func main() {
	var baseURL string
	var outPath string

	flag.StringVar(&baseURL, "base-url", "https://privnode.com", "OpenAPI server URL")
	flag.StringVar(&outPath, "out", "", "Output JSON file path; stdout when empty")
	flag.Parse()

	spec := buildSpec(strings.TrimRight(baseURL, "/"))
	payload, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		fatal(err)
	}
	payload = append(payload, '\n')

	if outPath == "" {
		_, err = os.Stdout.Write(payload)
		if err != nil {
			fatal(err)
		}
		return
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(outPath, payload, 0o644); err != nil {
		fatal(err)
	}
}

func buildSpec(baseURL string) openAPISpec {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	router.SetApiRouter(engine)
	router.SetDashboardRouter(engine)
	router.SetRelayRouter(engine)
	router.SetVideoRouter(engine)
	router.SetSSORouter(engine)
	router.SetMigrateRouter(engine)

	routes := engine.Routes()
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})

	pathByShape := map[string]string{}
	paths := map[string]map[string]operation{}
	tagSet := map[string]struct{}{}
	operationIDs := map[string]bool{}
	routeCount := 0

	for _, route := range routes {
		if shouldExcludeRoute(route) {
			continue
		}
		routeCount++
		path, params := openAPIPath(route.Path, pathByShape)
		method := strings.ToLower(route.Method)
		tag := tagForPath(path)
		tagSet[tag] = struct{}{}

		if paths[path] == nil {
			paths[path] = map[string]operation{}
		}
		paths[path][method] = operation{
			Tags:        []string{tag},
			Summary:     fmt.Sprintf("%s %s", route.Method, path),
			OperationID: uniqueOperationID(operationID(route.Method, path), route, operationIDs),
			Parameters:  params,
			RequestBody: requestBodyForMethod(route.Method),
			Responses: map[string]response{
				"200": {
					Description: "Successful response",
					Content: map[string]mediaObject{
						"application/json": {Schema: schemaRef("AnyJSON")},
					},
				},
				"default": {
					Description: "Error response",
					Content: map[string]mediaObject{
						"application/json": {Schema: schemaRef("Error")},
					},
				},
			},
			XGinPath:    route.Path,
			XGinHandler: route.Handler,
		}
	}

	tags := make([]tagObject, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tagObject{Name: tag})
	}
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Name < tags[j].Name
	})

	return openAPISpec{
		OpenAPI: "3.1.0",
		Info: infoObject{
			Title:   "PrivNode API",
			Version: "generated",
		},
		Servers: []serverObject{{URL: baseURL}},
		Tags:    tags,
		Paths:   paths,
		Components: componentsObject{
			SecuritySchemes: map[string]securityScheme{
				"BearerAuth": {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "API key or access token",
				},
				"ApiKeyAuth": {
					Type:        "apiKey",
					In:          "header",
					Name:        "Authorization",
					Description: "Use the token format expected by the route, for example a relay API key.",
				},
			},
			Schemas: map[string]schemaObject{
				"AnyJSON": {
					Description:          "Machine-generated placeholder for route responses. The generator is route-table based and does not hand-author endpoint schemas.",
					AdditionalProperties: true,
				},
				"Error": {
					Type: "object",
					Properties: map[string]schemaObject{
						"error":   {AdditionalProperties: true},
						"message": {Type: "string"},
						"success": {Type: "boolean"},
					},
					AdditionalProperties: true,
				},
			},
		},
		XGeneratedBy: "cmd/openapi-gen from gin.Engine.Routes()",
		XGeneratedFrom: []string{
			"router.SetApiRouter",
			"router.SetDashboardRouter",
			"router.SetRelayRouter",
			"router.SetVideoRouter",
			"router.SetSSORouter",
		},
		XRouteCount: routeCount,
	}
}

func shouldExcludeRoute(route gin.RouteInfo) bool {
	path := route.Path
	if strings.HasPrefix(path, "/mj/") || strings.HasPrefix(path, "/suno/") {
		return true
	}
	if strings.Contains(path, "/mj/") && strings.HasPrefix(path, "/:mode/") {
		return true
	}
	return route.Handler == "github.com/QuantumNous/new-api/controller.RelayNotImplemented"
}

func requestBodyForMethod(method string) *requestBody {
	switch method {
	case "POST", "PUT", "PATCH":
		return &requestBody{
			Required: false,
			Content: map[string]mediaObject{
				"application/json":                  {Schema: schemaRef("AnyJSON")},
				"multipart/form-data":               {Schema: schemaRef("AnyJSON")},
				"application/x-www-form-urlencoded": {Schema: schemaRef("AnyJSON")},
			},
		}
	default:
		return nil
	}
}

func openAPIPath(ginPath string, pathByShape map[string]string) (string, []parameter) {
	segments := strings.Split(ginPath, "/")
	shapeSegments := make([]string, len(segments))
	pathSegments := make([]string, len(segments))
	paramNames := make([]string, 0)

	for i, segment := range segments {
		switch {
		case strings.HasPrefix(segment, ":"):
			name := cleanParamName(segment[1:], i)
			shapeSegments[i] = "{}"
			pathSegments[i] = "{" + name + "}"
			paramNames = append(paramNames, name)
		case strings.HasPrefix(segment, "*"):
			name := cleanParamName(segment[1:], i)
			shapeSegments[i] = "{}"
			pathSegments[i] = "{" + name + "}"
			paramNames = append(paramNames, name)
		default:
			shapeSegments[i] = segment
			pathSegments[i] = segment
		}
	}

	shape := strings.Join(shapeSegments, "/")
	if existingPath, ok := pathByShape[shape]; ok {
		return existingPath, parametersFromNames(extractPathParams(existingPath))
	}

	path := strings.Join(pathSegments, "/")
	pathByShape[shape] = path
	return path, parametersFromNames(paramNames)
}

func parametersFromNames(names []string) []parameter {
	if len(names) == 0 {
		return nil
	}
	params := make([]parameter, 0, len(names))
	for _, name := range names {
		params = append(params, parameter{
			Name:     name,
			In:       "path",
			Required: true,
			Schema:   schemaObject{Type: "string"},
		})
	}
	return params
}

var pathParamPattern = regexp.MustCompile(`\{([^}]+)\}`)

func extractPathParams(path string) []string {
	matches := pathParamPattern.FindAllStringSubmatch(path, -1)
	names := make([]string, 0, len(matches))
	for _, match := range matches {
		names = append(names, match[1])
	}
	return names
}

var invalidIdentifierChar = regexp.MustCompile(`[^A-Za-z0-9_]`)

func cleanParamName(name string, index int) string {
	name = strings.TrimSpace(name)
	name = invalidIdentifierChar.ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	if name == "" {
		name = fmt.Sprintf("param%d", index)
	}
	return name
}

func schemaRef(name string) schemaObject {
	return schemaObject{Ref: "#/components/schemas/" + name}
}

func tagForPath(path string) string {
	segments := splitPath(path)
	if len(segments) == 0 {
		return "root"
	}

	switch segments[0] {
	case "api":
		if len(segments) > 1 {
			return "api:" + strings.Trim(segments[1], "{}")
		}
		return "api"
	case "v1":
		if len(segments) > 1 && segments[1] == "dashboard" {
			return "dashboard"
		}
		return "relay:v1"
	case "v1beta":
		return "relay:v1beta"
	case "mj":
		return "midjourney"
	case "suno":
		return "suno"
	case "kling":
		return "kling"
	case "jimeng":
		return "jimeng"
	case "dashboard":
		return "dashboard"
	case "sso-beta":
		return "sso"
	case "pg":
		return "playground"
	default:
		if len(segments) > 1 && segments[1] == "mj" {
			return "midjourney"
		}
		return segments[0]
	}
}

func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

var nonOperationChar = regexp.MustCompile(`[^A-Za-z0-9]+`)

func operationID(method, path string) string {
	parts := append([]string{strings.ToLower(method)}, splitPath(path)...)
	id := nonOperationChar.ReplaceAllString(strings.Join(parts, "_"), "_")
	id = strings.Trim(id, "_")
	if id == "" {
		return strings.ToLower(method) + "_root"
	}
	return id
}

func uniqueOperationID(base string, route gin.RouteInfo, seen map[string]bool) string {
	if !seen[base] {
		seen[base] = true
		return base
	}

	sum := sha1.Sum([]byte(route.Method + " " + route.Path + " " + route.Handler))
	candidate := fmt.Sprintf("%s_%x", base, sum[:4])
	if !seen[candidate] {
		seen[candidate] = true
		return candidate
	}

	for i := 2; ; i++ {
		fallback := fmt.Sprintf("%s_%x_%d", base, sum[:4], i)
		if !seen[fallback] {
			seen[fallback] = true
			return fallback
		}
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "openapi-gen: %v\n", err)
	os.Exit(1)
}
