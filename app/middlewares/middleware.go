// app/middlewares/middleware.go (DIUBAH TOTAL UNTUK FLASH)

package middlewares

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/utils/sessions"
)

type contextKey string

const (
	CartCountKey contextKey = "cart_count"
)

func SessionManagerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("SessionManagerMiddleware: Menerima request untuk %s", r.URL.Path)

		next.ServeHTTP(w, r)

		session, err := sessions.Store.Get(r, sessions.SessionCartKey)
		if err != nil {
			log.Printf("SessionManagerMiddleware: Error getting session for saving on %s: %v", r.URL.Path, err)
			return
		}

		sessionID := session.ID
		if sessionID == "" {
			if cartID, ok := session.Values[sessions.CartSessionIDKey].(string); ok && cartID != "" {
				sessionID = "FALLBACK_CART_ID:" + cartID
			} else {
				sessionID = "EMPTY_SESSION_ID_AND_CART_ID"
			}
		}
		// log.Printf("SessionManagerMiddleware: Session saved successfully on %s. Session ID: %s", r.URL.Path, sessionID)

		if saveErr := session.Save(r, w); saveErr != nil {
			log.Printf("SessionManagerMiddleware: Error saving session on %s: %v", r.URL.Path, saveErr)
		}
	})
}

func CartCountMiddleware(cartRepo repositories.CartRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cartID, err := sessions.GetCartID(w, r)
			if err != nil {
				log.Printf("CartCountMiddleware: Error getting CartID: %v", err)
				next.ServeHTTP(w, r)
				return
			}
			// log.Printf("CartCountMiddleware: Using cart ID: %s for %s", cartID, r.URL.Path)
			count, err := cartRepo.GetCartItemCount(r.Context(), cartID)
			if err != nil {
				log.Printf("CartCountMiddleware: Error getting cart item count for cartID %s: %v", cartID, err)
				count = 0
			}

			ctx := context.WithValue(r.Context(), CartCountKey, count)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func MethodOverrideMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_ = r.ParseForm()
			override := r.Form.Get("_method")
			if override != "" {
				r.Method = strings.ToUpper(override)
			}
		}
		next.ServeHTTP(w, r)
	})
}
