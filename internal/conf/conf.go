package conf

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"

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

type Global struct {
	Protocol string `json:"protocol" validate:"required,oneof='openai' 'anthropic'"`
	BaseURL  string `json:"baseUrl" validate:"required"`
	Model    string `json:"model" validate:"required"`
	APIKey   string `json:"apiKey" validate:"required"`
}

func Load() (*Global, error) {
	var conf Global
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
	return &conf, nil
}

func tryRead(path string) ([]byte, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return b, true
}

func expandFields(v *Global) {
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
		}
	}
}
