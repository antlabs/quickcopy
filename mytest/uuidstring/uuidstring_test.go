package uuidstring

import (
	"testing"

	"github.com/google/uuid"
)

type SourceStruct struct {
	ID uuid.UUID
}

type TargetStruct struct {
	ID string
}

// :quickcopy
func CopyToTarget(dst *TargetStruct, src *SourceStruct) {

	dst.ID = func(u uuid.UUID) string {
		return u.String()
	}(src.ID)
}

// :quickcopy
func CopyToTarget2(dst *SourceStruct, src *TargetStruct) {

	dst.ID =
		func(s string) uuid.UUID {
			u, _ := uuid.Parse(s)
			return u

		}(src.ID)
}

func TestCopyToTarget(t *testing.T) {
	tests := []struct {
		name   string
		srcID  uuid.UUID
		wantID string
	}{
		{
			name:   "normal uuid",
			srcID:  uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
			wantID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		},
		{
			name:   "zero uuid",
			srcID:  uuid.Nil,
			wantID: "00000000-0000-0000-0000-000000000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := &SourceStruct{
				ID: tt.srcID,
			}
			dst := &TargetStruct{}

			CopyToTarget(dst, src)

			if dst.ID != tt.wantID {
				t.Errorf("CopyToTarget() got = %v, want %v", dst.ID, tt.wantID)
			}
		})
	}
}

func TestCopyToTarget2(t *testing.T) {
	tests := []struct {
		name    string
		srcID   string
		wantID  uuid.UUID
		wantErr bool
	}{
		{
			name:    "normal uuid string",
			srcID:   "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			wantID:  uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
			wantErr: false,
		},
		{
			name:    "empty string",
			srcID:   "",
			wantID:  uuid.Nil,
			wantErr: false,
		},
		{
			name:    "invalid uuid string",
			srcID:   "invalid-uuid",
			wantID:  uuid.Nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := &TargetStruct{
				ID: tt.srcID,
			}
			dst := &SourceStruct{}

			CopyToTarget2(dst, src)

			if dst.ID != tt.wantID {
				t.Errorf("CopyToTarget2() got = %v, want %v", dst.ID, tt.wantID)
			}
		})
	}
}
