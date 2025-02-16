package service

import (
	"avito-shop/internal/domain/models"
	"avito-shop/internal/repository"
	"context"
)

type UserService interface {
	Register(ctx context.Context, username, password string) error
	Login(ctx context.Context, username, password string) (string, error)
	TransferCoins(ctx context.Context, fromUserID int64, toUsername string, amount int) error
}

type MerchandiseService interface {
	GetAll(ctx context.Context) ([]*models.Merchandise, error)
	BuyItem(ctx context.Context, userID int64, itemName string) error
}

type InfoService interface {
	GetUserInfo(ctx context.Context, userID int64) (*models.InfoResponse, error)
}

type Services struct {
	Users       UserService
	Merchandise MerchandiseService
	Info        InfoService
	TokenSecret string
}

type ServicesDeps struct {
	Repos       *repository.Repositories
	TokenSecret string
}

func NewServices(deps ServicesDeps) *Services {
	return &Services{
		Users: NewUserService(
			deps.Repos.Users,
			deps.Repos.Transactions,
			deps.TokenSecret,
		),
		Merchandise: NewMerchandiseService(
			deps.Repos.Users,
			deps.Repos.Merchandise,
			deps.Repos.Inventory,
			deps.Repos.Transactions,
		),
		Info: NewInfoService(
			deps.Repos.Users,
			deps.Repos.Merchandise,
			deps.Repos.Transactions,
			deps.Repos.Inventory,
		),
		TokenSecret: deps.TokenSecret,
	}
}
