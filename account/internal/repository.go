package internal

import (
	"context"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

type Repository interface {
	Close()
	PutAccount(ctx context.Context, a Account) (*Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	ListAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error)
}

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(databaseURL string) (Repository, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
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

	err = db.AutoMigrate(&Account{})
	if err != nil {
		log.Println("Error during migrations:", err)
	}

	return &postgresRepository{db}, nil
}

func (repository *postgresRepository) Close() {
	sqlDB, err := repository.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func (repository *postgresRepository) Ping() error {
	sqlDB, err := repository.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

func (repository *postgresRepository) PutAccount(ctx context.Context, a Account) (*Account, error) {
	if err := repository.db.WithContext(ctx).Create(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (repository *postgresRepository) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	var account Account
	if err := repository.db.WithContext(ctx).First(&account, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (repository *postgresRepository) GetAccountByID(ctx context.Context, id string) (*Account, error) {
	var account Account
	if err := repository.db.WithContext(ctx).First(&account, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (repository *postgresRepository) ListAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	var accounts []Account
	if err := repository.db.WithContext(ctx).Offset(int(skip)).Limit(int(take)).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}
