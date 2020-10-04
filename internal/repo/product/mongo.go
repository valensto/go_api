package product

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/valensto/api_apbp/internal/repo"
	"github.com/valensto/api_apbp/pkg/filter"
	"github.com/valensto/api_apbp/pkg/pagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repo is a representation of product repository structure
type Repo struct {
	db  *mongo.Database
	ctx context.Context
	col *mongo.Collection
}

// NewRepo return a new product repository
func NewRepo(ctx context.Context, db *mongo.Database) PDB {
	r := &Repo{
		db:  db,
		ctx: ctx,
	}
	r.col = r.db.Collection("products")
	return r
}

// List return a list of products
func (r Repo) List(f filter.Query) (pagination.Meta, []Product, error) {
	total, products, err := r.retrieve(f)
	if err != nil {
		return total, products, err
	}
	return total, products, nil
}

// Categories func
func (r Repo) Categories(f filter.Query) ([]Category, error) {
	var fs []Category

	curs, err := r.col.Aggregate(r.ctx, categoryPipe(f))
	if err != nil {
		return fs, repo.ErrRepoOp{
			Op:   "order-aggregation",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during order aggregation. got=%w", err),
		}
	}

	if err := curs.All(r.ctx, &fs); err != nil {
		return fs, repo.ErrRepoOp{
			Op:   "order-aggregation",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during order retrieving. got=%w", err),
		}
	}

	return fs, nil
}

// Read return product by id
func (r Repo) Read(id string) (Product, error) {
	product := Product{}
	uid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return product, repo.ErrRepoOp{
			Op:   "parsing-product-id",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("id format is not correct. got=%w", err),
		}
	}

	if err := r.col.FindOne(r.ctx, bson.M{"_id": uid}).Decode(&product); err != nil {
		return product, repo.ErrRepoOp{
			Op:   "retrieving-product",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during retrieving product. got=%w", err),
		}
	}
	return product, nil
}

func (r Repo) retrieve(f filter.Query) (pagination.Meta, []Product, error) {
	resp := struct {
		Products []Product                `bson:"data"`
		Meta     []map[string]interface{} `bson:"meta"`
	}{}

	meta := pagination.Meta{}

	curs, err := r.col.Aggregate(r.ctx, listPipe(f))
	if err != nil {
		return meta, resp.Products, repo.ErrRepoOp{
			Op:   "product-aggregation",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during product aggregation. got=%w", err),
		}
	}

	for curs.Next(r.ctx) {
		if err = curs.Decode(&resp); err != nil {
			log.Println(err)
		}
	}

	if err := curs.Err(); err != nil {
		return meta, resp.Products, repo.ErrRepoOp{
			Op:   "retrieving-product",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during retrieving product. got=%w", err),
		}
	}

	if len(resp.Meta) <= 0 {
		return meta, resp.Products, nil
	}

	meta, err = pagination.NewMeta(resp.Meta[0])
	if err != nil {
		return meta, resp.Products, repo.ErrRepoOp{
			Op:   "retrieving-product",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during creating meta pagination. got=%w", err),
		}
	}

	return meta, resp.Products, nil
}

// Create product to repo
func (r Repo) Create(usr Product) error {
	_, err := r.col.InsertOne(r.ctx, usr)
	if err != nil {
		return repo.ErrRepoOp{
			Op:   "create-product",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during creating. got=%w", err),
		}
	}
	return nil
}

// UpdateFields product from repo
func (r Repo) UpdateFields(id string, updPct Product) (Product, error) {
	update := []bson.D{
		{primitive.E{
			Key:   "$set",
			Value: updPct,
		}},
		{primitive.E{
			Key: "$addFields",
			Value: bson.D{primitive.E{
				Key:   "modified_at",
				Value: time.Now(),
			}},
		}},
	}

	u, err := r.update(id, update)
	if err != nil {
		return u, repo.ErrRepoOp{
			Op:   "updating-user",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("error occured during updating. got=%w", err),
		}
	}

	return u, nil
}

func (r Repo) update(id string, update []bson.D) (Product, error) {
	var p Product

	uid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return p, repo.ErrRepoOp{
			Op:   "parsing-product-id",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("id format is not correct. got=%w", err),
		}
	}

	filter := bson.M{"_id": bson.M{"$eq": uid}}

	opts := options.FindOneAndUpdate()
	after := options.After
	opts.ReturnDocument = &after

	res := r.col.FindOneAndUpdate(
		r.ctx,
		filter,
		update,
		opts,
	)
	if res.Err() != nil {
		return p, repo.ErrRepoOp{
			Op:   "updating-product",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("error occured during updating. got=%w", res.Err()),
		}
	}

	if err = res.Decode(&p); err != nil {
		return p, repo.ErrRepoOp{
			Op:   "updating-product",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("error occured during updating. got=%w", err),
		}
	}

	return p, nil
}

// Delete product by id
func (r Repo) Delete(id string) error {
	uid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return repo.ErrRepoOp{
			Op:   "parsing-product-id",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("id format is not correct. got=%w", err),
		}
	}

	if _, err := r.col.DeleteOne(r.ctx, bson.M{"_id": uid}); err != nil {
		return repo.ErrRepoOp{
			Op:   "deleting-product",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during deleting product. got=%w", err),
		}
	}
	return nil
}
