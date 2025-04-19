package internal

import (
	"context"
	"errors"
	"strconv"

	"github.com/thomas/EcommerceAPI/pkg/auth"
	"github.com/thomas/EcommerceAPI/pkg/utils"
)

type Service interface {
	Register(ctx context.Context, name, email, password string) (string, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetAccount(ctx context.Context, id string) (*Account, error)
	GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error)
}

type Account struct {
	ID       uint   `gorm:"primaryKey;autoIncrement"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type accountService struct {
	repository  Repository
	authService auth.AuthService
}

func NewService(r Repository, j auth.AuthService) Service {
	return &accountService{r, j}
}

func (service accountService) Register(ctx context.Context, name, email, password string) (string, error) {
	_, err := service.repository.GetAccountByEmail(ctx, email)
	if err == nil {
		return "", errors.New("account already exists")
	}

	hashedPass, err := utils.HashPassword(password)
	if err != nil {
		return "", err
	}
	acc := Account{
		Name:     name,
		Email:    email,
		Password: hashedPass,
	}
	account, err := service.repository.PutAccount(ctx, acc)
	if err != nil {
		return "", err
	}
	token, err := service.authService.GenerateToken(strconv.Itoa(int(account.ID)))
	if err != nil {
		return "", err
	}
	return token, nil
}

func (service accountService) Login(ctx context.Context, email, password string) (string, error) {
	account, err := service.repository.GetAccountByEmail(ctx, email)
	if err == nil && utils.VerifyPassword(password, account.Password) {
		token, err := service.authService.GenerateToken(strconv.Itoa(int(account.ID)))
		if err == nil {
			return token, nil
		}
	}
	return "", err
}

func (service accountService) GetAccount(ctx context.Context, id string) (*Account, error) {
	return service.repository.GetAccountByID(ctx, id)
}

func (service accountService) GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}

	return service.repository.ListAccounts(ctx, skip, take)

}
