package handlers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/helpers"
	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/services"
	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
	"github.com/Rakhulsr/go-ecommerce/app/utils/sessions"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/unrolled/render"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	render       *render.Render
	userRepo     repositories.UserRepositoryImpl
	cartRepo     repositories.CartRepositoryImpl
	sessionStore sessions.SessionStore
	mailer       *services.Mailer
	validator    *validator.Validate
}

func NewAuthHandler(r *render.Render, userRepo repositories.UserRepositoryImpl, cartRepo repositories.CartRepositoryImpl, sessionStore sessions.SessionStore, mailer *services.Mailer, validator *validator.Validate) *AuthHandler {
	return &AuthHandler{
		render:       r,
		userRepo:     userRepo,
		cartRepo:     cartRepo,
		sessionStore: sessionStore,
		mailer:       mailer,
		validator:    validator,
	}
}

type UserForm struct {
	ID        string `form:"id"`
	FirstName string `form:"first_name" validate:"required,min=2,max=100"`
	LastName  string `form:"last_name" validate:"required,min=2,max=100"`
	Email     string `form:"email" validate:"required,email"`
	Password  string `form:"password" validate:"omitempty,min=6"`
}

func (h *AuthHandler) LoginGetHandler(w http.ResponseWriter, r *http.Request) {
	if userID, ok := r.Context().Value(helpers.ContextKeyUserID).(string); ok && userID != "" {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	breadcrumbs := []breadcrumb.Breadcrumb{
		{Name: "Home", URL: "/"},
		{Name: "Login", URL: "/login"},
	}

	pageSpecificData := map[string]interface{}{
		"title":         "Login",
		"Breadcrumbs":   breadcrumbs,
		"MessageStatus": r.URL.Query().Get("status"),
		"Message":       r.URL.Query().Get("message"),
		"IsAuthPage":    true,
	}

	data := helpers.GetBaseData(r, pageSpecificData)
	_ = h.render.HTML(w, http.StatusOK, "auth/login", data)
}

func (h *AuthHandler) LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("LoginPostHandler: Error parsing form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/login?status=error&message=%s", url.QueryEscape("Terjadi kesalahan saat memproses data.")), http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	rememberMe := r.FormValue("remember_me") == "on"

	user, err := h.userRepo.FindByEmail(r.Context(), email)
	if err != nil {
		log.Printf("LoginPostHandler: Error getting user by email '%s': %v", email, err)
		http.Redirect(w, r, fmt.Sprintf("/login?status=error&message=%s", url.QueryEscape("Terjadi kesalahan server.")), http.StatusSeeOther)
		return
	}
	if user == nil {
		log.Printf("LoginPostHandler: User not found for email: %s", email)
		http.Redirect(w, r, fmt.Sprintf("/login?status=error&message=%s", url.QueryEscape("Email atau password salah.")), http.StatusSeeOther)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Printf("LoginPostHandler: Password mismatch for email: %s", email)
		http.Redirect(w, r, fmt.Sprintf("/login?status=error&message=%s", url.QueryEscape("Email atau password salah.")), http.StatusSeeOther)
		return
	}

	err = h.sessionStore.SetUserID(w, r, user.ID)
	if err != nil {
		log.Printf("LoginPostHandler: Error setting user session: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/login?status=error&message=%s", url.QueryEscape("Gagal membuat sesi login.")), http.StatusSeeOther)
		return
	}

	if rememberMe {

		selector, verifierRaw, _, genErr := helpers.GenerateRememberTokenParts()
		if genErr != nil {
			log.Printf("LoginPostHandler: Failed to generate remember token parts: %v", genErr)
		} else {

			hashedVerifier, hashErr := bcrypt.GenerateFromPassword([]byte(verifierRaw), bcrypt.DefaultCost)
			if hashErr != nil {
				log.Printf("LoginPostHandler: Failed to hash verifier: %v", hashErr)
			} else {

				err = h.userRepo.UpdateRememberToken(r.Context(), user.ID, selector, string(hashedVerifier))
				if err != nil {
					log.Printf("LoginPostHandler: Failed to update remember token for user %s in DB: %v", user.ID, err)
				} else {
					log.Printf("LoginPostHandler: Remember token updated in DB for user %s. Middleware will handle cookie.", user.Email)
				}
			}
		}
	} else {

		err = h.userRepo.UpdateRememberToken(r.Context(), user.ID, "", "")
		if err != nil {
			log.Printf("LoginPostHandler: Failed to clear remember token for user %s in DB: %v", user.ID, err)
		}
	}

	userCart, err := h.cartRepo.GetOrCreateCartByUserID(r.Context(), user.ID)
	if err != nil {
		log.Printf("LoginPostHandler: Failed to get or create cart for user %s: %v", user.ID, err)

	} else {

		if err := h.sessionStore.SetCartID(w, r, userCart.ID); err != nil {
			log.Printf("LoginPostHandler: Failed to set cart ID in session for user %s: %v", user.ID, err)
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/?status=success&message=%s", url.QueryEscape(fmt.Sprintf("Selamat datang, %s!", user.FirstName))), http.StatusSeeOther)
}

func (h *AuthHandler) RegisterGetHandler(w http.ResponseWriter, r *http.Request) {
	if userID, ok := r.Context().Value(helpers.ContextKeyUserID).(string); ok && userID != "" {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	breadcrumbs := []breadcrumb.Breadcrumb{
		{Name: "Home", URL: "/"},
		{Name: "Daftar", URL: "/register"},
	}

	pageSpecificData := map[string]interface{}{
		"title":         "Daftar Akun Baru",
		"Breadcrumbs":   breadcrumbs,
		"MessageStatus": r.URL.Query().Get("status"),
		"Message":       r.URL.Query().Get("message"),
		"IsAuthPage":    true,
	}

	data := helpers.GetBaseData(r, pageSpecificData)
	_ = h.render.HTML(w, http.StatusOK, "auth/register", data)
}

func (h *AuthHandler) RegisterPostHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("RegisterPostHandler: Error parsing form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/register?status=error&message=%s", url.QueryEscape("Terjadi kesalahan saat memproses data.")), http.StatusSeeOther)
		return
	}

	firstName := r.FormValue("firstname")
	lastName := r.FormValue("lastname")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if firstName == "" || lastName == "" || email == "" || password == "" || confirmPassword == "" {
		http.Redirect(w, r, fmt.Sprintf("/register?status=error&message=%s", url.QueryEscape("Semua kolom harus diisi.")), http.StatusSeeOther)
		return
	}

	if password != confirmPassword {
		http.Redirect(w, r, fmt.Sprintf("/register?status=error&message=%s", url.QueryEscape("Konfirmasi password tidak cocok.")), http.StatusSeeOther)
		return
	}

	user := &models.User{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Password:  password,
		Role:      models.RoleCustomer,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	existingUser, err := h.userRepo.FindByEmail(r.Context(), email)
	if err != nil {
		log.Printf("RegisterPostHandler: Error checking existing user: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/register?status=error&message=%s", url.QueryEscape("Terjadi kesalahan server saat mendaftar.")), http.StatusSeeOther)
		return
	}
	if existingUser != nil {
		http.Redirect(w, r, fmt.Sprintf("/register?status=error&message=%s", url.QueryEscape("Email sudah terdaftar. Silakan login atau gunakan email lain.")), http.StatusSeeOther)
		return
	}

	err = h.userRepo.Create(r.Context(), user)
	if err != nil {
		log.Printf("RegisterPostHandler: Error creating user: %v", err)

		if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "remember_token_selector") {
			log.Printf("RegisterPostHandler: Duplicate remember_token_selector error. This should not happen with NULLable unique index: %v", err)
			http.Redirect(w, r, fmt.Sprintf("/register?status=error&message=%s", url.QueryEscape("Terjadi kesalahan internal. Silakan coba lagi.")), http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/register?status=error&message=%s", url.QueryEscape("Gagal mendaftar. Silakan coba lagi.")), http.StatusSeeOther)
		return
	}

	log.Printf("RegisterPostHandler: User %s (%s) registered successfully.", user.Email, user.ID)

	newCart, err := h.cartRepo.CreateCartForUser(r.Context(), user.ID)
	if err != nil {
		log.Printf("RegisterPostHandler: Failed to create cart for new user %s: %v", user.ID, err)

	} else {
		log.Printf("RegisterPostHandler: Cart %s created for new user %s.", newCart.ID, user.ID)
	}

	http.Redirect(w, r, fmt.Sprintf("/login?status=success&message=%s", url.QueryEscape("Akun Anda berhasil dibuat! Silakan login.")), http.StatusSeeOther)
}

func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(helpers.ContextKeyUserID).(string)
	if ok && userID != "" {
		err := h.userRepo.UpdateRememberToken(r.Context(), userID, "", "")
		if err != nil {
			log.Printf("LogoutHandler: Failed to clear remember token in DB for user %s: %v", userID, err)
		}
	}

	err := h.sessionStore.ClearUserID(w, r)
	if err != nil {
		log.Printf("LogoutHandler: Error clearing user session: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/?status=error&message=%s", url.QueryEscape("Gagal logout.")), http.StatusSeeOther)
		return
	}

	helpers.ClearCookie(w, "remember_token")

	http.Redirect(w, r, "/?status=success&message=Anda%20telah%20berhasil%20logout.", http.StatusSeeOther)
}

func (h *AuthHandler) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		log.Printf("ProfileHandler: UserID not found in context for /profile. Redirecting to login.")
		http.Redirect(w, r, fmt.Sprintf("/login?status=warning&message=%s", url.QueryEscape("Anda harus login untuk mengakses halaman ini.")), http.StatusSeeOther)
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		log.Printf("ProfileHandler: Error getting user %s from DB: %v", userID, err)
		http.Redirect(w, r, fmt.Sprintf("/?status=error&message=%s", url.QueryEscape("Gagal mengambil data profil.")), http.StatusSeeOther)
		return
	}
	if user == nil {
		log.Printf("ProfileHandler: User ID %s not found in DB despite being logged in. Clearing session.", userID)
		h.sessionStore.ClearUserID(w, r)
		helpers.ClearCookie(w, "remember_token")
		http.Redirect(w, r, fmt.Sprintf("/login?status=warning&message=%s", url.QueryEscape("Sesi Anda tidak valid. Silakan login kembali.")), http.StatusSeeOther)
		return
	}

	breadcrumbs := []breadcrumb.Breadcrumb{
		{Name: "Home", URL: "/"},
		{Name: "Profil Saya", URL: "/profile"},
	}

	pageSpecificData := map[string]interface{}{
		"title":         "Profil Saya",
		"Breadcrumbs":   breadcrumbs,
		"User":          user,
		"MessageStatus": r.URL.Query().Get("status"),
		"Message":       r.URL.Query().Get("message"),
		"IsAuthPage":    false,
	}

	data := helpers.GetBaseData(r, pageSpecificData)
	_ = h.render.HTML(w, http.StatusOK, "auth/profile", data)
}

func (h *AuthHandler) ForgotPasswordGetHandler(w http.ResponseWriter, r *http.Request) {

	if userID, ok := r.Context().Value(helpers.ContextKeyUserID).(string); ok && userID != "" {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	breadcrumbs := []breadcrumb.Breadcrumb{
		{Name: "Home", URL: "/"},
		{Name: "Lupa Kata Sandi", URL: "/forgot-password"},
	}

	pageSpecificData := map[string]interface{}{
		"title":         "Lupa Kata Sandi",
		"Breadcrumbs":   breadcrumbs,
		"MessageStatus": r.URL.Query().Get("status"),
		"Message":       r.URL.Query().Get("message"),
		"IsAuthPage":    true,
	}

	data := helpers.GetBaseData(r, pageSpecificData)
	_ = h.render.HTML(w, http.StatusOK, "auth/forgot_password", data)
}

func (h *AuthHandler) ForgotPasswordPostHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ForgotPasswordPostHandler: Error parsing form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/forgot-password?status=error&message=%s", url.QueryEscape("Terjadi kesalahan saat memproses data.")), http.StatusSeeOther)
		return
	}

	emailAddress := r.FormValue("email")
	if emailAddress == "" {
		http.Redirect(w, r, fmt.Sprintf("/forgot-password?status=error&message=%s", url.QueryEscape("Email harus diisi.")), http.StatusSeeOther)
		return
	}

	user, err := h.userRepo.FindByEmail(r.Context(), emailAddress)
	if err != nil {
		log.Printf("ForgotPasswordPostHandler: Error finding user by email '%s': %v", emailAddress, err)

		http.Redirect(w, r, fmt.Sprintf("/forgot-password?status=success&message=%s", url.QueryEscape("Jika email Anda terdaftar, kode verifikasi telah dikirimkan.")), http.StatusSeeOther)
		return
	}
	if user == nil {
		log.Printf("ForgotPasswordPostHandler: User with email '%s' not found.", emailAddress)

		http.Redirect(w, r, fmt.Sprintf("/forgot-password?status=success&message=%s", url.QueryEscape("Jika email Anda terdaftar, kode verifikasi telah dikirimkan.")), http.StatusSeeOther)
		return
	}

	otpCodeInt := rand.Intn(900000) + 100000
	otpCode := strconv.Itoa(otpCodeInt)
	expiryDuration := 5 * time.Minute
	expiresAt := time.Now().Add(expiryDuration)

	otpPtr := &otpCode
	expiresAtPtr := &expiresAt

	err = h.userRepo.SavePasswordResetToken(r.Context(), user.ID, otpPtr, expiresAtPtr)
	if err != nil {
		log.Printf("ForgotPasswordPostHandler: Failed to save OTP for user %s: %v", user.ID, err)
		http.Redirect(w, r, fmt.Sprintf("/forgot-password?status=error&message=%s", url.QueryEscape("Gagal memproses permintaan. Silakan coba lagi.")), http.StatusSeeOther)
		return
	}

	subject := "Kode Verifikasi Reset Kata Sandi Anda"
	htmlBody := services.BuildOTPEmailBody(otpCode, int(expiryDuration.Minutes()))

	err = h.mailer.SendHTMLEmail(user.Email, subject, htmlBody)
	if err != nil {
		log.Printf("ForgotPasswordPostHandler: Gagal mengirim email OTP ke %s: %v", user.Email, err)
		http.Redirect(w, r, fmt.Sprintf("/forgot-password?status=success&message=%s", url.QueryEscape("Jika email Anda terdaftar, kode verifikasi telah dikirimkan. Silakan cek email Anda (mungkin di folder spam).")), http.StatusSeeOther)
		return
	}

	log.Printf("ForgotPasswordPostHandler: Kode OTP berhasil dikirim ke %s. OTP: %s", user.Email, otpCode)

	http.Redirect(w, r, fmt.Sprintf("/verify-otp?email=%s&status=success&message=%s", url.QueryEscape(emailAddress), url.QueryEscape("Kode verifikasi telah dikirimkan. Silakan masukkan di bawah.")), http.StatusSeeOther)
}

func (h *AuthHandler) ResetPasswordGetHandler(w http.ResponseWriter, r *http.Request) {

	if userID, ok := r.Context().Value(helpers.ContextKeyUserID).(string); ok && userID != "" {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, r, fmt.Sprintf("/login?status=error&message=%s", url.QueryEscape("Tautan reset kata sandi tidak valid.")), http.StatusSeeOther)
		return
	}

	user, err := h.userRepo.FindByPasswordResetToken(r.Context(), token)
	if err != nil {
		log.Printf("ResetPasswordGetHandler: Error finding user by reset token '%s': %v", token, err)
		http.Redirect(w, r, fmt.Sprintf("/login?status=error&message=%s", url.QueryEscape("Terjadi kesalahan server saat memverifikasi tautan.")), http.StatusSeeOther)
		return
	}
	if user == nil {
		log.Printf("ResetPasswordGetHandler: Invalid or expired reset token: %s", token)
		http.Redirect(w, r, fmt.Sprintf("/login?status=error&message=%s", url.QueryEscape("Tautan reset kata sandi tidak valid atau sudah kedaluwarsa.")), http.StatusSeeOther)
		return
	}

	breadcrumbs := []breadcrumb.Breadcrumb{
		{Name: "Home", URL: "/"},
		{Name: "Reset Kata Sandi", URL: "/reset-password"},
	}

	pageSpecificData := map[string]interface{}{
		"title":         "Reset Kata Sandi",
		"Breadcrumbs":   breadcrumbs,
		"Token":         token,
		"MessageStatus": r.URL.Query().Get("status"),
		"Message":       r.URL.Query().Get("message"),
		"IsAuthPage":    true,
	}

	data := helpers.GetBaseData(r, pageSpecificData)
	_ = h.render.HTML(w, http.StatusOK, "auth/reset_password", data)
}

func (h *AuthHandler) ResetPasswordPostHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ResetPasswordPostHandler: Error parsing form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/login?status=error&message=%s", url.QueryEscape("Terjadi kesalahan saat memproses data.")), http.StatusSeeOther)
		return
	}

	token := r.FormValue("token")
	newPassword := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if token == "" || newPassword == "" || confirmPassword == "" {
		http.Redirect(w, r, fmt.Sprintf("/reset-password?token=%s&status=error&message=%s", token, url.QueryEscape("Semua kolom harus diisi.")), http.StatusSeeOther)
		return
	}

	if newPassword != confirmPassword {
		http.Redirect(w, r, fmt.Sprintf("/reset-password?token=%s&status=error&message=%s", token, url.QueryEscape("Konfirmasi kata sandi tidak cocok.")), http.StatusSeeOther)
		return
	}

	user, err := h.userRepo.FindByPasswordResetToken(r.Context(), token)
	if err != nil {
		log.Printf("ResetPasswordPostHandler: Error finding user by reset token '%s': %v", token, err)
		http.Redirect(w, r, fmt.Sprintf("/login?status=error&message=%s", url.QueryEscape("Terjadi kesalahan server saat memverifikasi tautan.")), http.StatusSeeOther)
		return
	}
	if user == nil {
		log.Printf("ResetPasswordPostHandler: Invalid or expired reset token during post: %s", token)
		http.Redirect(w, r, fmt.Sprintf("/login?status=error&message=%s", url.QueryEscape("Tautan reset kata sandi tidak valid atau sudah kedaluwarsa.")), http.StatusSeeOther)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ResetPasswordPostHandler: Failed to hash new password for user %s: %v", user.ID, err)
		http.Redirect(w, r, fmt.Sprintf("/reset-password?token=%s&status=error&message=%s", token, url.QueryEscape("Gagal mengatur ulang kata sandi. Silakan coba lagi.")), http.StatusSeeOther)
		return
	}

	err = h.userRepo.UpdatePassword(r.Context(), user.ID, string(hashedPassword))
	if err != nil {
		log.Printf("ResetPasswordPostHandler: Failed to update password for user %s: %v", user.ID, err)
		http.Redirect(w, r, fmt.Sprintf("/reset-password?token=%s&status=error&message=%s", token, url.QueryEscape("Gagal mengatur ulang kata sandi. Silakan coba lagi.")), http.StatusSeeOther)
		return
	}

	err = h.userRepo.ClearPasswordResetToken(r.Context(), user.ID)
	if err != nil {
		log.Printf("ResetPasswordPostHandler: Failed to clear reset password token for user %s: %v", user.ID, err)

	}

	http.Redirect(w, r, fmt.Sprintf("/login?status=success&message=%s", url.QueryEscape("Kata sandi Anda berhasil diatur ulang. Silakan login.")), http.StatusSeeOther)
}

func (h *AuthHandler) VerifyOTPGetHandler(w http.ResponseWriter, r *http.Request) {

	emailFromParam := r.URL.Query().Get("email")

	breadcrumbs := []breadcrumb.Breadcrumb{
		{Name: "Home", URL: "/"},
		{Name: "Verifikasi OTP", URL: "/verify-otp"},
	}

	pageSpecificData := map[string]interface{}{
		"title":         "Verifikasi Kode OTP",
		"Breadcrumbs":   breadcrumbs,
		"Email":         emailFromParam,
		"MessageStatus": r.URL.Query().Get("status"),
		"Message":       r.URL.Query().Get("message"),
		"IsAuthPage":    true,
	}

	data := helpers.GetBaseData(r, pageSpecificData)

	_ = h.render.HTML(w, http.StatusOK, "auth/verify_otp", data)
}

func (h *AuthHandler) VerifyOTPPostHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("VerifyOTPPostHandler: Error parsing form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/verify-otp?status=error&message=%s", url.QueryEscape("Terjadi kesalahan saat memproses data.")), http.StatusSeeOther)
		return
	}

	emailAddress := r.FormValue("email")
	enteredOTP := r.FormValue("otp_code")

	if emailAddress == "" || enteredOTP == "" {
		http.Redirect(w, r, fmt.Sprintf("/verify-otp?email=%s&status=error&message=%s", url.QueryEscape(emailAddress), url.QueryEscape("Email dan Kode OTP harus diisi.")), http.StatusSeeOther)
		return
	}

	user, err := h.userRepo.FindByEmail(r.Context(), emailAddress)
	if err != nil {
		log.Printf("VerifyOTPPostHandler: Error finding user by email '%s': %v", emailAddress, err)
		http.Redirect(w, r, fmt.Sprintf("/verify-otp?email=%s&status=error&message=%s", url.QueryEscape(emailAddress), url.QueryEscape("Terjadi kesalahan. Silakan coba lagi.")), http.StatusSeeOther)
		return
	}
	if user == nil {
		log.Printf("VerifyOTPPostHandler: User with email '%s' not found during OTP verification.", emailAddress)
		http.Redirect(w, r, fmt.Sprintf("/verify-otp?email=%s&status=error&message=%s", url.QueryEscape(emailAddress), url.QueryEscape("Email tidak terdaftar atau tidak valid.")), http.StatusSeeOther)
		return
	}

	if user.PasswordResetToken == nil || *user.PasswordResetToken != enteredOTP {
		log.Printf("VerifyOTPPostHandler: Invalid OTP entered for user %s. Expected: %v, Got: %s", user.ID, user.PasswordResetToken, enteredOTP)
		http.Redirect(w, r, fmt.Sprintf("/verify-otp?email=%s&status=error&message=%s", url.QueryEscape(emailAddress), url.QueryEscape("Kode OTP tidak valid.")), http.StatusSeeOther)
		return
	}

	if user.PasswordResetExpires == nil || user.PasswordResetExpires.Before(time.Now()) {
		log.Printf("VerifyOTPPostHandler: Expired OTP for user %s.", user.ID)
		h.userRepo.ClearPasswordResetToken(r.Context(), user.ID)
		http.Redirect(w, r, fmt.Sprintf("/verify-otp?email=%s&status=error&message=%s", url.QueryEscape(emailAddress), url.QueryEscape("Kode OTP sudah kedaluwarsa. Silakan minta kode baru.")), http.StatusSeeOther)
		return
	}

	resetSessionToken := uuid.New().String()

	resetSessionTokenPtr := &resetSessionToken
	resetSessionExpires := time.Now().Add(15 * time.Minute)
	resetSessionExpiresPtr := &resetSessionExpires

	err = h.userRepo.SavePasswordResetToken(r.Context(), user.ID, resetSessionTokenPtr, resetSessionExpiresPtr)
	if err != nil {
		log.Printf("VerifyOTPPostHandler: Failed to save reset session token for user %s: %v", user.ID, err)
		http.Redirect(w, r, fmt.Sprintf("/verify-otp?email=%s&status=error&message=%s", url.QueryEscape(emailAddress), url.QueryEscape("Gagal memproses permintaan. Silakan coba lagi.")), http.StatusSeeOther)
		return
	}

	log.Printf("VerifyOTPPostHandler: OTP verified for user %s. Redirecting to reset password with session token: %s", user.ID, resetSessionToken)

	http.Redirect(w, r, fmt.Sprintf("/reset-password?token=%s", resetSessionToken), http.StatusSeeOther)
}

func (h *AuthHandler) UpdateProfilePage(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		http.Redirect(w, r, "/login?status=error&message=Silakan login terlebih dahulu", http.StatusSeeOther)
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil || user == nil {
		http.Redirect(w, r, "/?status=error&message=Gagal memuat profil", http.StatusSeeOther)
		return
	}

	formData := UserForm{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}

	// Gunakan map untuk data khusus halaman
	pageData := map[string]interface{}{
		"Title":      "Edit Profil",
		"IsAuthPage": true,
		"Breadcrumbs": []breadcrumb.Breadcrumb{
			{Name: "Beranda", URL: "/"},
			{Name: "Profil", URL: "/profile"},
			{Name: "Edit", URL: ""},
		},
		"UserForm": &formData,
		"Errors":   map[string]string{},
	}

	// Tambahkan base data (seperti CartCount, UserID, dll)
	data := helpers.GetBaseData(r, pageData)

	h.render.HTML(w, http.StatusOK, "auth/profile/edit", data)
}

func (h *AuthHandler) UpdateProfilePost(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(helpers.ContextKeyUserID).(string)
	if !ok || userID == "" {
		http.Redirect(w, r, "/login?status=error&message=Unauthorized", http.StatusSeeOther)
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil || user == nil {
		log.Printf("UpdateProfile: Gagal ambil data user dari database: %v", err)
		http.Redirect(w, r, "/profile?status=error&message=Gagal memuat data profil", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("UpdateProfile: Gagal parsing form: %v", err)
		http.Redirect(w, r, "/profile?status=error&message=Gagal memproses form", http.StatusSeeOther)
		return
	}

	var form UserForm
	form.FirstName = r.PostFormValue("first_name")
	form.LastName = r.PostFormValue("last_name")
	form.Email = r.PostFormValue("email")
	form.Password = r.PostFormValue("password")
	form.ID = user.ID

	if err := h.validator.Struct(&form); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		formattedErrors := helpers.FormatValidationErrors(validationErrors)

		data := map[string]interface{}{
			"title":      "Edit Profil",
			"UserData":   &form,
			"Errors":     formattedErrors,
			"IsAuthPage": false,
		}
		datas := helpers.GetBaseData(r, data)
		h.render.HTML(w, http.StatusOK, "auth/profile", datas)
		return
	}

	if user.Email != form.Email {
		existingUser, _ := h.userRepo.FindByEmail(r.Context(), form.Email)
		if existingUser != nil && existingUser.ID != user.ID {
			http.Redirect(w, r, "/profile?status=error&message=Email sudah digunakan oleh pengguna lain", http.StatusSeeOther)
			return
		}
	}

	user.FirstName = form.FirstName
	user.LastName = form.LastName
	user.Email = form.Email

	if form.Password != "" {
		if len(form.Password) < 6 {
			http.Redirect(w, r, "/profile?status=error&message=Password minimal 6 karakter", http.StatusSeeOther)
			return
		}
		user.Password = helpers.HashPassword(form.Password)
	}

	err = h.userRepo.UpdateUser(r.Context(), user)
	if err != nil {
		log.Printf("UpdateProfile: Gagal update user: %v", err)
		http.Redirect(w, r, "/profile?status=error&message=Gagal memperbarui profil", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/profile?status=success&message=Profil berhasil diperbarui", http.StatusSeeOther)
}
