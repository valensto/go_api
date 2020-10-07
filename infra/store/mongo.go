package store

import (
	"context"
	"fmt"
	"time"

	"github.com/valensto/api_apbp/infra/repo/order"
	"github.com/valensto/api_apbp/infra/repo/product"
	"github.com/valensto/api_apbp/infra/repo/user"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BindBD bind database with server struct
func (s *DBStore) BindBD(n string) error {
	if s.Client == nil {
		return fmt.Errorf("You need to open client first")
	}
	s.DB = s.Client.Database(n)
	return nil
}

// Open connect client to mongo
func (s *DBStore) Open() error {
	URI := fmt.Sprintf("mongodb://%v:%v@%v:%v", s.Config.Username, s.Config.Password, s.Config.Host, s.Config.Port)
	client, err := mongo.NewClient(options.Client().ApplyURI(URI))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("err connect: got=%w", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("err ping: got=%w", err)
	}

	s.Client = client
	s.Ctx = context.TODO()
	return nil
}

// Close disconnect client to mongo
func (s DBStore) Close() error {
	s.Client.Disconnect(s.Ctx)
	return nil
}

// User is a representation of user repository
func (s DBStore) User() user.UDB {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	ur := user.NewRepo(ctx, s.DB)
	return ur
}

// Product is a representation of product repository
func (s DBStore) Product() product.PDB {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	pr := product.NewRepo(ctx, s.DB)
	return pr
}

// Order is a representation of product repository
func (s DBStore) Order() order.ODB {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	or := order.NewRepo(ctx, s.DB)
	return or
}
