package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	dotenv "github.com/joho/godotenv"
	"github.com/open-portfolios/codefolio/pkg/stdx"
)

var (
	relativePath = filepath.Join(".codefolio", "config.json")

	loadPaths []string
)

func init() {
	// Load dotenv
	if err := dotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}

	// Conf besides executable
	//
	// Loaded first, covered by later configurations if any.
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	loadPaths = append(loadPaths, filepath.Join(filepath.Dir(exe), relativePath))

	// Conf besides cwd
	//
	// Loaded and cover the existing configurations if any.
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	loadPaths = append(loadPaths, filepath.Join(cwd, relativePath))
}

type Struct struct {
	Protocol   string      `json:"protocol" validate:"required,oneof='openai' 'anthropic'"`
	BaseURL    string      `json:"baseUrl" validate:"required"`
	Model      string      `json:"model" validate:"required"`
	APIKey     string      `json:"apiKey" validate:"required"`
	MCPServers []MCPServer `json:"mcpServers"`
}

type MCPServer struct {
	Name      string            `json:"name"`
	Transport string            `json:"transport"`
	Command   string            `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	URL       string            `json:"url,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	TimeoutMS int               `json:"timeoutMs,omitempty"`
	Enabled   *bool             `json:"enabled,omitempty"`
}

func Load() (*Struct, error) {
	var conf Struct
	for _, p := range loadPaths {
		if content, ok := tryRead(p); ok {
			if err := json.Unmarshal(content, &conf); err != nil {
				return nil, err
			}
		}
	}
	expandFields(&conf)

	v := validator.New(validator.WithRequiredStructEnabled())
	if err := v.Struct(&conf); err != nil {
		return nil, err
	}
	if err := validateMCPServers(conf.MCPServers); err != nil {
		return nil, err
	}
	return &conf, nil
}

func (c MCPServer) IsEnabled() bool {
	return c.Enabled == nil || *c.Enabled
}

func validateMCPServers(servers []MCPServer) error {
	names := make(map[string]struct{}, len(servers))
	for _, server := range servers {
		name := strings.TrimSpace(server.Name)
		if name == "" {
			return errors.New("mcp server name is required")
		}
		if _, ok := names[name]; ok {
			return fmt.Errorf("duplicate mcp server name %q", name)
		}
		names[name] = struct{}{}

		switch server.Transport {
		case "stdio":
			if strings.TrimSpace(server.Command) == "" || server.URL != "" {
				return fmt.Errorf("mcp server %q with stdio transport requires command and no url", name)
			}
		case "streamable", "sse":
			if server.Command != "" || server.URL == "" {
				return fmt.Errorf("mcp server %q with %s transport requires url and no command", name, server.Transport)
			}
			parsed, err := url.ParseRequestURI(server.URL)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
				return fmt.Errorf("mcp server %q has invalid url", name)
			}
		default:
			return fmt.Errorf("mcp server %q has unsupported transport %q", name, server.Transport)
		}
		if server.TimeoutMS < 0 {
			return fmt.Errorf("mcp server %q timeoutMs must be positive", name)
		}
	}
	return nil
}

func tryRead(path string) ([]byte, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return b, true
}

func expandFields(v *Struct) {
	q := stdx.NewSliceQueue[reflect.Value]()
	q.PushBack(reflect.ValueOf(v))

	for q.Len() > 0 {
		elem := q.PopFront()

		switch elem.Kind() {
		case reflect.Pointer:
			q.PushBack(elem.Elem())
		case reflect.String:
			elem.SetString(os.ExpandEnv(elem.String()))
		case reflect.Struct:
			for ty, field := range elem.Fields() {
				// unexported field that can only be accessed in certain pkg.
				if ty.PkgPath != "" {
					continue
				}
				q.PushBack(field)
			}
		case reflect.Array, reflect.Slice:
			for i := range elem.Len() {
				q.PushBack(elem.Index(i))
			}
		case reflect.Map:
			iter := elem.MapRange()
			for iter.Next() {
				value := iter.Value()
				if value.Kind() == reflect.String {
					expanded := reflect.ValueOf(os.ExpandEnv(value.String()))
					elem.SetMapIndex(iter.Key(), expanded)
				}
			}
		}
	}
}
