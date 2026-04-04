package output

import (
	"encoding/json"
	"io"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

type JSONFormatter struct {
	w io.Writer
}

func NewJSONFormatter(w io.Writer) *JSONFormatter {
	return &JSONFormatter{w: w}
}

func (f *JSONFormatter) FormatComparison(resp models.CompareResponse) error {
	return f.encode(resp)
}

func (f *JSONFormatter) FormatRun(run models.RunResponse) error {
	return f.encode(run)
}

func (f *JSONFormatter) FormatSuites(suites []models.Suite) error {
	return f.encode(suites)
}

func (f *JSONFormatter) FormatSuite(suite models.Suite) error {
	return f.encode(suite)
}

func (f *JSONFormatter) encode(v any) error {
	enc := json.NewEncoder(f.w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
