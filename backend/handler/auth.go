// 認証の HTTP 入口。Cookie の出し入れとリダイレクトだけを担い、判断は usecase に任せる。
package handler

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

// Cookie 名。
const (
	sessionCookie = "session"
	stateCookie   = "oauth_state"
	nonceCookie   = "oauth_nonce"
)

// state / nonce は Google を往復する数十秒しか使わないので短命にする。
const oauthFlowTTL = 10 * time.Minute

type AuthConfig struct {
	// ログイン後に戻すフロントの URL。
	FrontendURL string
	// 本番 (HTTPS) では true。Cookie に Secure 属性を付ける。
	SecureCookie bool
}

// AuthHandler は認証の HTTP 入口。Cookie の出し入れとリダイレクトだけを担い、判断は usecase に任せる。
// Start は GET /auth/google。state / nonce を Cookie に預けて、Google の同意画面へ 302 で送る。
// Callback は GET /auth/google/callback。state を照合してログインを成立させ、セッション Cookie を配る。
// Me は GET /auth/me。ログイン中のユーザーを JSON で返す。未ログインなら 401。
// Logout は POST /auth/logout。セッションを失効させて Cookie を消す。
type AuthHandler struct {
	uc  *usecase.AuthUsecase
	cfg AuthConfig
}

func NewAuthHandler(uc *usecase.AuthUsecase, cfg AuthConfig) *AuthHandler {
	return &AuthHandler{uc: uc, cfg: cfg}
}

// Start - GET /auth/google
//
// ここでは何も返さない。302 でブラウザを Google の同意画面に送るだけ。
// 照合用の state / nonce は Cookie に預けておき、callback で突き合わせる。
func (h *AuthHandler) Start(w http.ResponseWriter, r *http.Request) {
	login := h.uc.StartLogin()

	h.setCookie(w, stateCookie, login.State, oauthFlowTTL)
	h.setCookie(w, nonceCookie, login.Nonce, oauthFlowTTL)

	http.Redirect(w, r, login.AuthURL, http.StatusFound)
}

// Callback - GET /auth/google/callback
//
// Google がユーザーのブラウザを送り返してくる先。?code=... が付いている。
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	// 使い終わった一時 Cookie は成否によらず消す。
	defer h.clearCookie(w, stateCookie)
	defer h.clearCookie(w, nonceCookie)

	// ユーザーが同意画面で「キャンセル」を押した場合はここに来る。
	if reason := r.URL.Query().Get("error"); reason != "" {
		h.redirectWithError(w, r, "access_denied")

		return
	}

	if err := h.verifyState(r); err != nil {
		log.Printf("oauth callback: %v", err)
		h.redirectWithError(w, r, "invalid_state")

		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		h.redirectWithError(w, r, "missing_code")

		return
	}

	nonce, err := r.Cookie(nonceCookie)
	if err != nil {
		h.redirectWithError(w, r, "invalid_state")

		return
	}

	result, err := h.uc.CompleteLogin(r.Context(), code, nonce.Value)
	if err != nil {
		log.Printf("oauth callback: complete login: %v", err)
		h.redirectWithError(w, r, "login_failed")

		return
	}

	// ここで初めてログイン成立。生トークンは Cookie にしか出さない。
	h.setCookie(w, sessionCookie, result.SessionToken, time.Until(result.ExpiresAt))

	http.Redirect(w, r, h.cfg.FrontendURL, http.StatusFound)
}

// Me - GET /auth/me
// ログイン中のユーザーを返す。フロントが起動時に「今ログインしてる？」を確かめる用。
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")

		return
	}

	writeJSON(w, http.StatusOK, userResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Handle:      user.Handle,
		Level:       user.Level,
	})
}

// Logout - POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		if err := h.uc.Logout(r.Context(), c.Value); err != nil {
			log.Printf("logout: %v", err)
		}
	}

	h.clearCookie(w, sessionCookie)
	w.WriteHeader(http.StatusNoContent)
}

// currentUser は Cookie からログイン中のユーザーを引く。他のハンドラからも使える。
func (h *AuthHandler) currentUser(r *http.Request) (*domain.User, error) {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}

	return h.uc.Authenticate(r.Context(), c.Value)
}

// verifyState は callback のクエリと Cookie の state を突き合わせる。
//
// これが無いと、攻撃者が自分の code を踏ませて
// 被害者を攻撃者のアカウントでログインさせられる (ログイン CSRF)。
func (h *AuthHandler) verifyState(r *http.Request) error {
	c, err := r.Cookie(stateCookie)
	if err != nil {
		return errors.New("state cookie is missing")
	}

	got := r.URL.Query().Get("state")
	if subtle.ConstantTimeCompare([]byte(c.Value), []byte(got)) != 1 {
		return errors.New("state mismatch")
	}

	return nil
}

func (h *AuthHandler) setCookie(w http.ResponseWriter, name, value string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:  name,
		Value: value,
		Path:  "/",
		// JS から読めないようにする。XSS でセッションを抜かれないため。
		HttpOnly: true,
		Secure:   h.cfg.SecureCookie,
		// Strict にすると Google からの戻りで Cookie が送られず state 照合が落ちる。
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(ttl.Seconds()),
	})
}

func (h *AuthHandler) clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.SecureCookie,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// redirectWithError は失敗理由をクエリに載せてフロントへ戻す。
// 詳細はログにだけ残し、外には粗い分類しか出さない。
func (h *AuthHandler) redirectWithError(w http.ResponseWriter, r *http.Request, reason string) {
	http.Redirect(w, r, h.cfg.FrontendURL+"/login?error="+reason, http.StatusFound)
}

type userResponse struct {
	ID          string  `json:"id"`
	Email       *string `json:"email"`
	DisplayName string  `json:"display_name"`
	Handle      string  `json:"handle"`
	Level       int     `json:"level"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// ここまででヘッダーは送信済みなのでステータスは変えられない。握りつぶさずログに残す。
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// writeJSONErrorCode は機械可読なコード付きの失敗レスポンス。
// 文言の翻訳はクライアントが code を鍵にして行う（サーバーはロケールを知らない）。
func writeJSONErrorCode(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"error": message, "code": code})
}
