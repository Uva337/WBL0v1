package interfaces

import (
	"context"
	"github.com/Uva337/WBL0v1/internal/models"
)

type Repository interface {
	UpsertOrder(ctx context.Context, o models.Order) error
	GetOrder(ctx context.Context, id string) (models.Order, bool, error)
	GetAll(ctx context.Context) ([]models.Order, error)
}
type Cache interface {
	Get(id string) (models.Order, bool)
	Set(id string, order models.Order)
	BulkSet(list []models.Order)
}
type Validator interface {
	Struct(s any) error
}