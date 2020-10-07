package user

import (
	mongorepo "github.com/valensto/api_apbp/infra/repo/mongo"
	"github.com/valensto/api_apbp/pkg/filter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo"
)

func listPipe(f filter.Query, admin bool) mongo.Pipeline {
	var pipeline mongo.Pipeline

	pipeline = searchTerm(pipeline, f.Term)

	pipeline = append(pipeline, bson.D{primitive.E{Key: "$match", Value: bson.M{"delete_at": nil}}})

	pipeline = hasAdmin(pipeline, admin)
	pipeline = mongorepo.PaginatePipeline(pipeline, f.Pagination)

	return pipeline
}

func hasAdmin(pipeline mongo.Pipeline, b bool) mongo.Pipeline {
	hasAdmin := bson.D{primitive.E{Key: "$match", Value: bson.D{primitive.E{Key: "role", Value: bson.D{primitive.E{Key: "$ne", Value: "admin"}}}}}}
	if b {
		hasAdmin = bson.D{primitive.E{Key: "$match", Value: bson.D{primitive.E{Key: "role", Value: bson.D{primitive.E{Key: "$eq", Value: "admin"}}}}}}
	}

	return append(pipeline, hasAdmin)
}

func searchTerm(pipeline mongo.Pipeline, str string) mongo.Pipeline {
	if str != "" {
		pipeline = append(pipeline, bson.D{primitive.E{
			Key: "$match",
			Value: bson.D{primitive.E{
				Key: "$or",
				Value: bson.A{
					bson.D{primitive.E{
						Key: "lastname",
						Value: bson.M{
							"$regex":   str,
							"$options": "i",
						},
					}},
					bson.D{primitive.E{
						Key: "firstname",
						Value: bson.M{
							"$regex":   str,
							"$options": "i",
						},
					}},
					bson.D{primitive.E{
						Key: "email",
						Value: bson.M{
							"$regex":   str,
							"$options": "i",
						},
					}},
					bson.D{primitive.E{
						Key: "phone",
						Value: bson.M{
							"$regex":   str,
							"$options": "i",
						},
					}},
				},
			}},
		}})
	}

	return pipeline
}
