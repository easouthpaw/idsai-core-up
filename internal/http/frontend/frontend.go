package frontend

import "embed"

// Files contains the dev frontend bundle.
//
//go:embed 404.html landing.html author.html login.html admin.html projects.html project.html invites.html professor.html professor-reviews.html professor-criteria.html professor-grading.html settings.html profile.html groups.html css/* js/* assets/*
var Files embed.FS
