package context

import (
	"context"

	"github.com/liang21/heka/internal/domain/shared"
)

type contextKey struct{ name string }

var (
	userIDKey    = &contextKey{name: "userID"}
	projectIDKey = &contextKey{name: "projectID"}
)

func WithUserID(ctx context.Context, id shared.ID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func UserIDFromContext(ctx context.Context) (shared.ID, bool) {
	id, ok := ctx.Value(userIDKey).(shared.ID)
	return id, ok
}

func WithProjectID(ctx context.Context, id shared.ID) context.Context {
	return context.WithValue(ctx, projectIDKey, id)
}

func ProjectIDFromContext(ctx context.Context) (shared.ID, bool) {
	id, ok := ctx.Value(projectIDKey).(shared.ID)
	return id, ok
}
