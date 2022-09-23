package service

import (
	"context"
	"user-balance-api/internal/entity"
)

type AccountService struct {
	repo AccountRepo
	wapi ConverterWEBAPI
}

func NewAccountService(repo AccountRepo, wapi ConverterWEBAPI) *AccountService {
	return &AccountService{
		repo: repo,
		wapi: wapi,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context) (int, error) {
	return s.repo.CreateAccount(ctx)
}

func (s *AccountService) DeleteAccount(ctx context.Context, id int) error {
	return s.repo.DeleteAccount(ctx, id)
}

func (s *AccountService) WriteOff(ctx context.Context, id, amount int) error {
	return s.repo.WriteOff(ctx, id, amount)
}

func (s *AccountService) GetAccount(ctx context.Context, id int) (entity.Account, error) {
	return s.repo.GetAccount(ctx, id)
}

func (s *AccountService) MakeDeposit(ctx context.Context, id, amount int) error {
	return s.repo.MakeDeposit(ctx, id, amount)
}

func (s *AccountService) TransferMoney(ctx context.Context, idFrom, idTo, amount int) error {
	return s.repo.TransferMoney(ctx, idFrom, idTo, amount)
}

func (s *AccountService) ConvertToCurrency(ctx context.Context, currencyTo string, amount float64) (float64, error) {
	return s.wapi.ConvertToCurrency(ctx, currencyTo, amount)
}

func (s *AccountService) AccountData(ctx context.Context, id int) (entity.Account, error) {
	return s.repo.AccountData(ctx, id)
}
