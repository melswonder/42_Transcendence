package domain

import "errors"

// 層をまたいで扱う失敗の種類。
// infrastructure は DB 固有のエラー（pq のエラーコードなど）をここに翻訳し、
// usecase は errors.Is で分岐する。これで usecase は DB の存在を知らずに済む。
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrSessionNotFound   = errors.New("session not found or expired")
	ErrHandleTaken       = errors.New("handle already taken")
	ErrEmailTaken        = errors.New("email already registered")
	ErrOAuthAccountTaken = errors.New("oauth account already linked")
	ErrHandleUnavailable = errors.New("could not allocate a unique handle")
)
