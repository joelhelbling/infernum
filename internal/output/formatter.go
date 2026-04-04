package output

import (
	"io"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

type Formatter interface {
	FormatComparison(resp models.CompareResponse) error
	FormatRun(run models.RunResponse) error
	FormatSuites(suites []models.Suite) error
	FormatSuite(suite models.Suite) error
}

func New(format string, w io.Writer) Formatter {
	switch format {
	case "json":
		return NewJSONFormatter(w)
	default:
		return NewTableFormatter(w)
	}
}
