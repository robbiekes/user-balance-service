package repo

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"strconv"
	"user-balance-api/internal/entity"
	"user-balance-api/internal/service/rediscache"
	"user-balance-api/pkg/postgres"
)

type AccountRepo struct {
	*postgres.Postgres
	*rediscache.RedisLib
}

func NewAccountRepo(pg *postgres.Postgres, redisCache *rediscache.RedisLib) *AccountRepo {
	return &AccountRepo{
		Postgres: pg,
		RedisLib: redisCache,
	}
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

	return nil
}

func (a *AccountRepo) WriteOff(ctx context.Context, id, amount int) error {
	account, err := a.AccountData(ctx, id)
	if err != nil {
		return err
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

	return nil
}

func (a *AccountRepo) GetAccount(ctx context.Context, id int) (entity.Account, error) {
	value, err := a.RedisLib.Get(ctx, strconv.Itoa(id))
	if value != nil {
		return value.(entity.Account), err
	}

	account, err := a.AccountData(ctx, id)
	if err != nil {
		return entity.Account{}, err
	}
	err = a.RedisLib.Set(ctx, strconv.Itoa(id), account)
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

	return nil
}

func (a *AccountRepo) TransferMoney(ctx context.Context, idFrom, idTo, amount int) error {
	account, err := a.AccountData(ctx, idFrom)
	if err != nil {
		return err
	}

	if account.Balance-amount < 0 {
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
	return nil
}

func (a *AccountRepo) AccountData(ctx context.Context, id int) (entity.Account, error) {
	sql, args, err := a.Builder.
		Select("balance").
		From("accounts").
		Where("id = ?", id).
		ToSql()

	if err != nil {
		return entity.Account{}, fmt.Errorf("repo - AccountRepo - getAccount - a.Builder: %w", err)
	}

	var account entity.Account
	err = a.Pool.QueryRow(ctx, sql, args...).Scan(&account.Balance)
	if err != nil {
		return entity.Account{}, fmt.Errorf("repo - AccountRepo - getAccount - a.Pool.QueryRow: %w", err)
	}

	return account, nil
}
