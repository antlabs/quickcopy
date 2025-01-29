package uuidstring

import "github.com/google/uuid"

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
