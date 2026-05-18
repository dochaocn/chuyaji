package apidocs

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed openapi.yaml
var openAPISpec []byte

const elementsVersion = "9.0.16"

func docsHTML() string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>初芽记 API 文档</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@stoplight/elements@%[1]s/styles.min.css"/>
  <script src="https://cdn.jsdelivr.net/npm/@stoplight/elements@%[1]s/web-components.min.js"></script>
  <style>
    html, body { margin: 0; height: 100%%; }
    elements-api { display: block; min-height: 100vh; }
  </style>
</head>
<body>
  <elements-api
    apiDescriptionUrl="/openapi.yaml"
    router="hash"
    layout="sidebar"
  ></elements-api>
</body>
</html>
`, elementsVersion)
}

func Register(r *gin.Engine) {
	r.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", openAPISpec)
	})
	r.GET("/docs", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, docsHTML())
	})
}
