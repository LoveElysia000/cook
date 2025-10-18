package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"recipe-agent/internal/handlers"
	"recipe-agent/internal/services"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("未找到.env文件，将使用系统环境变量")
	}

	// 设置Gin模式
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin路由器
	r := gin.Default()

	// 加载HTML模板
	r.LoadHTMLGlob("templates/*")

	// 静态资源服务
	r.Static("/static", "./static")

	// 依赖注入
	recipeService := services.NewRecipeService()
	aiService := services.NewAIService()
	agentHandler := handlers.NewAgentHandler(recipeService, aiService)

	// 路由定义
	r.GET("/", handlers.IndexHandler)
	r.POST("/api/recipes", agentHandler.GetRecipes)
	r.GET("/api/health", handlers.HealthHandler)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("智能食谱助手服务启动在端口 %s", port)
	log.Printf("访问地址: http://localhost:%s", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}