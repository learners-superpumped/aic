// Package output renders command results as table, JSON, or YAML.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"golang.org/x/text/width"
	"gopkg.in/yaml.v3"
)

// RowFunc maps one element of a slice to its table columns.
type RowFunc func(v any) []string

// Renderer writes values in a chosen format.
type Renderer struct {
	format string
	w      io.Writer
}

// Format reports the renderer's output format ("table", "json", or "yaml").
// Commands that print a human-readable summary in table mode but a full
// structured object in json/yaml mode branch on this.
func (r *Renderer) Format() string { return r.format }

// Writer returns the renderer's output writer, so table-mode summary lines go
// to the same destination as rendered output.
func (r *Renderer) Writer() io.Writer { return r.w }

// New returns a Renderer for "table", "json", or "yaml".
func New(format string, w io.Writer) (*Renderer, error) {
	switch format {
	case "table", "json", "yaml":
		return &Renderer{format: format, w: w}, nil
	default:
		return nil, fmt.Errorf("invalid output format %q (want table|json|yaml)", format)
	}
}

// Print renders v. For "table", headers and rowFn describe the columns; if v is
// a slice each element becomes a row, otherwise v is rendered as a single row
// via rowFn. For "json"/"yaml", headers and rowFn are ignored.
func (r *Renderer) Print(v any, headers []string, rowFn RowFunc) error {
	switch r.format {
	case "json":
		enc := json.NewEncoder(r.w)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	case "yaml":
		// Encode via JSON so YAML honors the same `json` tags and struct
		// embedding as JSON output — one source of truth, no per-DTO yaml tags.
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		var generic any
		if err := json.Unmarshal(b, &generic); err != nil {
			return err
		}
		enc := yaml.NewEncoder(r.w)
		if err := enc.Encode(generic); err != nil {
			_ = enc.Close()
			return err
		}
		return enc.Close()
	default:
		return r.printTable(v, headers, rowFn)
	}
}

// tableGap is the number of spaces between columns.
const tableGap = 2

func (r *Renderer) printTable(v any, headers []string, rowFn RowFunc) error {
	var rows [][]string
	if len(headers) > 0 {
		rows = append(rows, headers)
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ {
			rows = append(rows, rowFn(rv.Index(i).Interface()))
		}
	} else if rowFn != nil {
		rows = append(rows, rowFn(v))
	}

	// Column widths measured by on-screen cell width, so CJK (double-width) text
	// doesn't push later columns out of alignment the way text/tabwriter does.
	var widths []int
	for _, row := range rows {
		for i, cell := range row {
			if w := displayWidth(cell); i >= len(widths) {
				widths = append(widths, w)
			} else if w > widths[i] {
				widths[i] = w
			}
		}
	}

	var b strings.Builder
	for _, row := range rows {
		for i, cell := range row {
			b.WriteString(cell)
			if i < len(row)-1 { // pad every column but the last (no trailing space)
				b.WriteString(strings.Repeat(" ", widths[i]-displayWidth(cell)+tableGap))
			}
		}
		b.WriteByte('\n')
	}
	_, err := io.WriteString(r.w, b.String())
	return err
}

// displayWidth returns the number of terminal cells s occupies, counting East
// Asian Wide and Fullwidth runes as 2.
func displayWidth(s string) int {
	w := 0
	for _, run := range s {
		switch width.LookupRune(run).Kind() {
		case width.EastAsianWide, width.EastAsianFullwidth:
			w += 2
		default:
			w++
		}
	}
	return w
}
