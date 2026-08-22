package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// アバター画像の制限。API 仕様（apispec/media.go）と揃える。
const (
	AvatarMaxBytes     = 5 << 20 // 5MB
	MediaPurposeAvatar = "avatar"
)

// AllowedAvatarMIMEs はアバターとして受け付ける画像形式。
// 判定は拡張子や申告ではなく、実際のバイト列から行う。
var AllowedAvatarMIMEs = []string{"image/png", "image/jpeg", "image/webp"}

// media_assets.status の値。
const (
	MediaStatusActive  = "active"
	MediaStatusDeleted = "deleted"
)

// アバターまわりの失敗。handler は errors.Is で 413 / 415 / 400 に振り分ける。
var (
	ErrMediaNotFound        = errors.New("media asset not found")
	ErrMediaTooLarge        = errors.New("media file too large")
	ErrUnsupportedMediaType = errors.New("unsupported media type")
	ErrInvalidImage         = errors.New("file is not a valid image")
	ErrInvalidAvatarAsset   = errors.New("avatar asset is not usable")
	ErrInvalidDisplayName   = errors.New("invalid display name")
	ErrInvalidHandle        = errors.New("invalid handle")
	ErrInvalidLocale        = errors.New("unsupported locale")
)

// MediaAsset はアップロードされた 1 ファイル。storage_key は外に出さない。
type MediaAsset struct {
	ID               uuid.UUID
	OwnerUserID      uuid.UUID
	Purpose          string
	OriginalFilename string
	MimeType         string
	SizeBytes        int64
	Width            *int
	Height           *int
	ChecksumSHA256   string
	Status           string
	CreatedAt        time.Time
}
