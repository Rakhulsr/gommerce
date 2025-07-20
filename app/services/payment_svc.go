package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"

	"github.com/Rakhulsr/go-ecommerce/app/configs"
	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"gorm.io/gorm"
)

type MidtransNotificationPayload struct {
	TransactionStatus string `json:"transaction_status"`
	OrderID           string `json:"order_id"`
	PaymentType       string `json:"payment_type"`
	FraudStatus       string `json:"fraud_status"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`

	StatusCode string `json:"status_code"`
	Currency   string `json:"currency"`
}

type PaymentService struct {
	orderRepo             repositories.OrderRepository
	paymentRepo           repositories.PaymentRepositoryImpl
	db                    *gorm.DB
	midtransCoreAPIClient coreapi.Client
}

func NewPaymentService(
	orderRepo repositories.OrderRepository,
	paymentRepo repositories.PaymentRepositoryImpl,
	db *gorm.DB,
) *PaymentService {
	coreAPIClient := configs.GetMidtransCoreAPIClient()
	return &PaymentService{
		orderRepo:             orderRepo,
		paymentRepo:           paymentRepo,
		db:                    db,
		midtransCoreAPIClient: coreAPIClient,
	}
}

func (s *PaymentService) ProcessMidtransNotification(ctx context.Context, payload MidtransNotificationPayload) (
	newPaymentStatus string,
	newOrderStatus int,
	shouldReduceStock bool,
	shouldClearCart bool,
	shouldRefundStock bool,
	order *models.Order,
	err error,
) {
	log.Printf("PaymentService: Midtrans Notification Received for OrderCode: %s, Status: %s, FraudStatus: %s", payload.OrderID, payload.TransactionStatus, payload.FraudStatus)

	var transactionStatus *coreapi.TransactionStatusResponse
	var midtransErr *midtrans.Error

	transactionStatus, midtransErr = s.midtransCoreAPIClient.CheckTransaction(payload.OrderID)

	log.Printf("DEBUG: PaymentService: Result of CheckTransaction for %s: transactionStatus=%+v, midtransErr=%v", payload.OrderID, transactionStatus, midtransErr)

	if midtransErr != nil {
		log.Printf("ERROR: PaymentService: Failed to check transaction status with Midtrans API for OrderCode %s: %v", payload.OrderID, midtransErr.Error())
		return "", 0, false, false, false, nil, fmt.Errorf("failed to verify transaction with Midtrans (API error): %w", midtransErr.RawError)
	}

	if transactionStatus == nil {
		log.Printf("ERROR: PaymentService: Midtrans API returned nil transaction status for OrderCode %s (no API error). This indicates an unexpected response structure.", payload.OrderID)
		return "", 0, false, false, false, nil, errors.New("invalid transaction status from Midtrans API (nil response)")
	}

	if transactionStatus.StatusCode == "404" {
		log.Printf("WARNING: PaymentService: Order %s not found in Midtrans system (StatusCode: 404).", payload.OrderID)
		return "", 0, false, false, false, nil, errors.New("order not found in Midtrans system")
	}
	if len(transactionStatus.StatusCode) > 0 && transactionStatus.StatusCode[0] == '5' {
		log.Printf("ERROR: PaymentService: Midtrans API returned server error for OrderCode %s (StatusCode: %s).", payload.OrderID, transactionStatus.StatusCode)
		return "", 0, false, false, false, nil, fmt.Errorf("midtrans API server error: %s", transactionStatus.StatusCode)
	}

	if transactionStatus.TransactionStatus != payload.TransactionStatus ||
		transactionStatus.FraudStatus != payload.FraudStatus {
		log.Printf("WARNING: PaymentService: Mismatch in transaction status for OrderCode %s. API: %s/%s, Notification: %s/%s. Proceeding with API status.",
			payload.OrderID,
			transactionStatus.TransactionStatus, transactionStatus.FraudStatus,
			payload.TransactionStatus, payload.FraudStatus)
	}

	order, err = s.orderRepo.FindByCodeWithDetails(ctx, payload.OrderID)
	if err != nil {
		log.Printf("ERROR: PaymentService: Failed to find order %s: %v", payload.OrderID, err)
		return "", 0, false, false, false, nil, fmt.Errorf("order not found or database error: %w", err)
	}
	if order == nil {
		log.Printf("WARNING: PaymentService: Order %s not found in database.", payload.OrderID)
		return "", 0, false, false, false, nil, errors.New("order not found")
	}

	payment, err := s.paymentRepo.FindByOrderID(ctx, order.ID)
	if err != nil {
		log.Printf("ERROR: PaymentService: Failed to get payment for order %s: %v", order.ID, err)
		return "", 0, false, false, false, nil, fmt.Errorf("payment record not found or database error: %w", err)
	}
	if payment == nil {
		log.Printf("WARNING: PaymentService: Payment record for OrderID %s not found. This should not happen.", order.ID)
		return "", 0, false, false, false, nil, errors.New("payment record not found")
	}

	if order.Status == models.OrderStatusCompleted ||
		order.Status == models.OrderStatusCancelled ||
		order.Status == models.OrderStatusFailed ||
		order.Status == models.OrderStatusRefunded {
		log.Printf("INFO: PaymentService: Order %s already in final status (%d). Skipping update.", order.ID, order.Status)
		return payment.Status, order.Status, false, false, false, order, nil
	}
	if payment.Status == "Paid" || payment.Status == "Failed" || payment.Status == "Cancelled" || payment.Status == "Refunded" {
		log.Printf("INFO: PaymentService: Payment %s already in final status (%s). Skipping update.", payment.ID, payment.Status)
		return payment.Status, order.Status, false, false, false, order, nil
	}

	switch transactionStatus.TransactionStatus {
	case "capture", "settlement":
		if transactionStatus.FraudStatus == "accept" {
			newPaymentStatus = "Paid"
			newOrderStatus = models.OrderStatusProcessing

			if order.Status == models.OrderStatusPending {
				shouldReduceStock = true
				shouldClearCart = true
				log.Printf("INFO: PaymentService: Order %s is settling. Flags set for stock reduction and cart clearing.", order.ID)
			} else {
				log.Printf("INFO: PaymentService: Order %s already in a processed state (%d). No new stock reduction/cart clearing flags needed.", order.ID, order.Status)
			}
		} else {
			newPaymentStatus = "Failed"
			newOrderStatus = models.OrderStatusFailed

			if order.PaymentStatus == "Paid" || order.Status == models.OrderStatusProcessing || order.Status == models.OrderStatusCompleted {
				shouldRefundStock = true
				log.Printf("INFO: PaymentService: Order %s failed due to fraud. Flag set for stock refund.", order.ID)
			} else {
				log.Printf("INFO: PaymentService: Order %s failed due to fraud, but stock was not previously reduced. No refund flag needed.", order.ID)
			}
		}
	case "pending":
		newPaymentStatus = "Pending"
		newOrderStatus = models.OrderStatusPending
		log.Printf("INFO: PaymentService: Order %s is still pending. No stock/cart flags applied.", order.ID)
	case "deny", "expire", "cancel":
		newPaymentStatus = "Failed"
		newOrderStatus = models.OrderStatusCancelled

		if order.Status == models.OrderStatusProcessing || order.Status == models.OrderStatusCompleted {
			shouldRefundStock = true
			log.Printf("INFO: PaymentService: Order %s was cancelled/expired/denied after being processed. Flag set for stock refund.", order.ID)
		} else {
			log.Printf("INFO: PaymentService: Order %s was cancelled/expired/denied while pending. No stock/cart flags needed.", order.ID)
		}
	case "refund", "partial_refund":
		newPaymentStatus = "Refunded"
		newOrderStatus = models.OrderStatusRefunded
		shouldRefundStock = true
		log.Printf("INFO: PaymentService: Order %s is being refunded. Flag set for stock refund.", order.ID)
	default:
		log.Printf("WARNING: PaymentService: Unhandled transaction status from Midtrans: %s", transactionStatus.TransactionStatus)
		return "", 0, false, false, false, nil, errors.New("unhandled transaction status")
	}

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		err = s.paymentRepo.UpdatePaymentStatusTx(ctx, tx, payment.ID, newPaymentStatus)
		if err != nil {
			return fmt.Errorf("failed to update payment status for payment ID %s: %w", payment.ID, err)
		}

		err = s.orderRepo.UpdatePaymentStatusAndOrderStatus(ctx, tx, order.ID, newPaymentStatus, newOrderStatus)
		if err != nil {
			return fmt.Errorf("failed to update order status for order ID %s: %w", order.ID, err)
		}
		return nil
	})

	if txErr != nil {
		log.Printf("ERROR: PaymentService: During Midtrans notification transaction for OrderID %s (status update only): %v", order.ID, txErr)
		return "", 0, false, false, false, order, fmt.Errorf("internal server error during status update: %w", txErr)
	}

	log.Printf("SUCCESS: PaymentService: Order %s and Payment updated to PaymentStatus: %s, OrderStatus: %d. Flags: ReduceStock=%t, ClearCart=%t, RefundStock=%t", order.ID, newPaymentStatus, newOrderStatus, shouldReduceStock, shouldClearCart, shouldRefundStock)
	return newPaymentStatus, newOrderStatus, shouldReduceStock, shouldClearCart, shouldRefundStock, order, nil
}
