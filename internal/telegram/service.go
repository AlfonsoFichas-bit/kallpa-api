package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"kallpa-api/internal/chat"
	"kallpa-api/internal/customers"
	"kallpa-api/internal/domain"
	"kallpa-api/internal/sales"
	"kallpa-api/internal/stores"
)

// Service defines operations for processing incoming Telegram webhooks.
type Service interface {
	ProcessWebhook(ctx context.Context, token string, update *Update) error
}

type service struct {
	client      Client
	storeSvc    stores.Service
	chatSvc     chat.Service
	salesSvc    sales.SalesService
	chatRepo    chat.Repository
	customerSvc customers.CustomerService
}

// NewService creates a new webhook service.
func NewService(client Client, storeSvc stores.Service, chatSvc chat.Service, salesSvc sales.SalesService, chatRepo chat.Repository, customerSvc customers.CustomerService) Service {
	return &service{
		client:      client,
		storeSvc:    storeSvc,
		chatSvc:     chatSvc,
		salesSvc:    salesSvc,
		chatRepo:    chatRepo,
		customerSvc: customerSvc,
	}
}

// ProcessWebhook interprets the telegram update and executes appropriate logic.
func (s *service) ProcessWebhook(ctx context.Context, botToken string, update *Update) error {
	if update.Message == nil || update.Message.From == nil || update.Message.Chat == nil {
		return nil // Ignore non-message updates
	}

	// 1. Identify Store by Bot Token
	store, err := s.getStoreByToken(ctx, botToken)
	if err != nil {
		slog.Error("Webhook received for unknown token", "error", err)
		return err // Or return nil to not retry from Telegram
	}
	if store == nil {
		return fmt.Errorf("store not found for token")
	}

	msg := update.Message
	text := strings.TrimSpace(msg.Text)
	chatID := msg.Chat.ID
	telegramUserID := msg.From.ID
	userName := msg.From.FirstName
	if userName == "" {
		userName = "Usuario"
	}

	// 2. Handle Commands
	if text == "/start" {
		reply := "¡Hola! Soy Kallpa Sales AI v2. ¿En qué puedo ayudarte?"
		_ = s.client.SendMessage(ctx, botToken, chatID, reply)
		return nil
	}

	if text == "/reset" {
		err := s.resetChatSession(ctx, store.ID, telegramUserID)
		reply := "Contexto reiniciado."
		if err != nil {
			reply = "Hubo un problema reiniciando el contexto."
		}
		_ = s.client.SendMessage(ctx, botToken, chatID, reply)
		return nil
	}

	// 3. Handle Photos (Payment proofs)
	if len(msg.Photo) > 0 {
		return s.handlePhoto(ctx, store.ID, botToken, telegramUserID, chatID, msg.Photo)
	}

	// 4. Handle Standard Text (Chat Orchestration)
	if text != "" {
		req := &chat.MessageRequest{
			StoreID:        store.ID,
			ChatID:         strconv.FormatInt(chatID, 10),
			TelegramUserID: telegramUserID,
			UserName:       userName,
			Message:        text,
		}

		res, err := s.chatSvc.ProcessMessage(ctx, req)
		if err != nil {
			slog.Error("Failed to process message with AI", "error", err)
			//nolint:misspell // "problemas" is Spanish
			_ = s.client.SendMessage(ctx, botToken, chatID, "Lo siento, estoy teniendo problemas técnicos en este momento.")
			return err
		}

		// Check for dynamic QR flag equivalent if order created
		if res.QRURL != nil {
			reply := res.Reply + "\n\nPor favor realiza el pago usando este link/QR: " + *res.QRURL
			_ = s.client.SendMessage(ctx, botToken, chatID, reply)
		} else {
			_ = s.client.SendMessage(ctx, botToken, chatID, res.Reply)
		}
		return nil
	}

	return nil
}

// getStoreByToken is a helper to find the store by bot_token.
func (s *service) getStoreByToken(ctx context.Context, token string) (*domain.Store, error) {
	// We need to fetch all active stores and match the token.
	// Since we don't have GetStoreByToken in the store service explicitly, we can list and filter.
	// In a real production system with many stores, we'd add a dedicated repo method.
	storesList, err := s.storeSvc.ListActiveStores(ctx)
	if err != nil {
		return nil, err
	}
	for i := range storesList {
		if storesList[i].BotToken != nil && *storesList[i].BotToken == token {
			return &storesList[i], nil
		}
	}
	return nil, nil
}

func (s *service) resetChatSession(ctx context.Context, storeID, customerTelegramID int64) error {
	// 1. Get customer
	customersList, err := s.customerSvc.GetCustomers(ctx, storeID, "")
	if err != nil {
		return err
	}

	var custID int64 = 0
	for _, c := range customersList {
		if c.TelegramID != nil && *c.TelegramID == customerTelegramID {
			custID = c.ID
			break
		}
	}
	if custID == 0 {
		return fmt.Errorf("customer not found")
	}

	// 2. Fetch active session
	session, err := s.chatRepo.GetActiveSession(ctx, storeID, custID)
	if err != nil {
		return err
	}

	// 3. Mark inactive
	if session != nil {
		session.IsActive = false
		return s.chatRepo.UpsertSession(ctx, session)
	}

	return nil
}

func (s *service) handlePhoto(ctx context.Context, storeID int64, botToken string, customerTelegramID, chatID int64, photos []PhotoSize) error {
	// Get highest resolution photo (last in array)
	bestPhoto := photos[len(photos)-1]

	_, err := s.client.GetFileURL(ctx, botToken, bestPhoto.FileID)
	if err != nil {
		slog.Error("Failed to get file URL", "error", err)
		_ = s.client.SendMessage(ctx, botToken, chatID, "No se pudo procesar la imagen.")
		return err
	}

	// 1. Get customer
	customersList, err := s.customerSvc.GetCustomers(ctx, storeID, "")
	if err != nil {
		return err
	}

	var custID int64 = 0
	for _, c := range customersList {
		if c.TelegramID != nil && *c.TelegramID == customerTelegramID {
			custID = c.ID
			break
		}
	}

	if custID != 0 {
		// 2. Fetch latest orders to find one pending payment
		orders, err := s.salesSvc.ListOrders(ctx, storeID, string(domain.OrderStatusPendingPayment))
		if err == nil {
			var pendingOrder *domain.Order
			for i := range orders {
				if orders[i].CustomerID == custID {
					pendingOrder = &orders[i]
					break
				}
			}

			// 3. Update order status
			if pendingOrder != nil {
				// In a real app we would save the URL to PaymentProofURL as well.
				// For now we just update the status to paid
				err = s.salesSvc.UpdateOrderStatus(ctx, storeID, pendingOrder.ID, domain.OrderStatusPaid)
				if err == nil {
					_ = s.client.SendMessage(ctx, botToken, chatID, "Imagen recibida. Pago confirmado y pedido actualizado exitosamente.")
					return nil
				}
			}
		}
	}

	_ = s.client.SendMessage(ctx, botToken, chatID, "Imagen recibida. Verificaremos tu comprobante pronto.")
	return nil
}
