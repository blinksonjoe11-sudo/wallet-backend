package repo

import (
	"errors"
	"fmt"
	"wallet-api/internal/app"
	"wallet-api/internal/model"
)

type UserRepo struct {
	db *app.App
}

func NewUserRepo(db *app.App) *UserRepo {

	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) Register(u model.User) (model.User, error) {

	result := r.db.DB.Create(&u)

	if result.Error != nil {
		return model.User{}, fmt.Errorf("error registering user:%v", result.Error)
	}

	return u, nil

}

func (r *UserRepo) FindEmail(input model.UserRegisterInput) error {

	var count int64

	r.db.DB.Model(&model.User{}).Where("email = ?", input.Email).Count(&count)

	if count > 0 {
    
    return errors.New("User already exist with this email, please use another one")
	}

	return nil
}

func (r *UserRepo) FindUserEmail(input model.UserLoginInput) (model.User, error) {

	var user model.User

	result := r.db.DB.Where("email = ?", input.Email).First(&user)

	if result.Error != nil {
    
    	return model.User{}, errors.New("User already exist with this email, please use another one")
	}

	return user, nil
}


