package repo

import (
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"testing"
	"user-balance-service/internal/entity"
	"user-balance-service/pkg/postgres"
)

func TestAuthRepo_CreateUser(t *testing.T) {
	// create mock
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Error()
	}
	defer mockPool.Close()

	mockPostgres := &postgres.Postgres{
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		Pool:    mockPool,
	}

	mockAuthRepo := NewAuthRepo(mockPostgres)

	rows := pgxmock.NewRows([]string{"id"}).AddRow(1)
	mockPool.ExpectQuery("INSERT INTO users").WillReturnRows(rows)

	id, err := mockAuthRepo.CreateUser(context.Background(), entity.User{Username: "qwe", Password: "qwe"})
	assert.NoError(t, err)
	assert.Equal(t, 1, id)

	err = mockPool.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAuthRepo_GetUser(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	mockPostgres := &postgres.Postgres{
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		Pool:    mockPool,
	}

	mockAuthRepo := NewAuthRepo(mockPostgres)

	type args struct {
		ctx          context.Context
		username     string
		passwordHash string
	}

	type MockBehaviour func(args args, user entity.User)

	testCases := []struct {
		name          string // test name
		args          args   // take arguments
		mockBehaviour MockBehaviour
		want          entity.User
		wantErr       bool
	}{
		{
			name: "success",
			args: args{
				ctx:          context.Background(),
				username:     "qwe",
				passwordHash: "qwe1",
			},
			mockBehaviour: func(args args, user entity.User) {
				rows := pgxmock.NewRows([]string{"id", "username", "password_hash"}).
					AddRow(user.Id, user.Username, user.Password)

				mockPool.ExpectQuery("SELECT id, username, password_hash FROM users").
					WithArgs(args.username, args.passwordHash).
					WillReturnRows(rows)
			},
			want: entity.User{
				Id:       1,
				Username: "qwe",
				Password: "qwe1",
			},
			wantErr: false,
		},
		{
			name: "fail",
			args: args{
				ctx:          context.Background(),
				username:     "qwe",
				passwordHash: "qwe1",
			},
			mockBehaviour: func(args args, user entity.User) {
				mockPool.ExpectQuery("SELECT id, username, password_hash FROM users").
					WithArgs(args.username, args.passwordHash).
					WillReturnError(errors.New("no such user"))
			},
			want:    entity.User{},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehaviour(tc.args, tc.want)

			got, err := mockAuthRepo.GetUser(tc.args.ctx, tc.args.username, tc.args.passwordHash)
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
