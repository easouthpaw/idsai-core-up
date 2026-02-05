package swagger

import "github.com/swaggo/swag"

const docTemplate = `{
  "swagger": "2.0",
  "info": { "title": "IDSAI Core API", "version": "0.1", "description": "Core platform for IDSAI projects (RBAC-driven)." },
  "basePath": "/",
  "paths": {
    "/health": {
      "get": {
        "summary": "Health check",
        "responses": { "200": { "description": "ok" }, "503": { "description": "db down" } }
      }
    }
  }
}`

var SwaggerInfo = &swag.Spec{
	Version:          "0.1",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{"http"},
	Title:            "IDSAI Core API",
	Description:      "Core platform for IDSAI projects (RBAC-driven).",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() { swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo) }
