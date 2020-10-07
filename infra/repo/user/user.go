package user

import (
	"time"

	"github.com/valensto/api_apbp/pkg/filter"
	"github.com/valensto/api_apbp/pkg/pagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// User structure representation
type User struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt  time.Time          `bson:"created_at"`
	ModifiedAt *time.Time         `bson:"modified_at,omitempty"`
	DeletedAt  *time.Time         `bson:"deleted_at,omitempty"`
	Lastname   string             `bson:"lastname"`
	Firstname  string             `bson:"firstname"`
	Phone      string             `bson:"phone,omitempty"`
	Email      string             `bson:"email,omitempty"`
	Password   string             `bson:"password,omitempty"`
	Address    *Addr              `bson:"address,omitempty"`
	Role       string             `bson:"role"`
}

// Addr structure representation
type Addr struct {
	StreetName string `bson:"streetName"`
	Number     string `bson:"number"`
	Postcode   string `bson:"postcode"`
	City       string `bson:"city"`
}

// HashPassword return bcrypt generate from string password
func HashPassword(pwd string) (string, error) {
	cryptPwd, err := bcrypt.GenerateFromPassword([]byte(pwd), 10)
	if err != nil {
		return "", err
	}
	return string(cryptPwd), nil
}

// UDB represents user repository interface
type UDB interface {
	Migrate() error

	FindByCredential(email string) (User, error)
	Read(id string) (User, error)
	Delete(id string) error
	List(f filter.Query, admin bool) (pagination.Meta, []User, error)
	Create(s User) error
	UpdateFields(id string, updUsr interface{}) (User, error)
	UpdateField(id, field string, v interface{}) (User, error)
}
