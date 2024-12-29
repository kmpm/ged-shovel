package message

import (
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

type HTTPURLLoader http.Client

var DefaultSchemas = []string{
	"https://eddn.edcd.io/schemas/journal/1",
	"https://eddn.edcd.io/schemas/fsssignaldiscovered/1",
	"https://eddn.edcd.io/schemas/fssdiscoveryscan/1",
	"https://eddn.edcd.io/schemas/navroute/1",
	"https://eddn.edcd.io/schemas/scanbarycentre/1",
	"https://eddn.edcd.io/schemas/fssallbodiesfound/1",
	"https://eddn.edcd.io/schemas/dockinggranted/1",
	"https://eddn.edcd.io/schemas/commodity/3",
	"https://eddn.edcd.io/schemas/fssbodysignals/1",
	"https://eddn.edcd.io/schemas/outfitting/2",
	"https://eddn.edcd.io/schemas/shipyard/2",
	"https://eddn.edcd.io/schemas/codexentry/1",
	"https://eddn.edcd.io/schemas/approachsettlement/1",
	"https://eddn.edcd.io/schemas/dockingdenied/1",
	"https://eddn.edcd.io/schemas/navbeaconscan/1",
	"https://eddn.edcd.io/schemas/fcmaterials_capi/1",
	"https://eddn.edcd.io/schemas/fcmaterials_journal/1",
}

func (l *HTTPURLLoader) Load(url string) (any, error) {
	client := (*http.Client)(l)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%s returned status code %d", url, resp.StatusCode)
	}
	defer resp.Body.Close()

	return jsonschema.UnmarshalJSON(resp.Body)
}

func newHTTPURLLoader(insecure bool) *HTTPURLLoader {
	httpLoader := HTTPURLLoader(http.Client{
		Timeout: 15 * time.Second,
	})
	if insecure {
		httpLoader.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	return &httpLoader
}

type Validator struct {
	loader   jsonschema.SchemeURLLoader
	compiler *jsonschema.Compiler
	cache    map[string]*jsonschema.Schema
	mu       sync.Mutex
}

func NewValidator() (*Validator, error) {

	loader := jsonschema.SchemeURLLoader{
		"file":  jsonschema.FileLoader{},
		"http":  newHTTPURLLoader(false),
		"https": newHTTPURLLoader(false),
	}
	c := jsonschema.NewCompiler()
	c.UseLoader(loader)

	v := &Validator{
		loader:   loader,
		compiler: c,
		cache:    make(map[string]*jsonschema.Schema),
	}

	//preload some schemas
	for _, schemaURL := range DefaultSchemas {
		_, err := v.GetSchema(schemaURL)
		if err != nil {
			return nil, err
		}
	}

	return v, nil
}

func (v *Validator) GetSchema(schemaURL string) (*jsonschema.Schema, error) {
	if sch, ok := v.cache[schemaURL]; ok {
		return sch, nil
	}
	sch, err := v.compiler.Compile(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("could not compile schema '%s': %w", schemaURL, err)
	}
	v.cache[schemaURL] = sch
	if err := appendToFile("var/schemas.txt", schemaURL); err != nil {
		slog.Error("could not write schema to file", "error", err)
	}
	return sch, nil
}

func appendToFile(filename, data string) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(data + "\n"); err != nil {
		return err
	}
	return nil
}

func (v *Validator) Validate(schemaURL string, r io.Reader) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	sch, err := v.GetSchema(schemaURL)
	if err != nil {
		return err
	}
	inst, err := jsonschema.UnmarshalJSON(r)
	if err != nil {
		slog.Info("could not unmarshal JSON", "error", err, "schema", schemaURL, "data", r)
		return fmt.Errorf("could not unmarshal JSON: %w", err)
	}
	err = sch.Validate(inst)
	if err != nil {
		return fmt.Errorf("validation failed for '%s': %w", schemaURL, err)
	}
	return nil
}
