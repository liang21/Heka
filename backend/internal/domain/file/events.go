package file

import (
	"time"

	"github.com/liang21/heka/internal/domain/shared"
)

type FileUploadedEvent struct {
	ProjectID shared.ID
	FileID    shared.ID
	time      time.Time
}

func (e *FileUploadedEvent) EventName() string {
	return "file.uploaded"
}

func (e *FileUploadedEvent) OccurredAt() time.Time {
	return e.time
}

type FileDeletedEvent struct {
	ProjectID shared.ID
	FileID    shared.ID
	time      time.Time
}

func (e *FileDeletedEvent) EventName() string {
	return "file.deleted"
}

func (e *FileDeletedEvent) OccurredAt() time.Time {
	return e.time
}
