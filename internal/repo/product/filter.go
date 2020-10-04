package product

import (
	mongorepo "github.com/valensto/api_apbp/internal/repo/mongo"
	"github.com/valensto/api_apbp/pkg/filter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func listPipe(f filter.Query) mongo.Pipeline {
	var pipeline mongo.Pipeline

	pipeline = searchTerm(pipeline, f.Term)
	pipeline = mongorepo.PaginatePipeline(pipeline, f.Pagination)

	return pipeline
}

func categoryPipe(f filter.Query) mongo.Pipeline {
	var pipeline mongo.Pipeline

	groupStage := bson.D{primitive.E{
		Key: "$group",
		Value: bson.D{
			primitive.E{
				Key: "_id",
				Value: bson.D{primitive.E{
					Key:   "name",
					Value: "$category",
				}},
			},
			primitive.E{
				Key: "products",
				Value: bson.D{primitive.E{
					Key:   "$addToSet",
					Value: "$$ROOT",
				}},
			},
		},
	}}

	pipeline = append(pipeline, groupStage)

	sortStage := bson.D{primitive.E{
		Key: "$sort",
		Value: bson.D{primitive.E{
			Key:   "name",
			Value: 1,
		}},
	}}

	pipeline = append(pipeline, sortStage)

	return pipeline
}

func searchTerm(pipeline mongo.Pipeline, str string) mongo.Pipeline {
	if str != "" {
		pipeline = append(pipeline, bson.D{primitive.E{
			Key: "$match",
			Value: bson.D{primitive.E{
				Key: "$or",
				Value: bson.A{
					bson.D{primitive.E{
						Key: "name",
						Value: bson.M{
							"$regex":   str,
							"$options": "i",
						},
					}},
					bson.D{primitive.E{
						Key: "ref",
						Value: bson.M{
							"$regex":   str,
							"$options": "i",
						},
					}},
					bson.D{primitive.E{
						Key: "category",
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
