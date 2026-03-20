package cli

import (
	"fmt"
	"net/http"

	"github.com/nazar256/shortcut-cli/internal/output"
)

type HTTPStatusError struct {
	StatusCode int
	Body       any
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("shortcut api returned status %d", e.StatusCode)
}

func RenderError(format string, err error, stderr anyWriter) {
	formatter := output.NewFormatter(format, ioDiscard{}, stderr)
	if statusErr, ok := err.(*HTTPStatusError); ok {
		if formatter.Format() == output.FormatText {
			message := fmt.Sprintf("Shortcut API error (%d)", statusErr.StatusCode)
			if payload, ok := statusErr.Body.(map[string]any); ok {
				if bodyMessage, ok := payload["message"].(string); ok && bodyMessage != "" {
					message += ": " + bodyMessage
				}
			}
			_ = formatter.PrintError(fmt.Errorf("%s", message))
			return
		}
		_ = formatter.PrintErrorValue(map[string]any{
			"error":  statusErr.Error(),
			"status": statusErr.StatusCode,
			"body":   statusErr.Body,
		})
		return
	}
	_ = formatter.PrintError(err)
}

type anyWriter interface {
	Write([]byte) (int, error)
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }

func EnsureHTTPSuccess(response *http.Response, body any) error {
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return nil
	}
	return &HTTPStatusError{StatusCode: response.StatusCode, Body: body}
}
