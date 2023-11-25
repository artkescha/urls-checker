package webui

import (
	"embed"
)

//go:embed static/*
var staticContent embed.FS

//go:embed templates/*
var templatesContent embed.FS
