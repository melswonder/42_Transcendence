// Package apispec は公開 API の仕様（OpenAPI / Swagger）だけを定義する。
//
// ここにあるのは「外に見せる型」と swag のアノテーションだけで、実装は一切持たない。
// docs/swagger 配下の spec はこのパッケージと cmd/swaggo/main.go から生成される。
//
// なぜ handler と分けているか:
//   - 実装より先に API の形（パス・リクエスト・レスポンス・エラー）を確定させ、
//     フロントや外部クライアントが spec を見て並行で作業できるようにするため
//   - infrastructure のモデルをそのまま外に出さないため。
//     password_hash / token_hash / storage_key は公開 API に載せてはいけない
//
// 実装を始めるときは、対応する型を handler 層へ移し、アノテーションを実際の
// ハンドラー関数のコメントへ移すとよい（その場合はこのパッケージを削除する）。
package apispec

// ErrorResponse は全エンドポイント共通の失敗時の形。
// code は機械可読な識別子、message は人間向けの説明。
type ErrorResponse struct {
	Code    string `json:"code"    example:"unauthorized"`
	Message string `json:"message" example:"認証が必要です"`
}

// Pagination は一覧系エンドポイントが共通で返すページング情報。
type Pagination struct {
	Total  int `json:"total"  example:"128"` // 条件に合致する総件数
	Limit  int `json:"limit"  example:"20"`
	Offset int `json:"offset" example:"0"`
}
