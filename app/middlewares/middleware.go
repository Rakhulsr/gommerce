package middlewares

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/helpers"
	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/utils/sessions"
)

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

func AuthAndCartSessionMiddleware(userRepo repositories.UserRepositoryImpl, cartRepo repositories.CartRepositoryImpl, sessionStore sessions.SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			currentUserID := sessionStore.GetUserID(r)

			var activeUserID string
			var activeCartID string
			var loggedInUser *models.User

			if currentUserID != "" {
				user, err := userRepo.FindByID(ctx, currentUserID)
				if err != nil || user == nil {
					log.Printf("AuthAndCartSessionMiddleware: User %s from session not found or error: %v. Clearing session.", currentUserID, err)
					sessionStore.ClearSession(w, r)
					helpers.ClearCookie(w, helpers.RememberMeCookieName)
				} else {
					activeUserID = user.ID
					loggedInUser = user
				}
			} else {

				rememberTokenFromCookie, err := helpers.GetCookie(r, helpers.RememberMeCookieName)
				if err == nil && rememberTokenFromCookie != "" {
					user, findErr := userRepo.FindByRememberToken(ctx, rememberTokenFromCookie)
					if findErr != nil {
						log.Printf("AuthAndCartSessionMiddleware: Error finding user by remember token: %v. Clearing cookie.", findErr)
						helpers.ClearCookie(w, helpers.RememberMeCookieName)
					}
					if user != nil {
						activeUserID = user.ID
						loggedInUser = user

						if err := sessionStore.SetUserID(w, r, activeUserID); err != nil {
							log.Printf("AuthAndCartSessionMiddleware: Failed to set session after remember me for user %s: %v", activeUserID, err)
						} else {

							newSelector, newVerifierRaw, newRememberTokenString, genErr := helpers.GenerateRememberTokenParts()
							if genErr != nil {
								log.Printf("AuthAndCartSessionMiddleware: Failed to generate new remember token parts for user %s: %v", activeUserID, genErr)
							} else {
								hashedVerifier := helpers.HashPassword(newVerifierRaw)
								if updateErr := userRepo.UpdateRememberToken(ctx, activeUserID, newSelector, hashedVerifier); updateErr != nil {
									log.Printf("AuthAndCartSessionMiddleware: Failed to update remember token in DB for user %s: %v", activeUserID, updateErr)
								}
								helpers.SetCookie(w, helpers.RememberMeCookieName, newRememberTokenString, 30*24*time.Hour)
							}
						}
					} else {

						helpers.ClearCookie(w, helpers.RememberMeCookieName)
					}
				}
			}

			if activeUserID != "" {
				currentCartID := sessionStore.GetCartID(r)
				if currentCartID == "" {
					userCart, err := cartRepo.GetOrCreateCartByUserID(ctx, activeUserID)
					if err != nil {
						log.Printf("AuthAndCartSessionMiddleware: Failed to get or create cart for logged-in user %s: %v", activeUserID, err)
					} else {
						activeCartID = userCart.ID
						if err := sessionStore.SetCartID(w, r, activeCartID); err != nil {
							log.Printf("AuthAndCartSessionMiddleware: Failed to set user cart ID %s in session: %v", activeCartID, err)
						}
					}
				} else {
					activeCartID = currentCartID
				}
			} else {

				if sessionStore.GetCartID(r) != "" {
					sessionStore.ClearCartID(w, r)
				}
				activeCartID = ""
			}

			requestPath := r.URL.Path
			requiresLoginPaths := []string{
				"/carts",
				"/carts/add",
				"/checkout",
				"/carts/delete",
				"/profile",
				"/addresses",
				"/orders",
				"/payment",
				"/shipment",
			}
			shouldRedirect := false
			for _, p := range requiresLoginPaths {
				if strings.HasPrefix(requestPath, p) {
					shouldRedirect = true
					break
				}
			}

			if shouldRedirect && activeUserID == "" &&
				requestPath != "/login" && requestPath != "/register" &&
				!strings.HasPrefix(requestPath, "/forgot-password") &&
				!strings.HasPrefix(requestPath, "/reset-password") {
				log.Printf("AuthAndCartSessionMiddleware: User not logged in, redirecting %s to /login", requestPath)
				http.Redirect(w, r, fmt.Sprintf("/login?status=warning&message=%s", url.QueryEscape("Anda harus login untuk mengakses halaman ini.")), http.StatusSeeOther)
				return
			}

			ctx = context.WithValue(ctx, helpers.ContextKeyUserID, activeUserID)
			ctx = context.WithValue(ctx, helpers.ContextKeyCartID, activeCartID)
			ctx = context.WithValue(ctx, helpers.ContextKeyUser, loggedInUser)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CartCountMiddleware(cartRepo repositories.CartRepositoryImpl) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cartID, ok := r.Context().Value(helpers.ContextKeyCartID).(string)
			if !ok || cartID == "" {
				log.Printf("CartCountMiddleware: CartID not found in context for %s. Setting count to 0.", r.URL.Path)
				ctx := context.WithValue(r.Context(), helpers.CartCountKey, 0)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			count, err := cartRepo.GetCartItemCount(r.Context(), cartID)
			if err != nil {
				log.Printf("CartCountMiddleware: Error getting cart item count for cartID %s: %v", cartID, err)
				count = 0
			}

			ctx := context.WithValue(r.Context(), helpers.CartCountKey, count)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AuthRequiredMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if userID, ok := r.Context().Value(helpers.ContextKeyUserID).(string); !ok || userID == "" {
			http.Redirect(w, r, fmt.Sprintf("/login?status=warning&message=%s", url.QueryEscape("Anda harus login untuk mengakses halaman ini.")), http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
