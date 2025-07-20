package helpers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/models/other"
	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
	"github.com/Rakhulsr/go-ecommerce/app/utils/sessions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const (
	ContextKeyUserID     contextKey = "userID"
	ContextKeyCartID     contextKey = "cartID"
	ContextKeyUser       contextKey = "user"
	CartCountKey         contextKey = "cartCount"
	RememberMeCookieName            = "remember_token"
	CSRFTokenKey         contextKey = "csrfToken"
	ContextKeyIsLoggedIn contextKey = "isLoggedIn"
	ContextKeyUserRole   contextKey = "userRole"
)

func GetTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"formatCurrency":     FormatCurrency,
		"add":                Add,
		"sub":                Sub,
		"mul":                Mul,
		"div":                Div,
		"mod":                Mod,
		"eq":                 Eq,
		"ne":                 Ne,
		"lt":                 Lt,
		"le":                 Le,
		"gt":                 Gt,
		"ge":                 Ge,
		"urlQueryEscape":     URLQueryEscape,
		"extractProvince":    ExtractProvinceFromLocationName,
		"extractCity":        ExtractCityFromLocationName,
		"extractDistrict":    ExtractDistrictFromLocationName,
		"extractSubdistrict": ExtractSubdistrictFromLocationName,
		"orderStatusText":    OrderStatusText,
		"paymentStatusText":  PaymentStatusText,
	}
}

func ClearCartIDFromSession(w http.ResponseWriter, r *http.Request, sessionStore sessions.SessionStore) {
	session, err := sessionStore.GetSession(w, r)
	if err != nil {
		log.Printf("Error getting session to clear cart ID: %v", err)
		return
	}

	delete(session.Values, string(ContextKeyCartID))

	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Error saving session after clearing cart ID: %v", err)
	}
	log.Println("CartID cleared from session.")
}
func FormatRupiah(amount float64) string {
	return fmt.Sprintf("Rp %.0f", amount)
}

func GetCartIDFromContext(r *http.Request) string {
	if cartID, ok := r.Context().Value(ContextKeyCartID).(string); ok {
		return cartID
	}
	return ""
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
				Phone:     user.Phone,
				Role:      user.Role,
				Addresses: user.Address,
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
	} else {
		pageSpecificData["User"] = nil
		pageSpecificData["IsLoggedIn"] = false
		pageSpecificData["UserID"] = ""
		pageSpecificData["IsAdminPage"] = false
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

func SplitRememberToken(token string) (selector string, verifier string, err error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid remember token format")
	}
	return parts[0], parts[1], nil
}

func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func GenerateOrderCode() string {

	return fmt.Sprintf("INV-%s-%s", time.Now().Format("20060102"), uuid.New().String()[:8])
}

func GetQuery(r *http.Request) url.Values {
	return r.URL.Query()
}

func PopulateBaseData(pageData *other.BasePageData, baseDataMap map[string]interface{}) {
	if title, ok := baseDataMap["Title"].(string); ok {
		pageData.Title = title
	}
	if isLoggedIn, ok := baseDataMap["IsLoggedIn"].(bool); ok {
		pageData.IsLoggedIn = isLoggedIn
	}
	if user, ok := baseDataMap["User"].(*other.UserForTemplate); ok {
		pageData.User = user
	}
	if userID, ok := baseDataMap["UserID"].(string); ok {
		pageData.UserID = userID
	}
	if cartCount, ok := baseDataMap["CartCount"].(int); ok {
		pageData.CartCount = cartCount
	}
	if csrfToken, ok := baseDataMap["CSRFToken"].(string); ok {
		pageData.CSRFToken = csrfToken
	}
	if message, ok := baseDataMap["Message"].(string); ok {
		pageData.Message = message
	}
	if messageStatus, ok := baseDataMap["MessageStatus"].(string); ok {
		pageData.MessageStatus = messageStatus
	}
	if query, ok := baseDataMap["Query"].(url.Values); ok {
		pageData.Query = query
	}
	if breadcrumbs, ok := baseDataMap["Breadcrumbs"].([]breadcrumb.Breadcrumb); ok {
		pageData.Breadcrumbs = breadcrumbs
	}
	if isAuthPage, ok := baseDataMap["IsAuthPage"].(bool); ok {
		pageData.IsAuthPage = isAuthPage
	}
	if isAdminPage, ok := baseDataMap["IsAdminPage"].(bool); ok {
		pageData.IsAdminPage = isAdminPage
	}
}

func GetUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(ContextKeyUserID).(string)
	if !ok {
		return ""
	}
	return userID
}

func SetUserIDToContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

func OrderStatusText(status int) string {
	switch status {
	case models.OrderStatusPending:
		return "Menunggu Pembayaran"
	case models.OrderStatusProcessing:
		return "Sedang Diproses"
	case models.OrderStatusShipped:
		return "Dalam Pengiriman"
	case models.OrderStatusCompleted:
		return "Selesai"
	case models.OrderStatusCancelled:
		return "Dibatalkan"
	case models.OrderStatusRefunded:
		return "Pengembalian Dana"
	case models.OrderStatusFailed:
		return "Gagal"
	default:
		return "Status Tidak Diketahui"
	}
}

func PaymentStatusText(status string) string {
	switch status {
	case "Paid":
		return "Lunas"
	case "Pending":
		return "Menunggu Pembayaran"
	case "Failed":
		return "Gagal"
	case "Cancelled":
		return "Dibatalkan"
	case "Refunded":
		return "Dikembalikan"
	case "settlement":
		return "Lunas"
	case "capture":
		return "Lunas"
	case "deny":
		return "Ditolak"
	case "expire":
		return "Kadaluarsa"
	case "challenge":
		return "Challenge"
	default:
		return "Tidak Diketahui"
	}
}

type AdminProductFormPageData struct {
	other.BasePageData
	ProductData models.Product
	Categories  []models.Category
	IsEdit      bool
	FormAction  string
	Errors      map[string]string
}

type AdminOrderPageData struct {
	other.BasePageData
	Orders []models.Order

	OrderStatusOptions map[int]string
}
