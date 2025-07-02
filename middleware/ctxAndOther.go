package middleware

import (
	"context"
	"log"
	"naevis/globals"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func WithContext(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()
		// Example: add values to context
		ctx = context.WithValue(ctx, globals.ParamIDKey, ps)
		r = r.WithContext(ctx)
		h(w, r, ps)
	}
}

// router.GET("/health", withContext(Index))

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("🔥 Panic recovered: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// return recoverMiddleware(loggingMiddleware(securityHeaders(c.Handler(router))))
