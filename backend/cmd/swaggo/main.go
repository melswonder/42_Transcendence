package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// swag init が生成する spec を副作用で登録する（import しないと UI が spec を読めない）
	_ "transcendence-backend/docs/swagger"
)

// @title			Sample API
// @version		1.0
// @description	swaggoの動作確認用APIです。
// @host			localhost:4000
// @BasePath		/api/v1
func main() {
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.GET("/hello", HelloHandler)
	}

	// Swagger UIのエンドポイントを追加
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run(":4000")
}

// HelloHandler 挨拶を返すハンドラ
//
//	@Summary		挨拶を取得
//	@Description	Simple hello world endpoint
//	@Tags			example
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/hello [get]
func HelloHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "hello world"})
}
