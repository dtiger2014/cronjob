package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleAdminIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}
