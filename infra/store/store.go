package store

import (
	"context"

	config "github.com/valensto/api_apbp/configs"
	"github.com/valensto/api_apbp/infra/repo/order"
	"github.com/valensto/api_apbp/infra/repo/product"
	"github.com/valensto/api_apbp/infra/repo/user"
	"go.mongodb.org/mongo-driver/mongo"
)

// DBStore structure representation
type DBStore struct {
	Config config.DB
	Client *mongo.Client
	Ctx    context.Context
	DB     *mongo.Database
}

func New(config config.DB) DBStore {
	return DBStore{
		Config: config,
	}
}

// Store interface representation
type Store interface {
	Open() error
	Close() error

	BindBD(n string) error

	User() user.UDB
	Product() product.PDB
	Order() order.ODB
}
