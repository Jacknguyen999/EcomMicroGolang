package internal

import (
	"context"

	"github.com/thomas/EcommerceAPI/order/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository interface {
	Close()
	PutOrder(ctx context.Context, order *models.Order) error
	GetOrdersForAccount(ctx context.Context, accountId string) ([]models.Order, error)
}

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(databaseURl string) (Repository, error) {
	db, err := gorm.Open(postgres.Open(databaseURl), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.Order{}, &models.ProductsInfo{})

	return &postgresRepository{db}, nil
}

func (repository *postgresRepository) Close() {
	sqlDB, err := repository.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func (repository *postgresRepository) PutOrder(ctx context.Context, order *models.Order) error {
	tx := repository.db.WithContext(ctx).Begin()

	err := tx.WithContext(ctx).Create(&order).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, product := range order.Products {
		orderedProduct := models.ProductsInfo{
			OrderID:   order.ID,
			ProductID: product.ID,
			Quantity:  int(product.Quantity),
		}
		err = tx.Create(&orderedProduct).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	if err = tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (repository *postgresRepository) GetOrdersForAccount(ctx context.Context, accountId string) ([]models.Order, error) {
	// First, get all orders for the account
	var orders []models.Order
	err := repository.db.WithContext(ctx).
		Where("account_id = ?", accountId).
		Find(&orders).Error

	if err != nil {
		return nil, err
	}

	// For each order, get its products
	for i := range orders {
		// Get product infos
		var productInfos []models.ProductsInfo
		err = repository.db.WithContext(ctx).
			Where("order_id = ?", orders[i].ID).
			Find(&productInfos).Error

		if err != nil {
			return nil, err
		}

		// Convert ProductsInfo to OrderedProduct
		for _, pi := range productInfos {
			orders[i].Products = append(orders[i].Products, &models.OrderedProduct{
				ID:       pi.ProductID,
				Quantity: uint32(pi.Quantity),
				// Name, Description, and Price will be populated by the server
			})
		}
	}

	return orders, nil
}
