package httpx

import (
	"context"
	"net/http"

	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
)

type actorKey struct{}

func WithActor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor := domain.Actor{
			UserID: r.Header.Get("X-Qiling-User-ID"),
			Role:   r.Header.Get("X-Qiling-Role"),
		}
		if actor.UserID == "" {
			actor.UserID = "usr_001"
		}
		if actor.Role == "" {
			actor.Role = "sales"
		}
		ctx := context.WithValue(r.Context(), actorKey{}, actor)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ActorFromRequest(r *http.Request) domain.Actor {
	actor, ok := r.Context().Value(actorKey{}).(domain.Actor)
	if !ok {
		return domain.Actor{UserID: "usr_001", Role: "sales"}
	}
	return actor
}
