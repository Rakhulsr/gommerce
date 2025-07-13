package middlewares

import (
	"log"
	"net/http"
	"net/url"

	"github.com/Rakhulsr/go-ecommerce/app/helpers"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
)

func AdminAuthMiddleware(userRepo repositories.UserRepositoryImpl) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(helpers.ContextKeyUserID).(string)
			if !ok || userID == "" {
				log.Println("AdminAuthMiddleware: User ID not found in context or empty. Redirecting to login.")

				http.Redirect(w, r, "/login?status=error&message="+url.QueryEscape("Anda harus login untuk mengakses admin panel."), http.StatusFound)
				return
			}

			user, err := userRepo.FindByID(r.Context(), userID)
			if err != nil || user == nil {
				log.Printf("AdminAuthMiddleware: Error finding user %s: %v. Redirecting to login.", userID, err)
				http.Redirect(w, r, "/login?status=error&message="+url.QueryEscape("Pengguna tidak ditemukan atau sesi tidak valid."), http.StatusFound)
				return
			}

			if user.Role != "admin" { // Pastikan field dan nilai role sesuai
				log.Printf("AdminAuthMiddleware: User %s (%s) attempted to access admin panel without admin role.", user.ID, user.Email)
				http.Redirect(w, r, "/?status=error&message="+url.QueryEscape("Anda tidak memiliki izin untuk mengakses halaman ini."), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
