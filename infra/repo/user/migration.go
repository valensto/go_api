package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var jsonSchema = bson.M{
	"bsonType": "object",
	"required": []string{"lastname", "firstname", "phone", "role"},
	"properties": bson.M{
		"lastname": bson.M{
			"bsonType":    "string",
			"description": "must be a string and is required",
		},
		"firstname": bson.M{
			"bsonType":    "string",
			"description": "must be a string and is required",
		},
		"phone": bson.M{
			"bsonType":    "string",
			"description": "must be a string and is required",
		},
		"email": bson.M{
			"bsonType":    "string",
			"description": "must be a string",
		},
		"password": bson.M{
			"bsonType":    "string",
			"description": "must be a string",
		},
		"role": bson.M{
			"enum":        []string{"admin", "customer"},
			"description": "must be a only admin or customer and is required",
		},
		"created_at": bson.M{
			"bsonType":    "date",
			"description": "must be a date and is required",
		},
		"modified_at": bson.M{
			"bsonType":    "date",
			"description": "must be a date and is required",
		},
		"deleted_at": bson.M{
			"bsonType":    "date",
			"description": "must be a date and is required",
		},
		"address": bson.M{
			"bsonType":    "object",
			"description": "must be an object",
			"required":    []string{"streetName", "number", "postcode", "city"},
			"properties": bson.M{
				"streetName": bson.M{
					"bsonType":    "string",
					"description": "must be a string and is required",
				},
				"number": bson.M{
					"bsonType":    "string",
					"description": "must be a string and is required",
				},
				"postcode": bson.M{
					"bsonType":    "string",
					"description": "must be a string and is required",
				},
				"city": bson.M{
					"bsonType":    "string",
					"description": "must be a string and is required",
				},
			},
		},
	},
}

var validator = bson.M{
	"$jsonSchema": jsonSchema,
}

// Migrate create users collection with schema and indexs
func (r *Repo) Migrate() error {
	opts := options.CreateCollection().SetValidator(validator)

	if err := r.db.CreateCollection(r.ctx, "users", opts); err != nil {
		return err
	}

	_, err := r.db.Collection("users").Indexes().CreateOne(r.ctx, mongo.IndexModel{
		Keys:    bson.M{"email": 1},
		Options: options.Index().SetUnique(true).SetSparse(true),
	})
	if err != nil {
		return err
	}

	_, err = r.db.Collection("users").Indexes().CreateOne(r.ctx, mongo.IndexModel{
		Keys:    bson.M{"phone": 1},
		Options: options.Index().SetUnique(true).SetSparse(true),
	})
	if err != nil {
		return err
	}

	u := User{
		CreatedAt: time.Now(),
		Lastname:  "admin",
		Firstname: "admin",
		Phone:     "123456789",
		Email:     "admin@exemple.com",
		Password:  "admin",
		Role:      "admin",
	}

	if err := r.Create(u); err != nil {
		return err
	}

	return nil
}
