package shortcut

import (
	"context"
	"fmt"
	"os"

	"github.com/nazar256/shortcut-cli/internal/config"
	shortcutv3 "github.com/nazar256/shortcut-cli/internal/gen/shortcutv3"
	"github.com/nazar256/shortcut-cli/internal/httpx"
	"github.com/nazar256/shortcut-cli/internal/output"
)

type Runtime struct {
	Config    *config.Config
	HTTP      *httpx.Client
	Client    *shortcutv3.ClientWithResponses
	Formatter *output.Formatter
}

func NewRuntime(ctx context.Context, outputFormat string, loadOptions config.LoadOptions) (*Runtime, error) {
	_ = ctx

	cfg, err := config.Load(loadOptions)
	if err != nil {
		return nil, err
	}

	httpClient := httpx.NewClient(cfg)
	client, err := shortcutv3.NewClientWithResponses(
		cfg.BaseURL,
		shortcutv3.WithHTTPClient(httpClient.HTTPClient()),
		shortcutv3.WithRequestEditorFn(httpClient.RequestEditorFn()),
	)
	if err != nil {
		return nil, fmt.Errorf("create shortcut client: %w", err)
	}

	return &Runtime{
		Config:    cfg,
		HTTP:      httpClient,
		Client:    client,
		Formatter: output.NewFormatter(outputFormat, os.Stdout, os.Stderr),
	}, nil

}

func (r *Runtime) GetHTTP() *httpx.Client {
	return r.HTTP
}
