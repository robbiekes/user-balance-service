package repo

import (
	"user-balance-api/internal/service/rediscache"
	"user-balance-api/pkg/postgres"
)

type Repository struct {
	*AuthRepo
	*AccountRepo
	*HistoryRepo
}

func New(pg *postgres.Postgres, redisCache *rediscache.RedisLib) *Repository {
	return &Repository{
		AuthRepo:    NewAuthRepo(pg),
		AccountRepo: NewAccountRepo(pg, redisCache),
		HistoryRepo: NewHistoryRepo(pg, redisCache),
	}
}
