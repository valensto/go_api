package order

import (
	"time"

	"github.com/valensto/api_apbp/internal/repo/user"
	"github.com/valensto/api_apbp/pkg/filter"
	"github.com/valensto/api_apbp/pkg/pagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Order structure representation
type Order struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt     time.Time          `bson:"created_at"`
	ModifiedAt    time.Time          `bson:"modified_at"`
	Ref           string             `bson:"ref"`
	RecoveryAt    time.Time          `bson:"recovery_at"`
	RelationShip  RelationShip       `bson:"relationShip"`
	ProductsLines []ProductLine      `bson:"products"`
	Status        string             `bson:"status"`
}

// RelationShip structure representation
type RelationShip struct {
	Customer primitive.ObjectID `bson:"customer"`
	Editor   primitive.ObjectID `bson:"editor"`
	Included *Included          `bson:"included,omitempty"`
}

// Included structure representation
type Included struct {
	Customer user.User `bson:"customer"`
	Editor   user.User `bson:"editor"`
}

// ProductLine structure representation
type ProductLine struct {
	Quantity float32 `bson:"quantity"`
	Unit     string  `bson:"unit"`
	Ref      string  `bson:"ref"`
	Name     string  `bson:"name"`
	AUW      float32 `bson:"auw"`
}

type ForecastProduct struct {
	Name string `bson:"name"`
	Ref  string `bson:"ref"`
}

type Forecast struct {
	Product  ForecastProduct `bson:"_id"`
	Quantity int             `bson:"quantity"`
}

// ODB represents order repository interface
type ODB interface {
	Migrate() error

	Forecast(f filter.Query, confirm bool) ([]Forecast, error)
	Read(id string, populate bool) (Order, error)
	Delete(id string) error
	List(f filter.Query) (pagination.Meta, []Order, error)
	Create(s Order) error
	UpdateFields(id string, upd interface{}) (Order, error)
	UpdateField(id, field string, v interface{}) (Order, error)
}
