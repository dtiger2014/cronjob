package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	GAPI *gin.Engine
)

func InitAPIServer() {

	GAPI = gin.Default()

	// Simple group: v1
	GAPI.LoadHTMLGlob("admin/web/*.html")
	GAPI.StaticFS("/assets", http.Dir(GConfig.WebRoot))
	adminGroup := GAPI.Group("/admin")
	{
		adminGroup.GET("/index", handleAdminIndex)
	}
	return
}
