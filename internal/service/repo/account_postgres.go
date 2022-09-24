package repo

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"user-balance-service/internal/entity"
	"user-balance-service/pkg/postgres"
	"user-balance-service/pkg/rediscache"
)

const accountRedisKeyPrefix = "account_id"

type AccountRepo struct {
	*postgres.Postgres
	*rediscache.Redis
}

func NewAccountRepo(pg *postgres.Postgres, redisCache *rediscache.Redis) *AccountRepo {
	return &AccountRepo{
		Postgres: pg,
		Redis:    redisCache,
	}
}

func accountRedisKey(id int) string {
	return fmt.Sprintf("%s_%d", accountRedisKeyPrefix, id)
}

func (a *AccountRepo) CreateAccount(ctx context.Context) (int, error) {
	sql, args, err := a.Builder.
		Insert("accounts").
		Columns("balance").
		Values(0).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return 0, fmt.Errorf("repo - AccountRepo - CreateAccount - a.Builder: %w", err)
	}

	var id int
	err = a.Pool.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("repo - AccountRepo - CreateAccount - a.Pool.QueryRow: %w", err)
	}

	return id, nil
}

func (a *AccountRepo) DeleteAccount(ctx context.Context, id int) error {
	sql, args, err := a.Builder.
		Delete("accounts").
		Where("id = ?", id).
		ToSql()

	if err != nil {
		return fmt.Errorf("repo - AccountRepo - DeleteAccount - a.Builder: %w", err)
	}

	_, err = a.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("repo - AccountRepo - DeleteAccount - a.Pool.Exec: %w", err)
	}

	// save changes in cache
	err = a.Redis.Set(ctx, accountRedisKey(id), nil)
	if err != nil {
		return fmt.Errorf("repo - AccountRepo - WriteOff - a.Redis.Set: %w", err)
	}

	return nil
}

func (a *AccountRepo) WriteOff(ctx context.Context, id, amount int) error {
	account, err := a.GetAccount(ctx, id)
	if err != nil {
		return fmt.Errorf("repo - AccountRepo - WriteOff - a.GetAccount: %w", err)
	}

	if account.Balance-amount >= 0 {
		sql, args, err := a.Builder.
			Update("accounts").
			Set("balance", squirrel.Expr("balance - ?", amount)).
			Where("id = ?", id).
			ToSql()
		if err != nil {
			return fmt.Errorf("repo - AccountRepo - WriteOff - a.Builder: %w", err)
		}

		_, err = a.Pool.Exec(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("repo - AccountRepo - WriteOff - a.Pool.Exec: %w", err)
		}
	} else {
		return fmt.Errorf("repo - AccountRepo - WriteOff - balance can't be less than 0")
	}

	// save changes in cache
	account.Balance -= amount
	err = a.Redis.Set(ctx, accountRedisKey(id), account)
	if err != nil {
		return fmt.Errorf("repo - AccountRepo - WriteOff - a.Redis.Set: %w", err)
	}

	return nil
}

func (a *AccountRepo) GetAccount(ctx context.Context, id int) (entity.Account, error) {
	// check in cache
	value, err := a.Redis.Get(ctx, accountRedisKey(id))
	if value != nil {
		account, ok := value.(entity.Account)
		if ok {
			return account, nil
		}
	}

	// do request
	var account entity.Account
	sql, args, err := a.Builder.
		Select("balance").
		From("accounts").
		Where("id = ?", id).
		ToSql()

	if err != nil {
		return entity.Account{}, fmt.Errorf("repo - AccountRepo - GetAccount - a.Builder: %w", err)
	}

	err = a.Pool.QueryRow(ctx, sql, args...).Scan(&account.Balance)
	if err != nil {
		return entity.Account{}, fmt.Errorf("repo - AccountRepo - GetAccount - a.Pool.QueryRow: %w", err)
	}

	// save in cache
	err = a.Redis.Set(ctx, accountRedisKey(id), account)
	if err != nil {
		return entity.Account{}, err
	}

	return account, nil
}

func (a *AccountRepo) MakeDeposit(ctx context.Context, id, amount int) error {
	sql, args, err := a.Builder.
		Update("accounts").
		Set("balance", squirrel.Expr("balance + ?", amount)).
		Where(squirrel.Eq{"id": id}).
		ToSql()

	if err != nil {
		return fmt.Errorf("repo - AccountRepo - MakeDeposit - r.Builder: %w", err)
	}

	_, err = a.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("repo - AccountRepo - MakeDeposit - r.Pool.Exec: %w", err)
	}

	// save changes in cache
	account, err := a.GetAccount(ctx, id)
	if err != nil {
		return fmt.Errorf("repo - AccountRepo - MakeDeposit - a.GetAccount: %w", err)
	}
	account.Balance += amount
	err = a.Redis.Set(ctx, accountRedisKey(id), account)
	if err != nil {
		return fmt.Errorf("repo - AccountRepo - MakeDeposit - a.Redis.Set: %w", err)
	}

	return nil
}

func (a *AccountRepo) TransferMoney(ctx context.Context, idFrom, idTo, amount int) error {
	accountFrom, err := a.GetAccount(ctx, idFrom)
	if err != nil {
		return err
	}
	accountTo, err := a.GetAccount(ctx, idTo)
	if err != nil {
		return err
	}

	if accountFrom.Balance-amount < 0 {
		return errors.New("repo - AccountRepo - TransferMoney - balance can't be less than 0")
	}

	tx, err := a.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repo - AccountRepo - TransferMoney - a.Pool.Begin: %w", err)
	}

	// first operation
	sql, args, err := a.Builder.
		Update("accounts").
		Set("balance", squirrel.Expr("balance - ?", amount)).
		Where(squirrel.Eq{"id": idFrom}).
		ToSql()
	if err != nil {
		return fmt.Errorf("repo - AccountRepo - TransferMoney - a.Builder: %w", err)
	}

	_, err = a.Pool.Exec(ctx, sql, args...)
	if err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("repo - AccountRepo - TransferMoney - a.Pool.Exec: %w", err)
	}

	// second operation
	sql, args, err = a.Builder.
		Update("accounts").
		Set("balance", squirrel.Expr("balance + ?", amount)).
		Where(squirrel.Eq{"id": idTo}).
		ToSql()
	if err != nil {
		return fmt.Errorf("repo - AccountRepo - TransferMoney - a.Builder: %w", err)
	}

	_, err = a.Pool.Exec(ctx, sql, args...)
	if err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("repo - AccountRepo - TransferMoney - a.Pool.Exec: %w", err)
	}

	tx.Commit(ctx)

	// save changes in cache
	accountFrom.Balance -= amount
	accountTo.Balance += amount

	err = a.Redis.Set(ctx, accountRedisKey(idFrom), accountFrom)
	if err != nil {
		return fmt.Errorf("repo - AccountRepo - TransferMoney - a.Redis.Set(idFrom): %w", err)
	}
	err = a.Redis.Set(ctx, accountRedisKey(idTo), accountTo)
	if err != nil {
		return fmt.Errorf("repo - AccountRepo - TransferMoney - a.Redis.Set(idTo): %w", err)
	}

	return nil
}
