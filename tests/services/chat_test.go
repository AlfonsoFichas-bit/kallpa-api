package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"kallpa-api/internal/ai"
	"kallpa-api/internal/chat"
	"kallpa-api/internal/domain"
	"kallpa-api/internal/inventory"
)

type MockCustomerService struct {
	mock.Mock
}
func (m *MockCustomerService) UpsertCustomerFromTelegram(ctx context.Context, storeID, telegramID int64, name string) (*domain.Customer, error) {
	args := m.Called(ctx, storeID, telegramID, name)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockCustomerService) GetCustomers(ctx context.Context, storeID int64, status string) ([]domain.Customer, error) {
	args := m.Called(ctx, storeID, status)
	if args.Get(0) != nil {
		return args.Get(0).([]domain.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockCustomerService) GetCustomer(ctx context.Context, storeID, id int64) (*domain.Customer, error) {
	return nil, nil // unused here
}

func TestChatService_ProcessMessage(t *testing.T) {
	mockRepo := new(MockChatRepository)
	mockCust := new(MockCustomerService)
	mockAI := new(MockAIProvider)
	mockSales := new(MockSalesService)

	svc := chat.NewService(mockRepo, mockCust, mockAI, mockSales)

	req := &chat.MessageRequest{
		StoreID:        1,
		ChatID:         "123",
		TelegramUserID: 456,
		UserName:       "User",
		Message:        "Hello",
	}

	t.Run("General Message Flow", func(t *testing.T) {
		customer := &domain.Customer{ID: 10}
		mockCust.On("UpsertCustomerFromTelegram", mock.Anything, int64(1), int64(456), "User").Return(customer, nil).Once()

		session := &domain.ChatSession{ID: 1, ContextData: []byte(`{}`)}
		mockRepo.On("GetActiveSession", mock.Anything, int64(1), int64(10)).Return(session, nil).Once()

		aiRes := &ai.ProcessResult{
			Reply: "Hi", Intent: "greeting", Score: 0.9,
		}
		mockAI.On("ProcessMessage", mock.Anything, int64(1), "Hello", mock.AnythingOfType("map[string]interface {}")).Return(aiRes, nil).Once()

		mockRepo.On("UpsertSession", mock.Anything, mock.MatchedBy(func(s *domain.ChatSession) bool {
			return *s.CurrentStep == "greeting"
		})).Return(nil).Once()

		mockRepo.On("CreateLog", mock.Anything, mock.AnythingOfType("*domain.AiLog")).Return(nil).Once()

		res, err := svc.ProcessMessage(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, "Hi", res.Reply)
		assert.Equal(t, "greeting", res.Intent)
		assert.False(t, res.RequiresConfirmation)
	})

	t.Run("Confirm Order - Success", func(t *testing.T) {
		reqOrder := &chat.MessageRequest{
			StoreID: 1, TelegramUserID: 456, UserName: "User", Message: "Si",
		}

		customer := &domain.Customer{ID: 10}
		mockCust.On("UpsertCustomerFromTelegram", mock.Anything, int64(1), int64(456), "User").Return(customer, nil).Once()

		mockRepo.On("GetActiveSession", mock.Anything, int64(1), int64(10)).Return(nil, nil).Once()

		aiRes := &ai.ProcessResult{
			Reply: "Confirmed", Intent: "confirm_order", Score: 0.9,
		}
		mockAI.On("ProcessMessage", mock.Anything, int64(1), "Si", mock.Anything).Return(aiRes, nil).Once()

		mockRepo.On("UpsertSession", mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("CreateLog", mock.Anything, mock.Anything).Return(nil).Once()

		order := &domain.Order{ID: 100}
		mockSales.On("CreateOrder", mock.Anything, int64(1), mock.MatchedBy(func(o *domain.Order) bool {
			return o.CustomerID == 10 && len(o.Items) == 1
		})).Return(order, nil).Once()

		res, err := svc.ProcessMessage(context.Background(), reqOrder)
		assert.NoError(t, err)
		assert.Equal(t, "confirm_order", res.Intent)
		assert.NotNil(t, res.OrderID)
		assert.Equal(t, int64(100), *res.OrderID)
		assert.NotNil(t, res.QRURL)
	})

	t.Run("Confirm Order - Stock Error", func(t *testing.T) {
		reqOrder := &chat.MessageRequest{
			StoreID: 1, TelegramUserID: 456, UserName: "User", Message: "Si",
		}

		customer := &domain.Customer{ID: 10}
		mockCust.On("UpsertCustomerFromTelegram", mock.Anything, int64(1), int64(456), "User").Return(customer, nil).Once()
		mockRepo.On("GetActiveSession", mock.Anything, int64(1), int64(10)).Return(nil, nil).Once()

		aiRes := &ai.ProcessResult{Reply: "Confirmed", Intent: "confirm_order"}
		mockAI.On("ProcessMessage", mock.Anything, int64(1), "Si", mock.Anything).Return(aiRes, nil).Once()

		mockRepo.On("UpsertSession", mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("CreateLog", mock.Anything, mock.Anything).Return(nil).Once()

		mockSales.On("CreateOrder", mock.Anything, int64(1), mock.Anything).Return(nil, inventory.ErrInsufficientStock).Once()

		res, err := svc.ProcessMessage(context.Background(), reqOrder)
		assert.NoError(t, err) // It doesn't fail the request, just maps the reply
		assert.Equal(t, "order_failed", res.Intent)
		assert.Contains(t, res.Reply, "no hay suficiente stock")
	})

	t.Run("AI Provider Error", func(t *testing.T) {
		customer := &domain.Customer{ID: 10}
		mockCust.On("UpsertCustomerFromTelegram", mock.Anything, int64(1), int64(456), "User").Return(customer, nil).Once()
		mockRepo.On("GetActiveSession", mock.Anything, int64(1), int64(10)).Return(nil, nil).Once()

		mockAI.On("ProcessMessage", mock.Anything, int64(1), "Hello", mock.Anything).Return(nil, errors.New("ai offline")).Once()

		res, err := svc.ProcessMessage(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}
