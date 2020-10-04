package order

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var jsonSchema = bson.M{
	"bsonType": "object",
	"required": []string{"ref", "created_at", "recovery_at", "products", "status"},
	"properties": bson.M{
		"ref": bson.M{
			"bsonType":    "string",
			"description": "must be a string and is required",
		},
		"created_at": bson.M{
			"bsonType":    "date",
			"description": "must be a date and is required",
		},
		"modified_at": bson.M{
			"bsonType":    "date",
			"description": "must be a date and is required",
		},
		"recovery_at": bson.M{
			"bsonType":    "date",
			"description": "must be a string and is required",
		},
		"relationShip": bson.M{
			"bsonType":    "object",
			"description": "must be an object",
			"required":    []string{"customer", "editor"},
			"properties": bson.M{
				"customer": bson.M{
					"bsonType":    "objectId",
					"description": "must be a objectId and is required",
				},
				"editor": bson.M{
					"bsonType":    "objectId",
					"description": "must be a objectId and is required",
				},
			},
		},
		"products": bson.M{
			"bsonType":    "array",
			"description": "must be a string and is required",
			"minItems":    1,
			"uniqueItems": true,
			"items": bson.M{
				"bsonType":    "object",
				"description": "must be an object and is required",
				"required":    []string{"quantity", "unit", "ref", "name", "auw"},
				"properties": bson.M{
					"quantity": bson.M{
						"bsonType":    "double",
						"description": "must be a string and is required",
					},
					"unit": bson.M{
						"enum":        []string{"gr", "p"},
						"description": "must be a string and is required",
					},
					"ref": bson.M{
						"bsonType":    "string",
						"description": "must be a string and is required",
					},
					"name": bson.M{
						"bsonType":    "string",
						"description": "must be a string and is required",
					},
					"auw": bson.M{
						"bsonType":    "double",
						"description": "must be a string and is required",
					},
				},
			},
		},
		"status": bson.M{
			"enum":        []string{"waiting", "confirm", "ready", "delivered"},
			"description": "must be a string and is required",
		},
	},
}

var validator = bson.M{
	"$jsonSchema": jsonSchema,
}

// Migrate create product collection with schema and indexs
func (r *Repo) Migrate() error {
	opts := options.CreateCollection().SetValidator(validator)
	if err := r.db.CreateCollection(r.ctx, "orders", opts); err != nil {
		return err
	}

	_, err := r.db.Collection("orders").Indexes().CreateOne(r.ctx, mongo.IndexModel{
		Keys:    bson.M{"ref": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	return nil
}
