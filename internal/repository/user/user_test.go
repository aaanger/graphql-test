package user

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aaanger/graphql-test/internal/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

type UserRepositorySuite struct {
	suite.Suite
	repo *UserRepository
	db   *sql.DB
	mock sqlmock.Sqlmock
}

func (suite *UserRepositorySuite) SetupTest() {
	var err error
	suite.db, suite.mock, err = sqlmock.New()
	assert.NoError(suite.T(), err)
	suite.repo = NewUserRepository(suite.db)
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositorySuite))
}

// Register
// =================

func (suite *UserRepositorySuite) TestRepository_RegisterSuccess() {
	req := &model.RegisterReq{
		Email:    "test",
		Username: "test",
		Password: "test",
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	suite.mock.ExpectQuery("INSERT INTO users").WithArgs(req.Email, req.Username, sqlmock.AnyArg()).
		WillReturnRows(rows)

	user, token, err := suite.repo.Register(context.Background(), req)

	suite.NotNil(user)
	suite.NotEmpty(token)
	suite.Nil(err)
}

func (suite *UserRepositorySuite) TestRepository_RegisterEmptyFields() {
	req := &model.RegisterReq{
		Password: "test",
	}

	user, token, err := suite.repo.Register(context.Background(), req)

	suite.Nil(user)
	suite.Empty(token)
	suite.NotNil(err)
}

func (suite *UserRepositorySuite) TestRepository_RegisterPasswordHashError() {
	req := &model.RegisterReq{
		Email:    "test",
		Username: "test",
		Password: "",
	}

	user, token, err := suite.repo.Register(context.Background(), req)

	suite.Nil(user)
	suite.Empty(token)
	suite.NotNil(err)
}

// Login
// ==========================

func (suite *UserRepositorySuite) TestRepository_LoginSuccess() {
	req := &model.LoginReq{
		Email:    "test",
		Password: "test",
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	rows := sqlmock.NewRows([]string{"id", "username", "password_hash"}).AddRow(1, "test", string(hashedPassword))
	suite.mock.ExpectQuery(`SELECT (.+) FROM users WHERE (.+)`).
		WithArgs(req.Email).WillReturnRows(rows)

	user, token, err := suite.repo.Login(context.Background(), req)

	suite.NotNil(user)
	suite.NotEmpty(token)
	suite.Nil(err)
}

func (suite *UserRepositorySuite) TestRepository_LoginEmptyFields() {
	req := &model.LoginReq{}

	user, token, err := suite.repo.Login(context.Background(), req)

	suite.Nil(user)
	suite.Empty(token)
	suite.NotNil(err)
}

func (suite *UserRepositorySuite) TestRepository_InvalidEmail() {
	req := &model.LoginReq{
		Email:    "invalid",
		Password: "test",
	}

	suite.mock.ExpectQuery(`SELECT id, password_hash FROM users WHERE email=`).
		WithArgs(req.Email).
		WillReturnError(errors.New("sql: no rows in result set"))

	user, token, err := suite.repo.Login(context.Background(), req)

	suite.Nil(user)
	suite.Empty(token)
	suite.NotNil(err)
}

func (suite *UserRepositorySuite) TestRepository_LoginWrongPassword() {
	req := &model.LoginReq{
		Email:    "test",
		Password: "wrong",
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)

	rows := sqlmock.NewRows([]string{"id", "password_hash"}).AddRow(1, string(hashedPassword))
	suite.mock.ExpectQuery(`SELECT id, password_hash FROM users WHERE email=`).
		WithArgs(req.Email).
		WillReturnRows(rows)

	user, token, err := suite.repo.Login(context.Background(), req)

	suite.Nil(user)
	suite.Empty(token)
	suite.NotNil(err)
}
