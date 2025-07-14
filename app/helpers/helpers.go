package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/models/other"
	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const (
	ContextKeyUserID     contextKey = "userID"
	ContextKeyCartID     contextKey = "cartID"
	ContextKeyUser       contextKey = "userObject"
	CartCountKey         contextKey = "cart_count"
	RememberMeCookieName            = "remember_token"
	CSRFTokenKey         contextKey = "csrfToken"
)

func FormatRupiah(amount float64) string {
	return fmt.Sprintf("Rp %.0f", amount)
}

func GetBaseData(r *http.Request, pageSpecificData map[string]interface{}) map[string]interface{} {
	if pageSpecificData == nil {
		pageSpecificData = make(map[string]interface{})
	}

	if _, exists := pageSpecificData["Title"]; !exists {
		pageSpecificData["Title"] = "Toko Bulan"
	}
	if _, exists := pageSpecificData["CartCount"]; !exists {
		pageSpecificData["CartCount"] = 0
	}
	if _, exists := pageSpecificData["IsLoggedIn"]; !exists {
		pageSpecificData["IsLoggedIn"] = false
	}

	if _, exists := pageSpecificData["User"]; !exists {
		pageSpecificData["User"] = nil
	}

	if _, exists := pageSpecificData["UserID"]; !exists {
		pageSpecificData["UserID"] = ""
	}

	if _, exists := pageSpecificData["Breadcrumbs"]; !exists {
		pageSpecificData["Breadcrumbs"] = []breadcrumb.Breadcrumb{}
	}
	if _, exists := pageSpecificData["IsAuthPage"]; !exists {
		pageSpecificData["IsAuthPage"] = false
	}

	if _, exists := pageSpecificData["IsAdminPage"]; !exists {
		pageSpecificData["IsAdminPage"] = false
	}

	if _, exists := pageSpecificData["HideAdminWelcomeMessage"]; !exists {
		pageSpecificData["HideAdminWelcomeMessage"] = false
	}

	if _, exists := pageSpecificData["Query"]; !exists {
		pageSpecificData["Query"] = r.URL.Query()
	}

	if cartCountVal := r.Context().Value(CartCountKey); cartCountVal != nil {
		if count, ok := cartCountVal.(int); ok {
			pageSpecificData["CartCount"] = count
		} else {
			log.Printf("GetBaseData: CartCount in context is not of type int. Value: %+v", cartCountVal)
		}
	}

	if userVal := r.Context().Value(ContextKeyUser); userVal != nil {
		if user, ok := userVal.(*models.User); ok && user != nil {
			userForTemplate := &other.UserForTemplate{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Email:     user.Email,
				Role:      user.Role,
			}
			pageSpecificData["User"] = userForTemplate
			pageSpecificData["IsLoggedIn"] = true
			pageSpecificData["UserID"] = user.ID

			if user.Role == "admin" {
				pageSpecificData["IsAdminPage"] = true
			}
		} else {
			log.Printf("GetBaseData: User in context is not of type *models.User or is nil. Value: %+v", userVal)
			pageSpecificData["User"] = nil
			pageSpecificData["IsLoggedIn"] = false
			pageSpecificData["UserID"] = ""
			pageSpecificData["IsAdminPage"] = false
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		pageSpecificData["MessageStatus"] = status
	} else {
		pageSpecificData["MessageStatus"] = ""
	}
	if msg := r.URL.Query().Get("message"); msg != "" {
		pageSpecificData["Message"] = msg
	} else {
		pageSpecificData["Message"] = ""
	}

	return pageSpecificData
}

func FormatValidationErrors(errs validator.ValidationErrors) map[string]string {
	errorMessages := make(map[string]string)
	for _, err := range errs {
		field := strings.ToLower(err.Field())
		switch err.Tag() {
		case "required":
			errorMessages[field] = fmt.Sprintf("%s wajib diisi.", err.Field())
		case "email":
			errorMessages[field] = fmt.Sprintf("%s harus berupa alamat email yang valid.", err.Field())
		case "numeric":
			errorMessages[field] = fmt.Sprintf("%s harus berupa angka.", err.Field())
		case "min":
			errorMessages[field] = fmt.Sprintf("%s minimal %s karakter/nilai.", err.Field(), err.Param())
		case "max":
			errorMessages[field] = fmt.Sprintf("%s maksimal %s karakter/nilai.", err.Field(), err.Param())
		default:
			errorMessages[field] = fmt.Sprintf("Validasi %s gagal pada field %s.", err.Tag(), err.Field())
		}
	}
	return errorMessages
}

func capitalizeFirstLetter(s string) string {
	if len(s) == 0 {
		return ""
	}
	s = strings.ReplaceAll(s, "_", " ")
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

func SetCookie(w http.ResponseWriter, name, value string, expires time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  time.Now().Add(expires),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

func GetCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func ClearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Expires:  time.Now().AddDate(-1, 0, 0),
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

func GenerateRememberTokenParts() (selector string, verifier string, tokenString string, err error) {
	selectorBytes := make([]byte, 16)
	if _, err := rand.Read(selectorBytes); err != nil {
		return "", "", "", fmt.Errorf("failed to generate selector: %w", err)
	}
	selector = base64.URLEncoding.EncodeToString(selectorBytes)

	verifierBytes := make([]byte, 16)
	if _, err := rand.Read(verifierBytes); err != nil {
		return "", "", "", fmt.Errorf("failed to generate verifier: %w", err)
	}
	verifier = base64.URLEncoding.EncodeToString(verifierBytes)

	tokenString = fmt.Sprintf("%s.%s", selector, verifier)

	return selector, verifier, tokenString, nil
}

func GenerateResetToken() (string, time.Time, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, err
	}
	token := hex.EncodeToString(b)
	expiresAt := time.Now().Add(1 * time.Hour)
	return token, expiresAt, nil
}

func PasswordCompare(hashPass string, password []byte) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPass), password)
	if err != nil {

		log.Printf("PasswordCompare: password does not match or error: %v", err)
		return false
	}
	return true
}
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return ""
	}
	return string(bytes)
}

func GenerateSlug(s string) string {
	s = strings.ToLower(s)
	reg := regexp.MustCompile("[^a-z0-9]+")
	s = reg.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}
