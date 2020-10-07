package user

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/valensto/api_apbp/infra/repo"
	"github.com/valensto/api_apbp/pkg/filter"
	"github.com/valensto/api_apbp/pkg/pagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/mongo"
)

// Repo is a representation of user repository structure
type Repo struct {
	db  *mongo.Database
	col *mongo.Collection
	ctx context.Context
}

// NewRepo return a new user repository
func NewRepo(ctx context.Context, db *mongo.Database) UDB {
	r := &Repo{
		db:  db,
		ctx: ctx,
	}
	r.col = r.db.Collection("users")
	return r
}

// FindByCredential find user by his credentials
func (r Repo) FindByCredential(email string) (User, error) {
	usr := User{}

	if err := r.col.FindOne(r.ctx, bson.M{"email": email}).Decode(&usr); err != nil {
		return usr, repo.ErrRepoOp{
			Op:   "retrieving-user",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("error occured during retrieving user. got=%w", err),
		}
	}
	return usr, nil
}

// List return a list of users
func (r Repo) List(f filter.Query, admin bool) (pagination.Meta, []User, error) {
	total, users, err := r.retrieve(f, listPipe(f, admin))
	if err != nil {
		return total, users, err
	}
	return total, users, nil
}

// Read return user by id
func (r Repo) Read(id string) (User, error) {
	usr := User{}
	uid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return usr, repo.ErrRepoOp{
			Op:   "parsing-user-id",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("id format is not correct. got=%w", err),
		}
	}

	if err := r.col.FindOne(r.ctx, bson.M{"_id": uid, "delete_at": nil}).Decode(&usr); err != nil {
		return usr, repo.ErrRepoOp{
			Op:   "retrieving-user",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during retrieving user. got=%w", err),
		}
	}
	return usr, nil
}

func (r Repo) retrieve(f filter.Query, pipeline mongo.Pipeline) (pagination.Meta, []User, error) {
	res := struct {
		Users []User                   `bson:"data"`
		Meta  []map[string]interface{} `bson:"meta"`
	}{}

	meta := pagination.Meta{}

	curs, err := r.col.Aggregate(r.ctx, pipeline)
	if err != nil {
		return meta, res.Users, repo.ErrRepoOp{
			Op:   "user-aggregation",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during user aggregation. got=%w", err),
		}
	}

	for curs.Next(r.ctx) {
		if err = curs.Decode(&res); err != nil {
			log.Println(err)
		}
	}

	if err := curs.Err(); err != nil {
		return meta, res.Users, repo.ErrRepoOp{
			Op:   "retrieving-user",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during retrieving user. got=%w", err),
		}
	}

	if len(res.Meta) <= 0 {
		return meta, res.Users, nil
	}

	meta, err = pagination.NewMeta(res.Meta[0])
	if err != nil {
		return meta, res.Users, repo.ErrRepoOp{
			Op:   "retrieving-user",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during creating meta pagination. got=%w", err),
		}
	}

	return meta, res.Users, nil
}

// Create user to repo
func (r Repo) Create(usr User) error {
	if usr.Password != "" {
		pwd, err := HashPassword(usr.Password)
		if err != nil {
			return err
		}
		usr.Password = pwd
	}

	_, err := r.col.InsertOne(r.ctx, usr)
	if err != nil {
		return repo.ErrRepoOp{
			Op:   "create-user",
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("error occured during creating. got=%w", err),
		}
	}
	return nil
}

// UpdateFields user from repo
func (r Repo) UpdateFields(id string, updUsr interface{}) (User, error) {
	update := []bson.D{
		{primitive.E{
			Key:   "$set",
			Value: updUsr,
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

// UpdateField user from repo
func (r Repo) UpdateField(id, field string, v interface{}) (User, error) {
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
			Op:   "updating-user",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("error occured during updating. got=%w", err),
		}
	}

	return u, nil
}

func (r Repo) update(id string, update []bson.D) (User, error) {
	var u User

	uid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return u, repo.ErrRepoOp{
			Op:   "parsing-user-id",
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
		return u, repo.ErrRepoOp{
			Op:   "updating-user",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("error occured during updating. got=%w", res.Err()),
		}
	}

	if err = res.Decode(&u); err != nil {
		return u, repo.ErrRepoOp{
			Op:   "updating-user",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("error occured during updating. got=%w", err),
		}
	}

	return u, nil
}

// Delete user by id
func (r Repo) Delete(id string) error {
	uid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return repo.ErrRepoOp{
			Op:   "parsing-user-id",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("id format is not correct. got=%w", err),
		}
	}

	filter := bson.M{"_id": bson.M{"$eq": uid}}
	update := bson.M{
		"_id":         uid,
		"delete_at":   time.Now(),
		"modified_at": time.Now(),
		"phone":       primitive.NewObjectID().Hex(),
		"lastname":    "",
		"firstname":   "",
		"role":        "customer",
	}
	_, err = r.col.ReplaceOne(
		r.ctx,
		filter,
		update,
	)
	if err != nil {
		return repo.ErrRepoOp{
			Op:   "updating-user",
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("error occured during updating. got=%w", err),
		}
	}
	return nil
}
