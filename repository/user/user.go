package user

import (
	"context"
	"database/sql"
	"github.com/aaanger/graphql-test/graph/model"
	"github.com/aaanger/graphql-test/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

//go:generate mockery --name=IUserRepository

type IUserRepository interface {
	Register(ctx context.Context, req *model.RegisterReq) (*model.User, string, error)
	Login(ctx context.Context, req *model.LoginReq) (*model.User, string, error)
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Register(ctx context.Context, req *model.RegisterReq) (*model.User, string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	passwordHash := string(hashedBytes)

	user := model.User{
		Email:    strings.ToLower(req.Email),
		Username: req.Username,
		Password: passwordHash,
	}

	row := r.db.QueryRowContext(ctx, `INSERT INTO users (email, username, password_hash) VALUES($1, $2, $3) RETURNING id;`, req.Email, req.Username, passwordHash)

	err = row.Scan(&user.ID)
	if err != nil {
		return nil, "", err
	}

	accessToken, err := jwt.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return &user, accessToken, nil
}

func (r *UserRepository) Login(ctx context.Context, req *model.LoginReq) (*model.User, string, error) {
	user := model.User{
		Email: strings.ToLower(req.Email),
	}

	row := r.db.QueryRowContext(ctx, `SELECT id, username, password_hash FROM users WHERE email = $1;`, req.Email)
	err := row.Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return nil, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, "", err
	}

	accessToken, err := jwt.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return &user, accessToken, nil
}
