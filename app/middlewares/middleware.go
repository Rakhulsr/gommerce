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
	"github.com/Rakhulsr/go-ecommerce/app/utils/calc"
	"github.com/Rakhulsr/go-ecommerce/app/utils/sessions"
	"golang.org/x/crypto/bcrypt"
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
			currentUserID := sessionStore.GetUserID(w, r)

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
					userWithAddresses, err := userRepo.GetUserByIDWithAddresses(ctx, activeUserID)
					if err != nil {
						log.Printf("AuthAndCartSessionMiddleware: Failed to load addresses for user %s: %v. Proceeding without addresses.", activeUserID, err)
						loggedInUser = user
					} else {
						loggedInUser = userWithAddresses
					}
				}
			} else {

				rememberTokenFromCookie, err := helpers.GetCookie(r, helpers.RememberMeCookieName)
				if err == nil && rememberTokenFromCookie != "" {
					selector, verifier, splitErr := helpers.SplitRememberToken(rememberTokenFromCookie)
					if splitErr != nil {
						log.Printf("AuthAndCartSessionMiddleware: Invalid remember token format, clearing cookie: %v", splitErr)
						helpers.ClearCookie(w, helpers.RememberMeCookieName)
					} else {
						user, findErr := userRepo.FindBySelector(ctx, selector)
						if findErr != nil || user == nil {
							log.Printf("AuthAndCartSessionMiddleware: Error finding user by selector: %v", findErr)
							helpers.ClearCookie(w, helpers.RememberMeCookieName)
						} else {
							if bcrypt.CompareHashAndPassword([]byte(user.RememberTokenHash), []byte(verifier)) != nil {
								log.Printf("AuthAndCartSessionMiddleware: Verifier mismatch, clearing cookie.")
								helpers.ClearCookie(w, helpers.RememberMeCookieName)
							} else {
								activeUserID = user.ID
								loggedInUser = user
								newSelector, newVerifierRaw, newToken, genErr := helpers.GenerateRememberTokenParts()
								if genErr == nil {
									hashedVerifier := helpers.HashPassword(newVerifierRaw)
									_ = userRepo.UpdateRememberToken(ctx, user.ID, newSelector, hashedVerifier)
									helpers.SetCookie(w, helpers.RememberMeCookieName, newToken, 8*time.Hour)
								}
								_ = sessionStore.SetUserID(w, r, user.ID)
							}
						}
					}
				}
			}

			if activeUserID != "" {

				cartIDFromSession := sessionStore.GetCartID(w, r)

				cart, err := cartRepo.GetOrCreateCartByUserID(ctx, cartIDFromSession, activeUserID)
				if err != nil {
					log.Printf("AuthAndCartSessionMiddleware: Failed to get or create cart for user %s: %v", activeUserID, err)

					activeCartID = ""

					if cartIDFromSession != "" {
						sessionStore.ClearCartID(w, r)
					}
				} else if cart != nil {
					activeCartID = cart.ID

					if cartIDFromSession != activeCartID {
						if err := sessionStore.SetCartID(w, r, activeCartID); err != nil {
							log.Printf("AuthAndCartSessionMiddleware: Failed to set cart ID %s in session: %v", activeCartID, err)
						}
					}

				} else {

					activeCartID = ""
					if cartIDFromSession != "" {
						sessionStore.ClearCartID(w, r)
					}

				}
			} else {

				if sessionStore.GetCartID(w, r) != "" {
					sessionStore.ClearCartID(w, r)
					log.Printf("AuthAndCartSessionMiddleware: User not logged in, clearing cart_id from session.")
				}
				activeCartID = ""
			}

			requestPath := r.URL.Path
			requiresLoginPaths := []string{
				"/carts", "/carts/add", "/checkout", "/carts/delete", "/profile", "/addresses", "/orders", "/payment", "/shipment",
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

				ctx := context.WithValue(r.Context(), helpers.CartCountKey, 0)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			cart, err := cartRepo.GetCartWithItems(r.Context(), cartID)
			if err != nil {
				log.Printf("CartCountMiddleware: Error getting cart with items for cartID '%s': %v. Setting count to 0.", cartID, err)
				ctx := context.WithValue(r.Context(), helpers.CartCountKey, 0)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if cart == nil {
				log.Printf("CartCountMiddleware: Cart with ID '%s' not found in DB (nil). Setting count to 0.", cartID)
				ctx := context.WithValue(r.Context(), helpers.CartCountKey, 0)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			cart.CalculateTotals(calc.GetTaxPercent())
			count := cart.TotalItems
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

func ContentSecurityPolicyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Kebijakan CSP yang diizinkan:
		// default-src: 'self' mengizinkan sumber daya dari domain yang sama
		// script-src: 'self' mengizinkan skrip dari domain yang sama
		//               https://app.sandbox.midtrans.com mengizinkan skrip dari domain Midtrans
		//               https://cdn.jsdelivr.net (untuk SweetAlert2)
		//               'unsafe-inline' mungkin diperlukan untuk beberapa inline script (hati-hati)
		//               'unsafe-eval' diperlukan karena Midtrans Snap menggunakan eval()
		// connect-src: 'self' mengizinkan koneksi dari domain yang sama
		//                https://app.sandbox.midtrans.com mengizinkan koneksi ke domain Midtrans
		//                https://api.sandbox.midtrans.com mengizinkan koneksi ke API Midtrans
		//                https://snap.midtrans.com (jika ada)
		//                https://snap.i.b-id-ca-eks-01.gopay.sh (dari log Anda, tambahkan ini juga)
		// frame-src: https://app.sandbox.midtrans.com mengizinkan iframe dari domain Midtrans
		// img-src: 'self' data: https://app.sandbox.midtrans.com (gambar Midtrans)
		// style-src: 'self' 'unsafe-inline'
		// font-src: 'self' https://cdnjs.cloudflare.com (untuk Font Awesome)

		csp := "default-src 'self';" +
			"script-src 'self' https://app.sandbox.midtrans.com https://cdn.jsdelivr.net 'unsafe-inline' 'unsafe-eval';" +
			"connect-src 'self' https://app.sandbox.midtrans.com https://api.sandbox.midtrans.com https://snap.midtrans.com https://snap.i.b-id-ca-eks-01.gopay.sh;" +
			"frame-src https://app.sandbox.midtrans.com;" +
			"img-src 'self' data: https://app.sandbox.midtrans.com;" +
			"style-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com;" +
			"font-src 'self' https://cdnjs.cloudflare.com;" +
			"object-src 'none'; " +
			"base-uri 'self';"

		w.Header().Set("Content-Security-Policy", csp)
		next.ServeHTTP(w, r)
	})
}
