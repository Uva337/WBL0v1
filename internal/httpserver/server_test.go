package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/Uva337/WBL0v1/internal/models"
)

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(id string) (models.Order, bool) {
	args := m.Called(id)
	return args.Get(0).(models.Order), args.Bool(1)
}

func (m *MockCache) Set(id string, order models.Order) {
	m.Called(id, order)
}

func (m *MockCache) BulkSet(list []models.Order) {
	m.Called(list)
}

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) UpsertOrder(ctx context.Context, o models.Order) error {
	args := m.Called(ctx, o)
	return args.Error(0)
}

func (m *MockRepo) GetOrder(ctx context.Context, id string) (models.Order, bool, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.Order), args.Bool(1), args.Error(2)
}

func (m *MockRepo) GetAll(ctx context.Context) ([]models.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Order), args.Error(1)
}

func TestHandleGetOrder_FromCache(t *testing.T) {
	mockCache := new(MockCache)
	mockRepo := new(MockRepo)
	server := New(mockCache, mockRepo)

	orderID := "test-order-123"
	expectedOrder := models.Order{OrderUID: orderID, CustomerID: "test-customer"}

	mockCache.On("Get", orderID).Return(expectedOrder, true)

	req := httptest.NewRequest("GET", "/api/order/"+orderID, nil)
	rr := httptest.NewRecorder()

	server.handleGetOrder(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var returnedOrder models.Order
	err := json.NewDecoder(rr.Body).Decode(&returnedOrder)
	assert.NoError(t, err)
	assert.Equal(t, expectedOrder, returnedOrder)

	mockCache.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetOrder")
}

func TestHandleGetOrder_FromRepo(t *testing.T) {
	mockCache := new(MockCache)
	mockRepo := new(MockRepo)
	server := New(mockCache, mockRepo)

	orderID := "test-order-456"
	expectedOrder := models.Order{
		OrderUID:    orderID,
		CustomerID:  "test-customer-2",
		DateCreated: time.Now(), 
	}

	
	mockCache.On("Get", orderID).Return(models.Order{}, false)

	mockRepo.On("GetOrder", mock.Anything, orderID).Return(expectedOrder, true, nil)

	mockCache.On("Set", orderID, expectedOrder).Return()

	req := httptest.NewRequest("GET", "/api/order/"+orderID, nil)
	rr := httptest.NewRecorder()

	server.handleGetOrder(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var returnedOrder models.Order
	err := json.NewDecoder(rr.Body).Decode(&returnedOrder)
	assert.NoError(t, err)


	assert.Equal(t, expectedOrder.OrderUID, returnedOrder.OrderUID)
	assert.Equal(t, expectedOrder.CustomerID, returnedOrder.CustomerID)

	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestHandleGetOrder_NotFound(t *testing.T) {
	mockCache := new(MockCache)
	mockRepo := new(MockRepo)
	server := New(mockCache, mockRepo)

	orderID := "not-found-id"

	
	mockCache.On("Get", orderID).Return(models.Order{}, false)

	mockRepo.On("GetOrder", mock.Anything, orderID).Return(models.Order{}, false, nil)

	req := httptest.NewRequest("GET", "/api/order/"+orderID, nil)
	rr := httptest.NewRecorder()

	server.handleGetOrder(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockCache.AssertNotCalled(t, "Set")

}
