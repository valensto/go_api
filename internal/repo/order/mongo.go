package order

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

// Repo is a representation of order repository structure
type Repo struct {
	db  *mongo.Database
	ctx context.Context
	col *mongo.Collection
}

// NewRepo return a new order repository
func NewRepo(ctx context.Context, db *mongo.Database) ODB {
	r := &Repo{
		db:  db,
		ctx: ctx,
	}
	r.col = r.db.Collection("orders")
	return r
}

// List return a list of orders
func (r Repo) List(f filter.Query) (pagination.Meta, []Order, error) {
	total, orders, err := r.retrieve(primitive.NilObjectID, f)
	if err != nil {
		return total, orders, err
	}
	return total, orders, nil
}

// Read return product by id
func (r Repo) Read(id string, populate bool) (Order, error) {
	order := Order{}

	uid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return order, repo.ErrRepoOp{
			Op:   "parsing-order-id",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("id format is not correct. got=%w", err),
		}
	}

	_, orders, err := r.retrieve(uid, filter.Query{Populate: populate})
	if err != nil {
		return order, err
	}

	if len(orders) <= 0 {
		return order, repo.ErrRepoOp{
			Op:   "retrieving-order",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("order not found. id=%v doesn't exist", id),
		}
	}
	return orders[0], nil

}

// Forecast calculate product quantity needed
func (r Repo) Forecast(f filter.Query, confirm bool) ([]Forecast, error) {
	var fs []Forecast

	curs, err := r.col.Aggregate(r.ctx, forecastPipe(f, confirm))
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

func (r Repo) retrieve(uid primitive.ObjectID, f filter.Query) (pagination.Meta, []Order, error) {
	res := struct {
		Orders []Order                  `bson:"data"`
		Meta   []map[string]interface{} `bson:"meta"`
	}{}

	meta := pagination.Meta{}

	curs, err := r.col.Aggregate(r.ctx, listPipe(uid, f))
	if err != nil {
		return meta, res.Orders, repo.ErrRepoOp{
			Op:   "order-aggregation",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during order aggregation. got=%w", err),
		}
	}

	if uid != primitive.NilObjectID {
		if err = curs.All(r.ctx, &res.Orders); err != nil {
			return meta, res.Orders, repo.ErrRepoOp{
				Op:   "retrieving-order",
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("error occured during retrieving order. got=%w", err),
			}
		}
	}

	for curs.Next(r.ctx) {
		if err = curs.Decode(&res); err != nil {
			log.Println(err)
		}
	}

	if err := curs.Err(); err != nil {
		return meta, res.Orders, repo.ErrRepoOp{
			Op:   "retrieving-order",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during retrieving order. got=%w", err),
		}
	}

	if len(res.Meta) <= 0 {
		return meta, res.Orders, nil
	}

	meta, err = pagination.NewMeta(res.Meta[0])
	if err != nil {
		return meta, res.Orders, repo.ErrRepoOp{
			Op:   "retrieving-order",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during creating meta pagination. got=%w", err),
		}
	}

	return meta, res.Orders, nil
}

// Create order to repo
func (r Repo) Create(usr Order) error {
	_, err := r.col.InsertOne(r.ctx, usr)
	if err != nil {
		return repo.ErrRepoOp{
			Op:   "create-order",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during creating. got=%w", err),
		}
	}
	return nil
}

// UpdateFields order from repo
func (r Repo) UpdateFields(id string, upd interface{}) (Order, error) {
	update := []bson.D{
		{primitive.E{
			Key:   "$set",
			Value: upd,
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
			Op:   "updating-order",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("error occured during updating. got=%w", err),
		}
	}

	return u, nil
}

// UpdateField order from repo
func (r Repo) UpdateField(id, field string, v interface{}) (Order, error) {
	update := []bson.D{
		{primitive.E{
			Key: "$set",
			Value: bson.D{primitive.E{
				Key: field,
				Value: bson.D{primitive.E{
					Key:   "$literal",
					Value: v,
				}},
			}},
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
			Op:   "updating-order",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("error occured during updating. got=%w", err),
		}
	}

	return u, nil
}

func (r Repo) update(id string, update []bson.D) (Order, error) {
	var o Order

	uid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return o, repo.ErrRepoOp{
			Op:   "parsing-order-id",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("id format is not correct. got=%w", err),
		}
	}

	filter := bson.M{"_id": bson.M{"$eq": uid}, "deleted_at": nil}

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
		return o, repo.ErrRepoOp{
			Op:   "updating-order",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("error occured during updating. got=%w", res.Err()),
		}
	}

	if err = res.Decode(&o); err != nil {
		return o, repo.ErrRepoOp{
			Op:   "updating-order",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("error occured during updating. got=%w", err),
		}
	}

	return o, nil
}

// Delete order by id
func (r Repo) Delete(id string) error {
	uid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return repo.ErrRepoOp{
			Op:   "parsing-order-id",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("id format is not correct. got=%w", err),
		}
	}

	if _, err := r.col.DeleteOne(r.ctx, bson.M{"_id": uid}); err != nil {
		return repo.ErrRepoOp{
			Op:   "deleting-order",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during deleting order. got=%w", err),
		}
	}
	return nil
}
