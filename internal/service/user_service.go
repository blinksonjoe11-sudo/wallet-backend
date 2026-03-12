package service

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"wallet-api/internal/auth"
	"wallet-api/internal/model"
	"wallet-api/internal/repo"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo      *repo.UserRepo
	wallet    *repo.WalletRepo
	jwtSecret string
}

func NewUserService(repo *repo.UserRepo, wallet *repo.WalletRepo, jwtSecret string) *UserService {
	return &UserService{
		repo:      repo,
		jwtSecret: jwtSecret,
		wallet:    wallet,
	}
}

type AuthResult struct {
	Token string           `json:"token"`
	User  model.PublicUser `json:"user"`
}

func (svc *UserService) RegisterService(input model.UserRegisterInput) (model.User, error) {

	email := strings.ToLower(strings.TrimSpace(input.Email))
	password := strings.ToLower(strings.TrimSpace(input.Password))
	name := strings.ToLower(strings.TrimSpace(input.Name))

	if email == "" || password == "" || name == "" {
		return model.User{}, errors.New("Name, Email and password must not be empty")
	}

	if len(password) < 6 {
		return model.User{}, errors.New("Password must be of lenght 6")
	}

	// check if user already exist in the database

	if err := svc.repo.FindEmail(input); err != nil {
		return model.User{}, err
	}

	// hash password

	hash, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	// get the repo user

	u := model.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hash),
		Role:         "user",
	}

	u, err := svc.repo.Register(u)

	if err != nil {
		return model.User{}, fmt.Errorf("Unable to register user: %v", err)
	}

	// create wallet when user signup

	w := model.Wallet{
		UserID:    u.ID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err = svc.wallet.CreateWallet(&w)

	if err != nil {
		return model.User{}, fmt.Errorf("Unable to create wallet: %v", err)
	}

	return u, nil

}

func (svc *UserService) Login(input model.UserLoginInput) (AuthResult, error) {

	//compare hash from db

	u, _ := svc.repo.FindUserEmail(input)

	// if err != nil {
	// 	return AuthResult{}, errors.New("Email a")
	// }

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(input.Password)); err != nil {
		return AuthResult{}, fmt.Errorf("Email/Password Incorrect: %v", err)
	}

	// generate jwt token

	token, err := auth.CreateToken(svc.jwtSecret, u.ID.String(), u.Role)

	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		Token: token,
		User:  *model.ToPublic(u),
	}, nil

}
