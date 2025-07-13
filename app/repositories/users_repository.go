package repositories

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time" // Import time

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRepositoryImpl adalah interface untuk operasi user repository.
type UserRepositoryImpl interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateRememberToken(ctx context.Context, userID string, selector string, verifierRaw string) error
	FindByRememberToken(ctx context.Context, tokenFromCookie string) (*models.User, error)
	// --- Tambahan untuk Forgot Password ---
	SavePasswordResetToken(ctx context.Context, userID string, token *string, expiresAt *time.Time) error
	FindByPasswordResetToken(ctx context.Context, token string) (*models.User, error)
	ClearPasswordResetToken(ctx context.Context, userID string) error
	UpdatePassword(ctx context.Context, userID string, newPasswordHash string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepositoryImpl {
	return &userRepository{db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password for user %s: %v", user.Email, err)
		return err
	}
	user.Password = string(hashPass)

	if user.Role == "" {
		user.Role = models.RoleCustomer
	}

	// Pastikan ini disetel ke nil untuk user baru
	user.RememberTokenSelector = nil
	user.RememberTokenHash = ""
	// Pastikan juga token reset password disetel ke nil untuk user baru
	user.PasswordResetToken = nil
	user.PasswordResetExpires = nil

	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) UpdateRememberToken(ctx context.Context, userID string, selector string, verifierRaw string) error {
	// Menggunakan Updates dengan map untuk mengontrol nilai NULL
	updates := map[string]interface{}{
		"remember_token_hash": string(verifierRaw), // verifierRaw sudah di-hash di handler atau di tempat lain jika tidak, hash di sini
		"updated_at":          time.Now(),
	}
	if selector == "" {
		updates["remember_token_selector"] = nil
	} else {
		updates["remember_token_selector"] = &selector // Simpan sebagai pointer ke string
	}

	result := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update remember token for user %s: %w", userID, result.Error)
	}
	return nil
}

func (r *userRepository) FindByRememberToken(ctx context.Context, tokenFromCookie string) (*models.User, error) {
	parts := strings.SplitN(tokenFromCookie, ".", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid remember token format")
	}

	selector := parts[0]
	verifierRaw := parts[1] // Ini adalah verifier mentah dari cookie

	var user models.User

	// Cari user berdasarkan selector
	err := r.db.WithContext(ctx).Where("remember_token_selector = ?", selector).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Selector tidak ditemukan
		}
		return nil, err
	}

	// Bandingkan verifier mentah dari cookie dengan hash yang disimpan
	// Pastikan user.RememberTokenHash tidak kosong sebelum membandingkan
	if user.RememberTokenHash == "" || bcrypt.CompareHashAndPassword([]byte(user.RememberTokenHash), []byte(verifierRaw)) != nil {
		// Jika hash tidak cocok atau hash kosong, token tidak valid
		return nil, nil
	}

	return &user, nil
}

// --- Metode Baru untuk Forgot Password ---

// SavePasswordResetToken menyimpan token reset password ke database
func (r *userRepository) SavePasswordResetToken(ctx context.Context, userID string, token *string, expiresAt *time.Time) error {
	updates := map[string]interface{}{
		"password_reset_token":   token,
		"password_reset_expires": expiresAt,
		"updated_at":             time.Now(),
	}
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to save password reset token for user %s: %w", userID, result.Error)
	}
	return nil
}

func (r *userRepository) FindByPasswordResetToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	// Cari token yang cocok dan belum kedaluwarsa
	result := r.db.WithContext(ctx).Where("password_reset_token = ? AND password_reset_expires > ?", token, time.Now()).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by password reset token: %w", result.Error)
	}
	return &user, nil
}

func (r *userRepository) ClearPasswordResetToken(ctx context.Context, userID string) error {
	updates := map[string]interface{}{
		"password_reset_token":   nil,
		"password_reset_expires": nil,
		"updated_at":             time.Now(),
	}
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to clear password reset token for user %s: %w", userID, result.Error)
	}
	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID string, newPasswordHash string) error {
	updates := map[string]interface{}{
		"password":   newPasswordHash,
		"updated_at": time.Now(),
	}
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update password for user %s: %w", userID, result.Error)
	}
	return nil
}
