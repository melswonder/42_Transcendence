package domain

import (
	"errors"
	"strings"
	"testing"
)

func strPtr(s string) *string { return &s }

func TestProfileUpdateValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		update  ProfileUpdate
		wantErr error
	}{
		{"何も変えない", ProfileUpdate{}, nil},
		{"表示名を変える", ProfileUpdate{DisplayName: strPtr("新しい名前")}, nil},
		{"空の表示名は拒否", ProfileUpdate{DisplayName: strPtr("   ")}, ErrInvalidDisplayName},
		{"長すぎる表示名は拒否", ProfileUpdate{DisplayName: strPtr(strings.Repeat("あ", DisplayNameMaxLen+1))}, ErrInvalidDisplayName},
		{"上限ちょうどの表示名は通る", ProfileUpdate{DisplayName: strPtr(strings.Repeat("あ", DisplayNameMaxLen))}, nil},
		{"handle を変える", ProfileUpdate{Handle: strPtr("new_handle_42")}, nil},
		{"短すぎる handle は拒否", ProfileUpdate{Handle: strPtr("ab")}, ErrInvalidHandle},
		{"大文字の handle は拒否", ProfileUpdate{Handle: strPtr("NewHandle")}, ErrInvalidHandle},
		{"記号入りの handle は拒否", ProfileUpdate{Handle: strPtr("bad-handle!")}, ErrInvalidHandle},
		{"対応ロケールは通る", ProfileUpdate{PreferredLocale: strPtr("fr")}, nil},
		{"未対応ロケールは拒否", ProfileUpdate{PreferredLocale: strPtr("de")}, ErrInvalidLocale},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.update.Validate(); !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
