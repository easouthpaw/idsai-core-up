package frontend

import "embed"

// Files contains the dev frontend bundle.
//
//go:embed landing.html author.html login.html admin.html projects.html project.html professor.html professor-reviews.html professor-criteria.html professor-grading.html css/* js/* assets/*
var Files embed.FS
