// api/internal/template/embed.go
package template

import "embed"

//go:embed all:html/*
var Files embed.FS

//go:embed static/*
var StaticFiles embed.FS
