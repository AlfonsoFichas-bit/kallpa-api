package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"kallpa-api/internal/chat"
	"kallpa-api/internal/domain"
	"kallpa-api/internal/telegram"
)

func TestTelegramService_ProcessWebhook(t *testing.T) {
	mockClient := new(MockTelegramClient)
	mockStoreSvc := new(MockStoreService)
	mockChatSvc := new(MockChatService)
	mockSalesSvc := new(MockSalesService)
	mockChatRepo := new(MockChatRepository)
	mockCustSvc := new(MockCustomerService)

	svc := telegram.NewService(mockClient, mockStoreSvc, mockChatSvc, mockSalesSvc, mockChatRepo, mockCustSvc)

	botToken := "bot123"
	storeToken := "bot123"

	// Mock getting the store by token
	stores := []domain.Store{
		{ID: 1, BotToken: &storeToken},
	}
	mockStoreSvc.On("ListActiveStores", mock.Anything).Return(stores, nil)

	t.Run("Start Command", func(t *testing.T) {
		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 100},
				From:      &telegram.User{ID: 200, FirstName: "Test"},
				Text:      "/start",
			},
		}

		mockClient.On("SendMessage", mock.Anything, botToken, int64(100), "¡Hola! Soy Kallpa Sales AI v2. ¿En qué puedo ayudarte?").Return(nil).Once()

		err := svc.ProcessWebhook(context.Background(), botToken, update)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Reset Command - Success", func(t *testing.T) {
		update := &telegram.Update{
			UpdateID: 2,
			Message: &telegram.Message{
				MessageID: 2,
				Chat:      &telegram.Chat{ID: 100},
				From:      &telegram.User{ID: 200, FirstName: "Test"},
				Text:      "/reset",
			},
		}

		telegramID := int64(200)
		customers := []domain.Customer{{ID: 5, TelegramID: &telegramID}}
		// Add .Return(customers, nil).Once() matching exactly what gets called
		mockCustSvc.On("GetCustomers", mock.Anything, int64(1), "").Return(customers, nil).Once()

		session := &domain.ChatSession{ID: 10, IsActive: true}
		mockChatRepo.On("GetActiveSession", mock.Anything, int64(1), int64(5)).Return(session, nil).Once()

		mockChatRepo.On("UpsertSession", mock.Anything, mock.MatchedBy(func(s *domain.ChatSession) bool { return s.ID == 10 && !s.IsActive })).Return(nil).Once()

		mockClient.On("SendMessage", mock.Anything, botToken, int64(100), "Contexto reiniciado.").Return(nil).Once()

		err := svc.ProcessWebhook(context.Background(), botToken, update)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Standard Text Message", func(t *testing.T) {
		update := &telegram.Update{
			UpdateID: 3,
			Message: &telegram.Message{
				MessageID: 3,
				Chat:      &telegram.Chat{ID: 100},
				From:      &telegram.User{ID: 200, FirstName: "Test"},
				Text:      "Quiero comprar",
			},
		}

		chatRes := &chat.MessageResponse{
			Reply:  "Claro, procesando.",
			Intent: "buy",
		}

		mockChatSvc.On("ProcessMessage", mock.Anything, mock.MatchedBy(func(req *chat.MessageRequest) bool {
			return req.Message == "Quiero comprar" && req.StoreID == 1
		})).Return(chatRes, nil).Once()

		mockClient.On("SendMessage", mock.Anything, botToken, int64(100), "Claro, procesando.").Return(nil).Once()

		err := svc.ProcessWebhook(context.Background(), botToken, update)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Photo Message (Payment Proof)", func(t *testing.T) {
		update := &telegram.Update{
			UpdateID: 4,
			Message: &telegram.Message{
				MessageID: 4,
				Chat:      &telegram.Chat{ID: 100},
				From:      &telegram.User{ID: 200, FirstName: "Test"},
				Photo: []telegram.PhotoSize{
					{FileID: "file1", FileSize: 100},
					{FileID: "file2", FileSize: 200}, // Highest res
				},
			},
		}

		mockClient.On("GetFileURL", mock.Anything, botToken, "file2").Return("http://url", nil).Once()

		telegramID := int64(200)
		customers := []domain.Customer{{ID: 5, TelegramID: &telegramID}}
		mockCustSvc.On("GetCustomers", mock.Anything, int64(1), "").Return(customers, nil).Once()

		pendingOrders := []domain.Order{
			{ID: 50, CustomerID: 5, Status: domain.OrderStatusPendingPayment},
		}
		mockSalesSvc.On("ListOrders", mock.Anything, int64(1), string(domain.OrderStatusPendingPayment)).Return(pendingOrders, nil).Once()

		mockSalesSvc.On("UpdateOrderStatus", mock.Anything, int64(1), int64(50), domain.OrderStatusPaid).Return(nil).Once()

		mockClient.On("SendMessage", mock.Anything, botToken, int64(100), "Imagen recibida. Pago confirmado y pedido actualizado exitosamente.").Return(nil).Once()

		err := svc.ProcessWebhook(context.Background(), botToken, update)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})
}
