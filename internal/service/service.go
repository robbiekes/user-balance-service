package service

import (
	"user-balance-api/internal/service/repo"
	"user-balance-api/internal/service/webapi"
)

type Service struct {
	Auth
	Account
	History
}

func New(repo *repo.Repository, wapi *webapi.ConverterAPI) *Service {
	return &Service{
		Auth:    NewAuthService(repo),
		Account: NewAccountService(repo, wapi),
		History: NewHistoryService(repo),
	}
}
