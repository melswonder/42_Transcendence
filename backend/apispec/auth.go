// 認証まわりの公開 API の仕様。sessions / oauth_accounts テーブル（§3）に対応する。
//
// 外部クライアントから叩けるようにするため、セッションは Bearer トークンで受け渡す。
// DB には raw token を置かず sessions.token_hash だけを保存するので、
// トークンの平文を返すのは発行時（register / login / refresh / OAuth コールバック）だけになる。
package apispec

import "time"

// RegisterRequest はメール+パスワードでの新規登録。
type RegisterRequest struct {
	Email           string `json:"email"            binding:"required" format:"email" example:"user@example.com"`
	Password        string `json:"password"         binding:"required" minLength:"8" example:"correct horse battery staple"`
	DisplayName     string `json:"display_name"     binding:"required" maxLength:"50" example:"melswonder"`
	Handle          string `json:"handle"           binding:"required" maxLength:"30" example:"mels"`
	PreferredLocale string `json:"preferred_locale" enums:"ja,en,fr" example:"ja"`
}

// LoginRequest はメール+パスワードでのログイン。
type LoginRequest struct {
	Email    string `json:"email"    binding:"required" format:"email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"correct horse battery staple"`
}

// TokenResponse はセッション発行時のレスポンス。
// access_token は再取得できないので、クライアント側で保持する。
type TokenResponse struct {
	AccessToken string      `json:"access_token" example:"v1.f7a1c0b9d2e34f56a7b8c9d0e1f23456"`
	TokenType   string      `json:"token_type"   example:"Bearer"`
	ExpiresAt   time.Time   `json:"expires_at"`
	User        UserPrivate `json:"user"`
}

// SessionResponse は sessions の 1 行。token_hash は決して返さない。
type SessionResponse struct {
	ID         string    `json:"id"      format:"uuid" example:"c1a2b3d4-5e6f-4071-8293-a4b5c6d7e8f9"`
	Current    bool      `json:"current" example:"true"` // このリクエストに使われているセッションか
	ExpiresAt  time.Time `json:"expires_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// SessionListResponse は有効なセッションの一覧。
type SessionListResponse struct {
	Items []SessionResponse `json:"items"`
}

// OAuthAccountResponse は oauth_accounts の 1 行。
type OAuthAccountResponse struct {
	ID                string    `json:"id"                  format:"uuid"`
	Provider          string    `json:"provider"            enums:"google,github,42" example:"google"`
	ProviderAccountID string    `json:"provider_account_id" example:"117452839201847362514"`
	ProviderEmail     *string   `json:"provider_email"      format:"email" example:"user@gmail.com"`
	CreatedAt         time.Time `json:"created_at"`
}

// OAuthAccountListResponse は連携済みプロバイダの一覧。
type OAuthAccountListResponse struct {
	Items []OAuthAccountResponse `json:"items"`
}

// OAuthAuthorizeResponse は OAuth 認可 URL。
// ブラウザ以外のクライアントがリダイレクトを追えるように URL を JSON でも返す。
type OAuthAuthorizeResponse struct {
	AuthorizationURL string `json:"authorization_url" example:"https://accounts.google.com/o/oauth2/v2/auth?..."`
	State            string `json:"state"             example:"9f8e7d6c5b4a"`
}

// Register godoc
//
//	@Summary		新規登録
//	@Description	メールとパスワードでユーザーを作り、そのままセッションを発行する。
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		RegisterRequest	true	"登録内容"
//	@Success		201		{object}	TokenResponse
//	@Failure		400		{object}	ErrorResponse	"バリデーションエラー"
//	@Failure		409		{object}	ErrorResponse	"email または handle が既に使われている"
//	@Router			/auth/register [post]
func Register() {}

// Login godoc
//
//	@Summary		ログイン
//	@Description	認証に成功するとセッションを 1 件作り、Bearer トークンを返す。
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		LoginRequest	true	"認証情報"
//	@Success		200		{object}	TokenResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse	"メールまたはパスワードが違う"
//	@Failure		403		{object}	ErrorResponse	"アカウントが停止されている"
//	@Router			/auth/login [post]
func Login() {}

// Logout godoc
//
//	@Summary		ログアウト
//	@Description	リクエストに使われたセッションを失効させる（sessions.revoked_at）。
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		204	"失効完了"
//	@Failure		401	{object}	ErrorResponse
//	@Router			/auth/logout [post]
func Logout() {}

// Refresh godoc
//
//	@Summary		セッションを更新
//	@Description	有効期限を延長した新しいトークンを発行し、古いトークンは失効させる。
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	TokenResponse
//	@Failure		401	{object}	ErrorResponse	"トークンが無効・期限切れ・失効済み"
//	@Router			/auth/refresh [post]
func Refresh() {}

// ListSessions godoc
//
//	@Summary		自分のセッション一覧
//	@Description	有効なセッション（未失効かつ期限内）を返す。他端末のログイン状況の確認に使う。
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	SessionListResponse
//	@Failure		401	{object}	ErrorResponse
//	@Router			/auth/sessions [get]
func ListSessions() {}

// RevokeSession godoc
//
//	@Summary		セッションを失効させる
//	@Description	指定した端末のセッションだけを切る。自分のセッションのみ操作できる。
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Param			sessionId	path	string	true	"セッション ID"	format(uuid)
//	@Success		204			"失効完了"
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse	"自分のセッションに存在しない"
//	@Router			/auth/sessions/{sessionId} [delete]
func RevokeSession() {}

// OAuthAuthorize godoc
//
//	@Summary		OAuth 認可を開始
//	@Description	プロバイダの認可 URL を返す。Accept: text/html のブラウザからは 302 でリダイレクトする。
//	@Tags			auth
//	@Produce		json
//	@Param			provider		path		string	true	"プロバイダ"	Enums(google, github, 42)
//	@Param			redirect_uri	query		string	false	"認可後に戻すフロントエンドの URL（許可リスト内のみ）"
//	@Success		200				{object}	OAuthAuthorizeResponse
//	@Failure		400				{object}	ErrorResponse	"未対応のプロバイダ"
//	@Router			/auth/oauth/{provider} [get]
func OAuthAuthorize() {}

// OAuthCallback godoc
//
//	@Summary		OAuth コールバック
//	@Description	認可コードを検証し、oauth_accounts に紐づくユーザーでセッションを発行する。未登録なら新規作成する。
//	@Tags			auth
//	@Produce		json
//	@Param			provider	path		string	true	"プロバイダ"	Enums(google, github, 42)
//	@Param			code		query		string	true	"認可コード"
//	@Param			state		query		string	true	"CSRF 対策の state"
//	@Success		200			{object}	TokenResponse
//	@Failure		400			{object}	ErrorResponse	"code / state が不正"
//	@Failure		409			{object}	ErrorResponse	"同じプロバイダアカウントが別ユーザーに紐づいている"
//	@Router			/auth/oauth/{provider}/callback [get]
func OAuthCallback() {}

// ListOAuthAccounts godoc
//
//	@Summary	連携済み OAuth アカウント一覧
//	@Tags		auth
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	OAuthAccountListResponse
//	@Failure	401	{object}	ErrorResponse
//	@Router		/auth/oauth/accounts [get]
func ListOAuthAccounts() {}

// UnlinkOAuthAccount godoc
//
//	@Summary		OAuth 連携を解除
//	@Description	解除するとログイン手段が無くなる場合（パスワード未設定で連携が 1 件だけ）は 409 を返す。
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Param			accountId	path	string	true	"oauth_accounts の ID"	format(uuid)
//	@Success		204			"解除完了"
//	@Failure		401			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		409			{object}	ErrorResponse	"最後のログイン手段は解除できない"
//	@Router			/auth/oauth/accounts/{accountId} [delete]
func UnlinkOAuthAccount() {}
