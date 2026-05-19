package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// setChiURLParam is a test helper that sets chi URL parameters in the request context
func setChiURLParam(r *http.Request, key, value string) *http.Request {
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
}
