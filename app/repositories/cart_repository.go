package repositories

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/utils/calc"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CartRepositoryImpl interface {
	GetCartWithItems(ctx context.Context, cartID string) (*models.Cart, error)
	GetByID(ctx context.Context, id string) (*models.Cart, error)
	CreateCart(ctx context.Context, cart *models.Cart) error
	UpdateCartSummary(ctx context.Context, cartID string) error
	GetCartItemCount(ctx context.Context, cartID string) (int, error)
	UpdateCart(ctx context.Context, cart *models.Cart) error
	GetOrCreateCartByUserID(ctx context.Context, userID string) (*models.Cart, error)
	DeleteCart(ctx context.Context, cartID string) error
	CreateCartForUser(ctx context.Context, userID string) (*models.Cart, error)
	GetAllCarts(ctx context.Context) ([]models.Cart, error)
}

type cartRepository struct {
	cartItemRepo CartItemRepositoryImpl
	db           *gorm.DB
}

func NewCartRepository(db *gorm.DB, cItemRepo CartItemRepositoryImpl) CartRepositoryImpl {
	return &cartRepository{db: db, cartItemRepo: cItemRepo}
}

func (r *cartRepository) GetCartWithItems(ctx context.Context, cartID string) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.WithContext(ctx).
		Preload("CartItems.Product.ProductImages").
		Preload("CartItems.Product").
		Preload("CartItems").
		Where("id = ?", cartID).
		First(&cart).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) GetByID(ctx context.Context, id string) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.WithContext(ctx).First(&cart, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) CreateCart(ctx context.Context, cart *models.Cart) error {
	return r.db.WithContext(ctx).Create(cart).Error
}

func (r *cartRepository) UpdateCartSummary(ctx context.Context, cartID string) error {
	var items []models.CartItem

	// Penting: Preload Product untuk mendapatkan diskon dari Product
	if err := r.db.WithContext(ctx).
		Preload("Product").
		Where("cart_id = ?", cartID).
		Find(&items).Error; err != nil {
		return err
	}

	var baseTotal, taxTotal, discountTotal, grandTotal decimal.Decimal
	var totalWeight int = 0

	var taxPercentCart, discountPercentCart decimal.Decimal // Untuk ringkasan cart

	for _, item := range items {
		// Pastikan Product dimuat sebelum mengakses diskon
		if item.Product.ID == "" {
			// Ini seharusnya tidak terjadi jika Preload berhasil, tapi sebagai fallback
			log.Printf("UpdateCartSummary: Product not preloaded for item %s, attempting to fetch.", item.ID)
			var product models.Product
			if err := r.db.WithContext(ctx).First(&product, "id = ?", item.ProductID).Error; err == nil {
				item.Product = product
			} else {
				log.Printf("UpdateCartSummary: Failed to fetch product %s for item %s: %v", item.ProductID, item.ID, err)
				continue // Lewati item ini jika produk tidak ditemukan
			}
		}

		// Hitung ulang BaseTotal, TaxAmount, DiscountAmount, GrandTotal untuk setiap item
		// Berdasarkan harga produk dan diskon produk
		item.BasePrice = item.Product.Price
		item.BaseTotal = item.BasePrice.Mul(decimal.NewFromInt(int64(item.Qty)))

		// Diskon diambil dari produk
		itemDiscountAmount := item.Product.DiscountAmount.Mul(decimal.NewFromInt(int64(item.Qty)))
		itemDiscountPercent := item.Product.DiscountPercent

		item.TaxPercent = calc.GetTaxPercent()                                     // Asumsi pajak global
		item.TaxAmount = calc.CalculateTax(item.BaseTotal.Sub(itemDiscountAmount)) // Pajak setelah diskon

		item.SubTotal = item.BaseTotal.Sub(itemDiscountAmount)
		item.GrandTotal = item.SubTotal.Add(item.TaxAmount)

		// Perbarui item di database jika ada perubahan
		if err := r.db.WithContext(ctx).Save(&item).Error; err != nil {
			log.Printf("UpdateCartSummary: Gagal memperbarui cart item %s: %v", item.ID, err)
		}

		baseTotal = baseTotal.Add(item.BaseTotal)
		taxTotal = taxTotal.Add(item.TaxAmount)
		discountTotal = discountTotal.Add(itemDiscountAmount) // Total diskon dari semua item
		grandTotal = grandTotal.Add(item.GrandTotal)
		if item.Product.ID != "" {
			totalWeight += int(item.Product.Weight.Mul(decimal.NewFromInt(int64(item.Qty))).IntPart())
		}
		// Untuk ringkasan cart, kita bisa mengambil diskon dari item pertama atau rata-rata
		// Untuk kesederhanaan, kita bisa mengambil dari item yang ada atau membiarkannya nol jika tidak ada diskon global
		if taxPercentCart.IsZero() && item.TaxPercent.GreaterThan(decimal.Zero) {
			taxPercentCart = item.TaxPercent
		}
		if discountPercentCart.IsZero() && itemDiscountPercent.GreaterThan(decimal.Zero) {
			discountPercentCart = itemDiscountPercent
		}
	}

	return r.db.WithContext(ctx).
		Model(&models.Cart{}).
		Where("id = ?", cartID).
		Updates(models.Cart{
			BaseTotalPrice:  baseTotal,
			TaxAmount:       taxTotal,
			TaxPercent:      taxPercentCart,
			DiscountAmount:  discountTotal,       // Total diskon dari semua item
			DiscountPercent: discountPercentCart, // Persentase diskon (bisa jadi rata-rata atau dari satu item)
			GrandTotal:      grandTotal,
			TotalWeight:     totalWeight,
			UpdatedAt:       time.Now(),
		}).Error
}
func (r *cartRepository) DeleteCart(ctx context.Context, cartID string) error {

	items, err := r.cartItemRepo.GetByCartID(ctx, cartID)
	if err != nil {
		return fmt.Errorf("failed to get cart items for deletion of cart %s: %w", cartID, err)
	}

	for _, item := range items {
		if err := r.cartItemRepo.Delete(ctx, item.CartID, item.ProductID); err != nil {
			log.Printf("Warning: Failed to delete cart item %s from cart %s during cart deletion: %v", item.ID, cartID, err)

		}
	}

	if err := r.db.WithContext(ctx).Delete(&models.Cart{}, "id = ?", cartID).Error; err != nil {
		return fmt.Errorf("failed to delete cart %s: %w", cartID, err)
	}
	return nil
}

func (r *cartRepository) GetCartItemCount(ctx context.Context, cartID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("cart_items").
		Where("cart_id = ?", cartID).
		Count(&count).Error

	return int(count), err
}

func (r *cartRepository) CreateCartForUser(ctx context.Context, userID string) (*models.Cart, error) {
	newCart := &models.Cart{
		ID:     uuid.New().String(),
		UserID: userID,
	}

	result := r.db.WithContext(ctx).Create(newCart)
	if result.Error != nil {
		log.Printf("CartRepository: Failed to create cart for user %s: %v", userID, result.Error)
		return nil, result.Error
	}
	log.Printf("CartRepository: Successfully created cart with ID: %s for user %s", newCart.ID, userID)
	return newCart, nil
}

func (r *cartRepository) GetOrCreateCartByUserID(ctx context.Context, userID string) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&cart).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {

			newCart := &models.Cart{
				ID:     uuid.New().String(),
				UserID: userID,
			}
			if createErr := r.db.WithContext(ctx).Create(newCart).Error; createErr != nil {
				log.Printf("CartRepository: Failed to create cart for user %s: %v", userID, createErr)
				return nil, createErr
			}
			log.Printf("CartRepository: Created new cart %s for user %s", newCart.ID, userID)
			return newCart, nil
		}
		log.Printf("CartRepository: Error finding cart by user ID %s: %v", userID, err)
		return nil, err
	}
	log.Printf("CartRepository: Found existing cart %s for user %s", cart.ID, userID)
	return &cart, nil
}

func (r *cartRepository) UpdateCart(ctx context.Context, cart *models.Cart) error {
	cart.UpdatedAt = time.Now()

	err := r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(cart).Error
	if err != nil {
		log.Printf("CartRepository.UpdateCart: Error updating cart %s: %v", cart.ID, err)
		return fmt.Errorf("failed to update cart: %w", err)
	}
	return nil
}

func (r *cartRepository) GetAllCarts(ctx context.Context) ([]models.Cart, error) {
	var carts []models.Cart

	if err := r.db.WithContext(ctx).Preload("CartItems.Product.ProductImages").Preload("CartItems.Product").Find(&carts).Error; err != nil {
		log.Printf("CartRepository.GetAllCarts: Error getting all carts: %v", err)
		return nil, fmt.Errorf("failed to get all carts: %w", err)
	}
	return carts, nil
}
