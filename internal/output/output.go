// Package output renders command results as table, JSON, or YAML.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"text/tabwriter"

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
		enc := yaml.NewEncoder(r.w)
		if err := enc.Encode(v); err != nil {
			_ = enc.Close()
			return err
		}
		return enc.Close()
	default:
		return r.printTable(v, headers, rowFn)
	}
}

func (r *Renderer) printTable(v any, headers []string, rowFn RowFunc) error {
	tw := tabwriter.NewWriter(r.w, 0, 2, 2, ' ', 0)
	if len(headers) > 0 {
		fmt.Fprintln(tw, join(headers))
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ {
			fmt.Fprintln(tw, join(rowFn(rv.Index(i).Interface())))
		}
	} else if rowFn != nil {
		fmt.Fprintln(tw, join(rowFn(v)))
	}
	return tw.Flush()
}

func join(cols []string) string {
	out := ""
	for i, c := range cols {
		if i > 0 {
			out += "\t"
		}
		out += c
	}
	return out
}
