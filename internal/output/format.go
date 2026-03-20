package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

type Formatter struct {
	format Format
	out    io.Writer
	errOut io.Writer
}

func NewFormatter(format string, out io.Writer, errOut io.Writer) *Formatter {
	resolved := Format(strings.ToLower(strings.TrimSpace(format)))
	if resolved != FormatJSON {
		resolved = FormatText
	}

	return &Formatter{
		format: resolved,
		out:    out,
		errOut: errOut,
	}
}

func (f *Formatter) Format() Format {
	return f.format
}

func (f *Formatter) Print(value any) error {
	if f.format == FormatJSON {
		return writeJSON(f.out, value)
	}

	_, err := fmt.Fprintln(f.out, ToText(value))
	return err
}

func (f *Formatter) PrintError(err error) error {
	if f.format == FormatJSON {
		return writeJSON(f.errOut, map[string]any{"error": err.Error()})
	}

	_, writeErr := fmt.Fprintf(f.errOut, "Error: %v\n", err)
	return writeErr
}

func (f *Formatter) PrintErrorValue(value any) error {
	if f.format == FormatJSON {
		return writeJSON(f.errOut, value)
	}
	_, err := fmt.Fprintf(f.errOut, "Error: %s\n", ToText(value))
	return err
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func ToText(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		formatted, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return fmt.Sprintf("%v", value)
		}
		return string(formatted)
	}
}
