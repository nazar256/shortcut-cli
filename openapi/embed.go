package openapi

import _ "embed"

// SpecBytes contains the vendored official Shortcut OpenAPI specification.
//
//go:embed shortcut.openapi.json
var SpecBytes []byte
