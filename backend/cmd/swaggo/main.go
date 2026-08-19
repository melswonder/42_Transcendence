// Swagger UI の確認用サーバー。
//
// spec 本体は apispec（手書きのアノテーション）から `make swagger` で生成される。
// ここは生成済み spec を表示するだけで、API の実装は持たない。
// 実装は cmd/serv 側に入る予定なので、そちらが出来たらこのコマンドは役目を終える。
package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// swag init が生成する spec を副作用で登録する（import しないと UI が spec を読めない）
	_ "transcendence-backend/docs/swagger"
)

// @title			ft_transcendence API
// @version		1.0
// @description	ユーザー・認証・フレンド・ブロック・アバターを扱う公開 API の仕様。
// @description	外部クライアントから叩けるよう、認証は Cookie ではなく Bearer トークンで行う。
// @description	現時点では仕様のみで、エンドポイントの実装はこれから追加される。
//
// @contact.name	42 Transcendence
// @license.name	MIT
//
// @host		localhost:4000
// @BasePath	/api/v1
// @schemes	http https
//
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				`Bearer {access_token}` 形式で指定する。トークンは POST /auth/login などで発行する。
func main() {
	r := gin.Default()

	// Swagger UI (http://localhost:4000/swagger/index.html)
	// spec そのものは /swagger/doc.json から取れる。外部クライアントはここからコード生成できる。
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if err := r.Run(":4000"); err != nil {
		fmt.Println(err)
	}
}
