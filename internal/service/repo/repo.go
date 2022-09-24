package repo

import (
	"user-balance-service/pkg/postgres"
	"user-balance-service/pkg/rediscache"
)

type Repository struct {
	*AuthRepo
	*AccountRepo
	*HistoryRepo
}

func New(pg *postgres.Postgres, redisCache *rediscache.Redis) *Repository {
	return &Repository{
		AuthRepo:    NewAuthRepo(pg),
		AccountRepo: NewAccountRepo(pg, redisCache),
		HistoryRepo: NewHistoryRepo(pg, redisCache),
	}
}
