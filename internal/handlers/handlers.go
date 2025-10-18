package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// IndexHandler 处理首页请求
func IndexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "智能食谱助手",
		"version": "2.1.0",
	})
}

// HealthHandler 处理健康检查请求
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"service": "recipe-agent",
		"version": "2.1.0",
		"environment": os.Getenv("GIN_MODE"),
	})
}