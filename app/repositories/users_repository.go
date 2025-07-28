package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepositoryImpl interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateRememberToken(ctx context.Context, userID string, selector string, verifierRaw string) error
	FindByRememberToken(ctx context.Context, tokenFromCookie string) (*models.User, error)
	FindByPhone(ctx context.Context, phone string) (*models.User, error)
	GetUserByIDWithAddresses(ctx context.Context, id string) (*models.User, error)

	SavePasswordResetToken(ctx context.Context, userID string, token *string, expiresAt *time.Time) error
	FindByPasswordResetToken(ctx context.Context, token string) (*models.User, error)
	ClearPasswordResetToken(ctx context.Context, userID string) error
	UpdatePassword(ctx context.Context, userID string, newPasswordHash string) error
	FindBySelector(ctx context.Context, selector string) (*models.User, error)

	GetAllUsers(ctx context.Context) ([]models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id string) error
	GetUserCount(ctx context.Context) (int64, error)
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

	user.RememberTokenSelector = nil
	user.RememberTokenHash = ""

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

func (r *userRepository) FindByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User

	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
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

	updates := map[string]interface{}{
		"remember_token_hash": string(verifierRaw),
		"updated_at":          time.Now(),
	}
	if selector == "" {
		updates["remember_token_selector"] = nil
	} else {
		updates["remember_token_selector"] = &selector
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
	verifierRaw := parts[1]

	var user models.User

	err := r.db.WithContext(ctx).Where("remember_token_selector = ?", selector).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	if user.RememberTokenHash == "" || bcrypt.CompareHashAndPassword([]byte(user.RememberTokenHash), []byte(verifierRaw)) != nil {

		return nil, nil
	}

	return &user, nil
}

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

func (r *userRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Find(&users).Error; err != nil {
		log.Printf("UserRepository.GetAllUsers: Error getting all users: %v", err)
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	return users, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	updates := map[string]interface{}{
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
		"phone":      user.Phone,
		"updated_at": user.UpdatedAt,
	}
	if user.Password != "" {
		updates["password"] = user.Password
	}

	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", user.ID).
		Updates(updates).Error; err != nil {
		log.Printf("UserRepository.UpdateUser: Error updating user %s: %v", user.ID, err)
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *userRepository) DeleteUser(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error; err != nil {
		log.Printf("UserRepository.DeleteUser: Error deleting user %s: %v", id, err)
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (r *userRepository) FindBySelector(ctx context.Context, selector string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Where("remember_token_selector = ?", selector).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByIDWithAddresses(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Preload("Address").First(&user, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserCount(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error; err != nil {
		log.Printf("UserRepository.GetUserCount: Failed to count users: %v", err)
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}
