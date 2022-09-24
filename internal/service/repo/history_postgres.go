package repo

import (
	"context"
	"fmt"
	"time"
	"user-balance-service/internal/entity"
	"user-balance-service/pkg/postgres"
	"user-balance-service/pkg/rediscache"
)

const (
	defaultPaginationCursor   = "1700-01-01"
	maxPaginationLimit        = 10
	allHistoryRedisKey        = "all_history_data"
	historyByIdRedisKeyPrefix = "history_by_id"
)

func historyByIdRedisKey(id int) string {
	return fmt.Sprintf("%s_%d", historyByIdRedisKeyPrefix, id)
}

func sortedHistoryRedisKey(sortType string, id int) string {
	return fmt.Sprintf("%s_%d", sortType, id)
}

func paginationHistoryRedisKey(limit int, cursor string, id int) string {
	return fmt.Sprintf("%d_%s_%d", limit, cursor, id)
}

type HistoryRepo struct {
	*postgres.Postgres
	*rediscache.Redis
}

func NewHistoryRepo(pg *postgres.Postgres, redisCache *rediscache.Redis) *HistoryRepo {
	return &HistoryRepo{
		Postgres: pg,
		Redis:    redisCache,
	}
}

func (h *HistoryRepo) ShowAll(ctx context.Context) ([]entity.History, error) {
	// search in cache
	value, err := h.Redis.Get(ctx, allHistoryRedisKey)
	if value != nil {
		accounts, ok := value.([]entity.History)
		if ok {
			return accounts, nil
		}
	}
	// if err != nil {
	// 	return nil, fmt.Errorf("repo - HistoryRepo - ShowAll - h.Redis.Get: %w", err)
	// }

	// do request
	var accounts []entity.History
	sql, args, err := h.Builder.
		Select("id", "type", "description", "amount", "account_id", "date").
		From("history").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowAll - a.Builder: %w", err)
	}

	rows, err := h.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowAll - a.Pool.Query: %w", err)
	}

	for rows.Next() {
		var account entity.History
		err := rows.Scan(&account.Id, &account.Type, &account.Description, &account.Amount, &account.AccountId, &account.Date)
		if err != nil {
			return nil, fmt.Errorf("repo - HistoryRepo - ShowAll - rows.Scan: %w", err)
		}
		accounts = append(accounts, account)
	}

	// save in cache
	err = h.Redis.Set(ctx, allHistoryRedisKey, accounts)
	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowAll - h.Redis.Set: %w", err)
	}

	return accounts, nil
}

func (h *HistoryRepo) ShowById(ctx context.Context, id int) ([]entity.History, error) {
	// search in cache
	value, err := h.Redis.Get(ctx, accountRedisKey(id))
	if value != nil {
		accounts, ok := value.([]entity.History)
		if ok {
			return accounts, nil
		}
	}
	// if err != nil {
	// 	return nil, fmt.Errorf("repo - HistoryRepo - ShowById - h.Redis.Get: %w", err)
	// }

	// do request
	var accounts []entity.History
	sql, args, err := h.Builder.
		Select("id", "type", "description", "amount", "account_id", "date").
		From("history").
		Where("account_id = ?", id).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowById - a.Builder: %w", err)
	}

	rows, err := h.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowById - a.Pool.Query: %w", err)
	}

	for rows.Next() {
		var account entity.History
		err := rows.Scan(&account.Id, &account.Type, &account.Description, &account.Amount, &account.AccountId, &account.Date)
		if err != nil {
			return nil, fmt.Errorf("repo - HistoryRepo - ShowById - rows.Scan: %w", err)
		}
		accounts = append(accounts, account)
	}

	// save in cache
	err = h.Redis.Set(ctx, historyByIdRedisKey(id), accounts)
	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowAll - h.Redis.Set: %w", err)
	}

	return accounts, nil
}

func (h *HistoryRepo) ShowSorted(ctx context.Context, sortType string, accountId int) ([]entity.History, error) {
	var (
		sql  string
		args []interface{}
		err  error
	)

	// search in cache
	value, err := h.Redis.Get(ctx, sortedHistoryRedisKey(sortType, accountId))
	if value != nil {
		accounts, ok := value.([]entity.History)
		if ok {
			return accounts, nil
		}
	}
	// if err != nil {
	// 	return nil, fmt.Errorf("repo - HistoryRepo - ShowSorted - h.Redis.Get: %w", err)
	// }

	// do request
	switch {
	case accountId == 0:
		sql, args, err = h.Builder.
			Select("id", "type", "description", "amount", "account_id", "date").
			From("history").OrderBy(sortType).
			ToSql()
	case accountId != 0:
		sql, args, err = h.Builder.
			Select("id", "type", "description", "amount", "account_id", "date").
			From("history").
			Where("account_id = ?", accountId).
			OrderBy(sortType).
			ToSql()
	}
	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowSorted - a.Builder: %w", err)
	}

	rows, err := h.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowSorted - a.Pool.Query: %w", err)
	}

	var accounts []entity.History
	for rows.Next() {
		var account entity.History
		err := rows.Scan(&account.Id, &account.Type, &account.Description, &account.Amount, &account.AccountId, &account.Date)
		if err != nil {
			return nil, fmt.Errorf("repo - HistoryRepo - ShowSorted - rows.Scan: %w", err)
		}
		accounts = append(accounts, account)
	}

	// save in cache
	err = h.Redis.Set(ctx, sortedHistoryRedisKey(sortType, accountId), accounts)
	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowSorted - h.Redis.Set: %w", err)
	}

	return accounts, nil
}

func (h *HistoryRepo) SaveHistory(ctx context.Context, input entity.History) (int, error) {
	sql, args, err := h.Builder.
		Insert("history").
		Columns("type", "description", "amount", "account_id", "date").
		Values(input.Type, input.Description, input.Amount, input.AccountId, time.Time(input.Date)).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("repo - HistoryRepo - SaveHistory - h.Builder: %w", err)
	}

	var id int
	err = h.Pool.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("repo - HistoryRepo - SaveHistory - h.Pool.QueryRow: %w", err)
	}

	return id, nil
}

func (h *HistoryRepo) Pagination(ctx context.Context, limit int, param string, accountId int) ([]entity.History, error) {
	var (
		cursor string
		sql    string
		args   []interface{}
	)

	// correct parameters
	switch {
	case len(param) == 0:
		cursor = defaultPaginationCursor
	case len(param) != 0:
		cursor = param
	}
	switch {
	case limit > maxPaginationLimit:
		limit = maxPaginationLimit
	}

	// search in cache
	value, err := h.Redis.Get(ctx, paginationHistoryRedisKey(limit, cursor, accountId))
	if value != nil {
		accounts, ok := value.([]entity.History)
		if ok {
			return accounts, nil
		}
	}
	// if err != nil {
	// 	return nil, fmt.Errorf("repo - HistoryRepo - Pagination - h.Redis.Get: %w", err)
	// }

	// do request
	switch {
	case accountId == 0:
		sql, args, err = h.Builder.
			Select("id", "type", "description", "amount", "account_id", "date").
			From("history").
			Where("date > ?", cursor).
			OrderBy("date DESC").
			Limit(uint64(limit)).
			ToSql()
	case accountId != 0:
		sql, args, err = h.Builder.
			Select("id", "type", "description", "amount", "account_id", "date").
			From("history").
			Where("account_id = ?", accountId).
			Where("date > ?", cursor).
			OrderBy("date DESC").
			Limit(uint64(limit)).
			ToSql()
	}

	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowSorted - a.Builder: %w", err)
	}

	rows, err := h.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowSorted - a.Pool.Query: %w", err)
	}

	var accounts []entity.History
	for rows.Next() {
		var account entity.History
		err := rows.Scan(&account.Id, &account.Type, &account.Description, &account.Amount, &account.AccountId, &account.Date)
		if err != nil {
			return nil, fmt.Errorf("repo - HistoryRepo - ShowSorted - rows.Scan: %w", err)
		}
		accounts = append(accounts, account)
	}

	// save in cache
	err = h.Redis.Set(ctx, paginationHistoryRedisKey(limit, cursor, accountId), accounts)
	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - Pagination - h.Redis.Set: %w", err)
	}

	return accounts, nil
}
