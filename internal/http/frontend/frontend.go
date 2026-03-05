package frontend

import "embed"

// Files contains the dev frontend bundle.
//
//go:embed login.html projects.html project.html css/* js/* assets/*
var Files embed.FS
