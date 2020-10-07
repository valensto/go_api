package api_apbp

import (
	"time"

	"github.com/valensto/api_apbp/infra/repo/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func MapUserToJSON(u user.User) JsonUser {
	addr := JsonAddr{
		StreetName: u.Address.StreetName,
		Number:     u.Address.Number,
		Postcode:   u.Address.Postcode,
		City:       u.Address.City,
	}
	return JsonUser{
		ID:         u.ID,
		CreatedAt:  u.CreatedAt,
		ModifiedAt: u.ModifiedAt,
		Lastname:   u.Lastname,
		Firstname:  u.Firstname,
		Phone:      u.Phone,
		Email:      u.Email,
		Password:   u.Password,
		Address:    &addr,
		Role:       u.Role,
	}
}

type PWDUser struct {
	Password string `json:"password"`
}

type JsonUser struct {
	ID         primitive.ObjectID `json:"-"`
	CreatedAt  time.Time          `json:"created_at"`
	ModifiedAt *time.Time         `json:"modified_at,omitempty"`
	DeletedAt  *time.Time         `json:"deleted_at,omitempty"`
	Lastname   string             `json:"lastname" validate:"required"`
	Firstname  string             `json:"firstname" validate:"required"`
	Phone      string             `json:"phone" validate:"required,phone"`
	Email      string             `json:"email,omitempty" validate:"rfe=Role:admin,omitempty,email"`
	Password   string             `json:"-" validate:"rfe=Role:admin,omitempty,pwd"`
	Address    *JsonAddr          `json:"address"`
	Role       string             `json:"role" validate:"required,oneof=admin customer"`
}

type JsonAddr struct {
	StreetName string `json:"streetName"`
	Number     string `json:"number"`
	Postcode   string `json:"postcode"`
	City       string `json:"city"`
}
