package server

import (
	"context"
	"net/http"
)

type contextKey int

const (
	contextKeyUserId contextKey = iota
)

func (s *Server) authenticate(sh ServerHandler) ServerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionToken := r.Header.Get("session")
		if sessionToken == "" {
			writeErrorResponse(w, ErrInvalidRequest)
			return
		}

		userId, err := s.Store.GetSession(r.Context(), sessionToken)
		if err != nil {
			writeErrorResponse(w, ErrMustBeLoggedIn)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUserId, userId)
		r = r.WithContext(ctx)

		sh(w, r)
	}
}
