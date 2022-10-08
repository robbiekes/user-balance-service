package repo

import (
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"testing"
	"user-balance-service/internal/entity"
	"user-balance-service/pkg/postgres"
	"user-balance-service/pkg/rediscache"
)

func TestAccountRepo_CreateAccount(t *testing.T) {
	// new mini Redis server
	miniRedis, err := miniredis.Run()
	if err != nil {
		t.Error()
	}
	client := redis.NewClient(&redis.Options{Addr: miniRedis.Addr()})
	redisCache := rediscache.New(client)

	// mock pool
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Error()
	}
	defer mockPool.Close()

	// mock Postgres
	postgresMock := &postgres.Postgres{
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		Pool:    mockPool,
	}

	accountRepoMock := NewAccountRepo(postgresMock, redisCache)

	rows := pgxmock.NewRows([]string{"id"}).AddRow(1)
	mockPool.ExpectQuery("INSERT INTO accounts").WillReturnRows(rows)

	id, err := accountRepoMock.CreateAccount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, id)

	err = mockPool.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAccountRepo_GetAccount(t *testing.T) {
	// new mini Redis server
	miniRedis, err := miniredis.Run()
	if err != nil {
		t.Error()
	}
	redisClientMock := redis.NewClient(&redis.Options{Addr: miniRedis.Addr()})
	redisCache := rediscache.New(redisClientMock)

	// mock pgx pool
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Error()
	}
	// defer mockPool.Close()

	// mock Postgres
	postgresMock := &postgres.Postgres{
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		Pool:    mockPool,
	}

	mockAccountRepo := NewAccountRepo(postgresMock, redisCache)

	type args struct {
		ctx context.Context
		id  int
	}

	type MockBehaviour func(args args, account entity.Account)

	testCases := []struct {
		name          string
		args          args
		mockBehaviour MockBehaviour
		want          entity.Account
		wantErr       bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			mockBehaviour: func(args args, account entity.Account) {
				rows := mockPool.NewRows([]string{"id", "balance"}).
					AddRow(1, 500)

				mockPool.ExpectQuery("SELECT id, balance FROM accounts").
					WithArgs(args.id).
					WillReturnRows(rows)
				miniRedis.Close()
			},
			want: entity.Account{
				Id:      1,
				Balance: 500,
			},
			wantErr: false,
		},
		{
			name: "fail when no such account",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			mockBehaviour: func(args args, account entity.Account) {
				mockPool.ExpectQuery("SELECT id, balance FROM accounts").
					WithArgs(args.id).
					WillReturnError(errors.New("no such account"))
				miniRedis.Close()
			},
			want:    entity.Account{},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehaviour(tc.args, tc.want)

			got, err := mockAccountRepo.GetAccount(tc.args.ctx, tc.args.id)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestAccountRepo_WriteOff(t *testing.T) {
	// new mini Redis server
	miniRedis, err := miniredis.Run()
	if err != nil {
		t.Error()
	}
	client := redis.NewClient(&redis.Options{Addr: miniRedis.Addr()})
	redisCache := rediscache.New(client)

	// mock pool
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Error()
	}
	defer mockPool.Close()

	mockPostgres := &postgres.Postgres{
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		Pool:    mockPool,
	}

	mockAccountRepo := NewAccountRepo(mockPostgres, redisCache)

	type args struct {
		ctx    context.Context
		id     int
		amount int
	}

	type MockBehaviour func(args args)

	testCases := []struct {
		name          string
		args          args
		mockBehaviour MockBehaviour
		wantErr       bool
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				id:     1,
				amount: 500,
			},
			mockBehaviour: func(args args) {
				rows := mockPool.NewRows([]string{"id", "balance"}).
					AddRow(args.id, 1000)

				mockPool.ExpectQuery("SELECT id, balance FROM accounts").
					WithArgs(args.id).
					WillReturnRows(rows)

				result := pgxmock.NewResult("UPDATE", 1)

				mockPool.ExpectExec("UPDATE accounts").
					WithArgs(args.amount, args.id).
					WillReturnResult(result)

				miniRedis.Close()
			},
			wantErr: false,
		},
		{
			name: "success with exact same amount",
			args: args{
				ctx:    context.Background(),
				id:     1,
				amount: 500,
			},
			mockBehaviour: func(args args) {
				rows := mockPool.NewRows([]string{"id", "balance"}).
					AddRow(args.id, 500)

				mockPool.ExpectQuery("SELECT id, balance FROM accounts").
					WithArgs(args.id).
					WillReturnRows(rows)

				result := pgxmock.NewResult("UPDATE", 1)

				mockPool.ExpectExec("UPDATE accounts").
					WithArgs(args.amount, args.id).
					WillReturnResult(result)

				miniRedis.Close()
			},
			wantErr: false,
		},
		{
			name: "fail when not enough money",
			args: args{
				ctx:    context.Background(),
				id:     1,
				amount: 500,
			},
			mockBehaviour: func(args args) {
				rows := mockPool.NewRows([]string{"id", "balance"}).
					AddRow(args.id, 400)

				mockPool.ExpectQuery("SELECT id, balance FROM accounts").
					WithArgs(args.id).
					WillReturnRows(rows)

				miniRedis.Close()
			},
			wantErr: true,
		},
		{
			name: "fail when wrong id",
			args: args{
				ctx:    context.Background(),
				id:     1,
				amount: 500,
			},
			mockBehaviour: func(args args) {
				// write a command for a called function (getAccount)
				mockPool.ExpectQuery("SELECT id, balance FROM accounts").
					WithArgs(args.id).
					WillReturnError(errors.New("no such account"))

				// no sense to update when getAccount returned an error
				miniRedis.Close()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehaviour(tc.args)

			err := mockAccountRepo.WriteOff(tc.args.ctx, tc.args.id, tc.args.amount)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestAccountRepo_TransferMoney(t *testing.T) {
	miniRedis, err := miniredis.Run()
	if err != nil {
		t.Error()
	}

	client := redis.NewClient(&redis.Options{Addr: miniRedis.Addr()})
	redisCache := rediscache.New(client)

	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Error()
	}

	mockPostgres := &postgres.Postgres{
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		Pool:    mockPool,
	}

	accountRepo := NewAccountRepo(mockPostgres, redisCache)

	type args struct {
		ctx    context.Context
		idFrom int
		idTo   int
		amount int
	}

	type MockBehaviour func(args args)

	testCases := []struct {
		name          string
		args          args
		mockBehaviour MockBehaviour
		wantErr       bool
	}{
		{
			name: "OK",
			args: args{
				ctx:    context.Background(),
				idFrom: 1,
				idTo:   2,
				amount: 500,
			},
			mockBehaviour: func(args args) {
				rows := mockPool.NewRows([]string{"id", "balance"}).
					AddRow(args.idFrom, 1000)

				mockPool.ExpectQuery("SELECT id, balance FROM accounts").
					WithArgs(args.idFrom).
					WillReturnRows(rows)

				rows = mockPool.NewRows([]string{"id", "balance"}).AddRow(args.idTo, 200)
				mockPool.ExpectQuery("SELECT id, balance FROM accounts").
					WithArgs(args.idTo).
					WillReturnRows(rows)

				mockPool.ExpectBegin()

				result := pgxmock.NewResult("UPDATE", 1)
				mockPool.ExpectExec("UPDATE accounts").
					WithArgs(args.amount, args.idFrom).WillReturnResult(result)

				result = pgxmock.NewResult("UPDATE", 1)
				mockPool.ExpectExec("UPDATE accounts").
					WithArgs(args.amount, args.idTo).WillReturnResult(result)

				mockPool.ExpectCommit()

				miniRedis.Close()
			},
		},
		{
			name: "Failure not enough money",
			args: args{
				ctx:    context.Background(),
				idFrom: 1,
				idTo:   2,
				amount: 500,
			},
			mockBehaviour: func(args args) {
				rows := mockPool.NewRows([]string{"id", "balance"}).
					AddRow(args.idFrom, 499)

				mockPool.ExpectQuery("SELECT id, balance FROM accounts").
					WithArgs(args.idFrom).
					WillReturnRows(rows)

				rows = mockPool.NewRows([]string{"id", "balance"}).AddRow(args.idTo, 200)
				mockPool.ExpectQuery("SELECT id, balance FROM accounts").
					WithArgs(args.idTo).
					WillReturnRows(rows)

				miniRedis.Close()
			},
			wantErr: true,
		},
		{
			name: "Failure incorrect input",
			args: args{
				ctx:    context.Background(),
				idFrom: 1,
				idTo:   2,
				amount: 0,
			},
			mockBehaviour: func(args args) {},
			wantErr:       true,
		},
		{
			name: "Failure incorrect id",
			args: args{
				ctx:    context.Background(),
				idFrom: 5,
				idTo:   2,
				amount: 100,
			},
			mockBehaviour: func(args args) {
				mockPool.ExpectQuery("SELECT id, balance FROM accounts").
					WithArgs(args.idFrom).
					WillReturnError(errors.New("no such account"))
			},
			wantErr: true,
		},
		{
			name: "Failure in 1st insert",
			args: args{
				ctx:    context.Background(),
				idFrom: 1,
				idTo:   2,
				amount: 500,
			},
			mockBehaviour: func(args args) {
				rows := mockPool.NewRows([]string{"id", "balance"}).
					AddRow(args.idFrom, 899)

				mockPool.ExpectQuery("SELECT id, balance FROM accounts").
					WithArgs(args.idFrom).
					WillReturnRows(rows)

				rows = mockPool.NewRows([]string{"id", "balance"}).AddRow(args.idTo, 200)
				mockPool.ExpectQuery("SELECT id, balance FROM accounts").
					WithArgs(args.idTo).
					WillReturnRows(rows)

				mockPool.ExpectBegin()

				mockPool.ExpectExec("UPDATE accounts").
					WithArgs(args.amount, args.idFrom).WillReturnError(errors.New("something went wrong"))

				mockPool.ExpectRollback()

				miniRedis.Close()
			},
			wantErr: true,
		},
		{
			name: "Failure in 2nd insert",
			args: args{
				ctx:    context.Background(),
				idFrom: 1,
				idTo:   2,
				amount: 500,
			},
			mockBehaviour: func(args args) {
				rows := mockPool.NewRows([]string{"id", "balance"}).
					AddRow(args.idFrom, 899)

				mockPool.ExpectQuery("SELECT id, balance FROM accounts").
					WithArgs(args.idFrom).
					WillReturnRows(rows)

				rows = mockPool.NewRows([]string{"id", "balance"}).AddRow(args.idTo, 200)
				mockPool.ExpectQuery("SELECT id, balance FROM accounts").
					WithArgs(args.idTo).
					WillReturnRows(rows)

				mockPool.ExpectBegin()

				result := pgxmock.NewResult("UPDATE", 1)
				mockPool.ExpectExec("UPDATE accounts").
					WithArgs(args.amount, args.idFrom).
					WillReturnResult(result)

				mockPool.ExpectExec("UPDATE accounts").
					WithArgs(args.amount, args.idTo).
					WillReturnError(errors.New("something went wrong"))

				mockPool.ExpectRollback()

				miniRedis.Close()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehaviour(tc.args)

			err := accountRepo.TransferMoney(tc.args.ctx, tc.args.idFrom, tc.args.idTo, tc.args.amount)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}

}
