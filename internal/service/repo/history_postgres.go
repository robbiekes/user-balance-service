package repo

import (
	"context"
	"fmt"
	"strconv"
	"time"
	v1 "user-balance-api/internal/controller/http/v1"
	"user-balance-api/internal/entity"
	"user-balance-api/internal/service/rediscache"
	"user-balance-api/pkg/postgres"
)

const (
	CURSORDEFAULT   = "1700-01-01"
	MAXLIMIT        = 10
	ALL_HISTORY     = "all"
	ACCOUNT_HISTORY = "history of "
)

type HistoryRepo struct {
	*postgres.Postgres
	*rediscache.RedisLib
}

func NewHistoryRepo(pg *postgres.Postgres, redisCache *rediscache.RedisLib) *HistoryRepo {
	return &HistoryRepo{
		Postgres: pg,
		RedisLib: redisCache,
	}
}

func (h *HistoryRepo) ShowAll(ctx context.Context) ([]entity.History, error) {
	// pull out from cache
	value, err := h.RedisLib.Get(ctx, ALL_HISTORY)
	if value != nil {
		accounts, ok := value.([]entity.History)
		if ok {
			return accounts, fmt.Errorf("repo - HistoryRepo - ShowAll - h.RedisLib.Get: %w", err)
		}
		return nil, err
	}

	// pull out from db
	sql, args, err := h.Builder.
		Select("id", "type", "description", "amount", "account_id", "date").
		From("history").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowAll - a.Builder: %w", err)
	}

	var accounts []entity.History
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

	// put in cache
	err = h.RedisLib.Set(ctx, ALL_HISTORY, accounts)
	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowAll - h.RedisLib.Set: %w", err)
	}

	return accounts, nil
}

func (h *HistoryRepo) ShowById(ctx context.Context, id int) ([]entity.History, error) {
	// pull out from cache
	accountId := strconv.Itoa(id)
	value, err := h.RedisLib.Get(ctx, ALL_HISTORY+" "+accountId)
	if value != nil {
		accounts, ok := value.([]entity.History)
		if ok {
			return accounts, fmt.Errorf("repo - HistoryRepo - ShowById - value.([]entity.History): %w", err)
		}
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowById - h.RedisLib.Get: %w", err)
	}

	// pull out from db
	sql, args, err := h.Builder.
		Select("id", "type", "description", "amount", "account_id", "date").
		From("history").
		Where("account_id = ?", id).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowById - a.Builder: %w", err)
	}

	var accounts []entity.History
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

	// put in cache
	err = h.RedisLib.Set(ctx, ALL_HISTORY+" "+accountId, accounts)
	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowAll - h.RedisLib.Set: %w", err)
	}

	return accounts, nil
}

func (h *HistoryRepo) ShowSorted(ctx context.Context, srt v1.SortArgs) ([]entity.History, error) {
	var (
		sql  string
		args []interface{}
		err  error
	)

	// pull out from cache
	accountId := strconv.Itoa(srt.AccountId)
	value, err := h.RedisLib.Get(ctx, srt.Type+" "+accountId)
	if value != nil {
		accounts, ok := value.([]entity.History)
		if ok {
			return accounts, nil
		}
		return nil, fmt.Errorf("repo - HistoryRepo - ShowSorted - value.([]entity.History): %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowSorted - h.RedisLib.Get: %w", err)
	}

	// pull out from db
	switch {
	case srt.AccountId == 0:
		sql, args, err = h.Builder.
			Select("id", "type", "description", "amount", "account_id", "date").
			From("history").OrderBy(srt.Type).
			ToSql()
	case srt.AccountId != 0:
		sql, args, err = h.Builder.
			Select("id", "type", "description", "amount", "account_id", "date").
			From("history").
			Where("account_id = ?", srt.AccountId).
			OrderBy(srt.Type).
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

	// put in cache
	err = h.RedisLib.Set(ctx, srt.Type+" "+accountId, accounts)
	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowAll - h.RedisLib.Set: %w", err)
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

func (h *HistoryRepo) Pagination(ctx context.Context, pgn v1.PaginationArgs) ([]entity.History, error) {
	var (
		cursor string
		sql    string
		args   []interface{}
		err    error
	)

	// pull out from cache
	accountId := strconv.Itoa(pgn.AccountId)
	limit := strconv.Itoa(pgn.Limit)
	value, err := h.RedisLib.Get(ctx, limit+" "+pgn.Param+" "+accountId)
	if value != nil {
		accounts, ok := value.([]entity.History)
		if ok {
			return accounts, nil
		}
		return nil, fmt.Errorf("repo - HistoryRepo - ShowSorted - value.([]entity.History): %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowSorted - h.RedisLib.Get: %w", err)
	}

	// correct parameters
	switch {
	case len(pgn.Param) == 0:
		cursor = CURSORDEFAULT
	case len(pgn.Param) != 0:
		cursor = pgn.Param
	}
	switch {
	case pgn.Limit > MAXLIMIT:
		pgn.Limit = MAXLIMIT
	}

	// do request
	switch {
	case pgn.AccountId == 0:
		sql, args, err = h.Builder.
			Select("id", "type", "description", "amount", "account_id", "date").
			From("history").
			Where("date > ?", cursor).
			OrderBy("date DESC").
			Limit(uint64(pgn.Limit)).
			ToSql()
	case pgn.AccountId != 0:
		sql, args, err = h.Builder.
			Select("id", "type", "description", "amount", "account_id", "date").
			From("history").
			Where("account_id = ?", pgn.AccountId).
			Where("date > ?", cursor).
			OrderBy("date DESC").
			Limit(uint64(pgn.Limit)).
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

	// put in cache
	err = h.RedisLib.Set(ctx, limit+" "+pgn.Param+" "+accountId, accounts)
	if err != nil {
		return nil, fmt.Errorf("repo - HistoryRepo - ShowAll - h.RedisLib.Set: %w", err)
	}

	return accounts, nil
}
