package product

import (
	"time"

	"github.com/valensto/api_apbp/pkg/filter"
	"github.com/valensto/api_apbp/pkg/pagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product structure representation
type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt   time.Time          `bson:"created_at"`
	ModifiedAt  time.Time          `bson:"modified_at"`
	Ref         string             `bson:"ref"`
	Name        string             `bson:"name"`
	Category    string             `bson:"category,omitempty"`
	Description string             `bson:"description,omitempty"`
	AUW         float32            `bson:"auw"`
}

type Category struct {
	Category map[string]string `bson:"_id"`
	Products []Product         `bson:"products"`
}

// PDB represents product repository interface
type PDB interface {
	Migrate() error

	Categories(f filter.Query) ([]Category, error)
	Read(id string) (Product, error)
	Delete(id string) error
	List(f filter.Query) (pagination.Meta, []Product, error)
	Create(s Product) error
	UpdateFields(id string, updPct Product) (Product, error)
}
