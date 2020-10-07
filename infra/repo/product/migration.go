package product

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var jsonSchema = bson.M{
	"bsonType": "object",
	"required": []string{"ref", "name", "auw"},
	"properties": bson.M{
		"ref": bson.M{
			"bsonType":    "string",
			"description": "must be a string and is required",
		},
		"name": bson.M{
			"bsonType":    "string",
			"description": "must be a string and is required",
		},
		"category": bson.M{
			"bsonType":    "string",
			"description": "must be a string",
		},
		"description": bson.M{
			"bsonType":    "string",
			"description": "must be a string",
		},
		"auw": bson.M{
			"bsonType":    "number",
			"description": "must be a number and is required",
		},
		"created_at": bson.M{
			"bsonType":    "date",
			"description": "must be a date",
		},
		"modified_at": bson.M{
			"bsonType":    "date",
			"description": "must be a date",
		},
	},
}

var validator = bson.M{
	"$jsonSchema": jsonSchema,
}

// Migrate create product collection with schema and indexs
func (r *Repo) Migrate() error {
	opts := options.CreateCollection().SetValidator(validator)
	if err := r.db.CreateCollection(r.ctx, "products", opts); err != nil {
		return err
	}

	_, err := r.db.Collection("products").Indexes().CreateOne(r.ctx, mongo.IndexModel{
		Keys:    bson.M{"ref": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	return nil
}
