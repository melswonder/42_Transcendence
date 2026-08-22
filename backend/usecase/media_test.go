package usecase

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"sync"
	"testing"

	"github.com/google/uuid"

	"transcendence-backend/domain"
)

// fakeMediaRepo / fakeFileStore はメモリ上の実装。
type fakeMediaRepo struct {
	mu     sync.Mutex
	assets map[uuid.UUID]*domain.MediaAsset
	keys   map[uuid.UUID]string
}

func newFakeMediaRepo() *fakeMediaRepo {
	return &fakeMediaRepo{
		assets: make(map[uuid.UUID]*domain.MediaAsset),
		keys:   make(map[uuid.UUID]string),
	}
}

func (r *fakeMediaRepo) CreateAsset(_ context.Context, asset *domain.MediaAsset, storageKey string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copied := *asset
	r.assets[asset.ID] = &copied
	r.keys[asset.ID] = storageKey
	return nil
}

func (r *fakeMediaRepo) GetOwnedAsset(_ context.Context, assetID, ownerID uuid.UUID) (*domain.MediaAsset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.assets[assetID]
	if !ok || a.OwnerUserID != ownerID {
		return nil, domain.ErrMediaNotFound
	}
	return a, nil
}

func (r *fakeMediaRepo) GetActiveAsset(_ context.Context, assetID uuid.UUID) (*domain.MediaAsset, string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.assets[assetID]
	if !ok || a.Status != domain.MediaStatusActive {
		return nil, "", domain.ErrMediaNotFound
	}
	return a, r.keys[assetID], nil
}

func (r *fakeMediaRepo) ListAssets(_ context.Context, ownerID uuid.UUID, _ string, _, _ int) ([]domain.MediaAsset, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []domain.MediaAsset
	for _, a := range r.assets {
		if a.OwnerUserID == ownerID && a.Status == domain.MediaStatusActive {
			out = append(out, *a)
		}
	}
	return out, len(out), nil
}

func (r *fakeMediaRepo) SoftDeleteAsset(_ context.Context, assetID, ownerID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.assets[assetID]
	if !ok || a.OwnerUserID != ownerID || a.Status != domain.MediaStatusActive {
		return domain.ErrMediaNotFound
	}
	a.Status = domain.MediaStatusDeleted
	return nil
}

type fakeFileStore struct {
	mu    sync.Mutex
	files map[string][]byte
}

func newFakeFileStore() *fakeFileStore {
	return &fakeFileStore{files: make(map[string][]byte)}
}

func (s *fakeFileStore) Save(key string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[key] = data
	return nil
}

func (s *fakeFileStore) Open(key string) (io.ReadSeekCloser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, ok := s.files[key]
	if !ok {
		return nil, domain.ErrMediaNotFound
	}
	return nopReadSeekCloser{bytes.NewReader(data)}, nil
}

type nopReadSeekCloser struct{ *bytes.Reader }

func (nopReadSeekCloser) Close() error { return nil }

func encodePNG(t *testing.T, w, h int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func encodeJPEG(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2)), nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestUploadAvatar(t *testing.T) {
	t.Parallel()

	owner := uuid.New()

	tests := []struct {
		name    string
		data    func(t *testing.T) []byte
		wantErr error
	}{
		{"PNG は通る", func(t *testing.T) []byte { return encodePNG(t, 3, 5) }, nil},
		{"JPEG は通る", func(t *testing.T) []byte { return encodeJPEG(t) }, nil},
		{"サイズ超過は拒否", func(_ *testing.T) []byte { return make([]byte, domain.AvatarMaxBytes+1) }, domain.ErrMediaTooLarge},
		{"空ファイルは拒否", func(_ *testing.T) []byte { return nil }, domain.ErrInvalidImage},
		{"画像以外は拒否", func(_ *testing.T) []byte { return []byte("<html>not an image</html>") }, domain.ErrUnsupportedMediaType},
		{
			// PNG のマジックナンバーだけ付けた偽装ファイル。sniff は通るがデコードで落ちる。
			"壊れた画像は拒否",
			func(_ *testing.T) []byte {
				return append([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}, bytes.Repeat([]byte{0}, 64)...)
			},
			domain.ErrInvalidImage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := NewMediaUsecase(newFakeMediaRepo(), newFakeFileStore())
			asset, err := uc.UploadAvatar(context.Background(), owner, "avatar.png", tt.data(t))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("UploadAvatar() error = %v, want %v", err, tt.wantErr)
			}
			if err == nil {
				if asset.Purpose != domain.MediaPurposeAvatar || asset.Status != domain.MediaStatusActive {
					t.Errorf("メタ情報が変: %+v", asset)
				}
				if asset.Width == nil || asset.Height == nil {
					t.Error("画像の寸法が入るはず")
				}
			}
		})
	}
}

func TestUploadAvatarStoresDimensions(t *testing.T) {
	t.Parallel()

	uc := NewMediaUsecase(newFakeMediaRepo(), newFakeFileStore())
	asset, err := uc.UploadAvatar(context.Background(), uuid.New(), "a.png", encodePNG(t, 7, 11))
	if err != nil {
		t.Fatal(err)
	}
	if *asset.Width != 7 || *asset.Height != 11 {
		t.Errorf("寸法はデコード結果から入るはず: %d x %d", *asset.Width, *asset.Height)
	}
	if asset.MimeType != "image/png" {
		t.Errorf("MIME は sniff の結果になるはず: %s", asset.MimeType)
	}
}

func TestDeleteAvatarIsIdempotentPerAsset(t *testing.T) {
	t.Parallel()

	repo := newFakeMediaRepo()
	uc := NewMediaUsecase(repo, newFakeFileStore())
	owner := uuid.New()
	asset, err := uc.UploadAvatar(context.Background(), owner, "a.png", encodePNG(t, 2, 2))
	if err != nil {
		t.Fatal(err)
	}

	if err := uc.Delete(context.Background(), asset.ID, owner); err != nil {
		t.Fatalf("削除できるはず: %v", err)
	}
	// 2 回目・他人のものは見つからない扱い。
	if err := uc.Delete(context.Background(), asset.ID, owner); !errors.Is(err, domain.ErrMediaNotFound) {
		t.Errorf("削除済みは not found のはず: %v", err)
	}
	if _, _, err := uc.repo.GetActiveAsset(context.Background(), asset.ID); !errors.Is(err, domain.ErrMediaNotFound) {
		t.Errorf("削除後は配信されないはず: %v", err)
	}
}
