package service

import (
	"context"
	"user-balance-service/internal/entity"
)

type (
	Auth interface {
		CreateUser(context.Context, entity.User) (int, error)
		GenerateToken(context.Context, string, string) (string, error)
		ParseToken(token string) (int, error)
	}

	Account interface {
		CreateAccount(ctx context.Context) (int, error)
		WriteOff(ctx context.Context, id, amount int) error
		GetAccount(ctx context.Context, id int) (entity.Account, error)
		MakeDeposit(ctx context.Context, id, amount int) error
		TransferMoney(ctx context.Context, idFrom, idTo, amount int) error
		ConvertToCurrency(ctx context.Context, currencyTo string, amount float64) (float64, error)
		DeleteAccount(ctx context.Context, id int) error
	}

	History interface {
		ShowAll(ctx context.Context) ([]entity.History, error)
		ShowById(ctx context.Context, id int) ([]entity.History, error)
		ShowSorted(ctx context.Context, sortType string, accountId int) ([]entity.History, error)
		Pagination(ctx context.Context, limit int, param string, accountId int) ([]entity.History, error)
		SaveHistory(ctx context.Context, input entity.History) (int, error)
	}

	AuthRepo interface {
		CreateUser(context.Context, entity.User) (int, error)
		GetUser(context.Context, string, string) (entity.User, error)
	}

	AccountRepo interface {
		CreateAccount(ctx context.Context) (int, error)
		WriteOff(ctx context.Context, id, amount int) error
		GetAccount(ctx context.Context, id int) (entity.Account, error)
		MakeDeposit(ctx context.Context, id, amount int) error
		TransferMoney(ctx context.Context, idFrom, idTo, amount int) error
		DeleteAccount(ctx context.Context, id int) error
	}

	HistoryRepo interface {
		ShowAll(ctx context.Context) ([]entity.History, error)
		ShowById(ctx context.Context, id int) ([]entity.History, error)
		ShowSorted(ctx context.Context, sortType string, accountId int) ([]entity.History, error)
		Pagination(ctx context.Context, limit int, param string, accountId int) ([]entity.History, error)
		SaveHistory(ctx context.Context, input entity.History) (int, error)
	}

	RedisCache interface {
		Set(ctx context.Context, key string, value any) error
		Get(ctx context.Context, key string) (any, error)
	}

	ConverterWEBAPI interface {
		ConvertToCurrency(ctx context.Context, currencyTo string, amount float64) (float64, error)
	}
)
