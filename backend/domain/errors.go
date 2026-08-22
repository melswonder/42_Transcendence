package domain

import "errors"

// 層をまたいで扱う失敗の種類。
// infrastructure は DB 固有のエラー（pq のエラーコードなど）をここに翻訳し、
// usecase は errors.Is で分岐する。これで usecase は DB の存在を知らずに済む。
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrSessionNotFound    = errors.New("session not found or expired")
	ErrHandleTaken        = errors.New("handle already taken")
	ErrEmailTaken         = errors.New("email already registered")
	ErrOAuthAccountTaken  = errors.New("oauth account already linked")
	ErrHandleUnavailable  = errors.New("could not allocate a unique handle")
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrWeakPassword       = errors.New("password does not meet requirements")
	ErrInvalidCredentials = errors.New("invalid email or password")

	ErrInvalidMatchMode    = errors.New("invalid match mode")
	ErrInvalidOpponent     = errors.New("invalid opponent id")
	ErrInvalidResultType   = errors.New("invalid result type")
	ErrInvalidMatchPeriod  = errors.New("finished_at is before started_at")
	ErrInvalidParticipants = errors.New("a match must have exactly two distinct participants")
	ErrInvalidOutcome      = errors.New("invalid outcome")
	ErrInconsistentOutcome = errors.New("participants' outcomes do not add up")
	ErrInvalidDateRange    = errors.New("to is before from")
)
