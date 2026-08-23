// Swagger UI の確認用スタンドアロンサーバー（本体を起動せず spec だけ見たいとき用）。
//
// spec 本体は apispec（手書きのアノテーション）から `make swagger` で生成される。
// 本体サーバー（cmd/serv）も /swagger/index.html で同じ UI を配信している。
// swag init -g はこのファイルを指しているため、API 全体の説明はここに書く。
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
// @description	セッション Cookie で使う通常 API と、外部開発者向け Public API（/v1）の仕様。
// @description	通常 API はログインで発行されるセッション Cookie でのみ利用できる。
// @description	Public API は API キー（Bearer）でのみ利用でき、Cookie では利用できない。
//
// @contact.name	42 Transcendence
// @license.name	MIT
//
// @BasePath	/
// @schemes	http https
//
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				`Bearer {api_key}` 形式で指定する。キーは POST /apikeys で発行し、/v1 の Public API でのみ有効。
func main() {
	r := gin.Default()

	// Swagger UI (http://localhost:4000/swagger/index.html)
	// spec そのものは /swagger/doc.json から取れる。外部クライアントはここからコード生成できる。
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if err := r.Run(":4000"); err != nil {
		fmt.Println(err)
	}
}
