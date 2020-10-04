package mongo

import (
	"github.com/valensto/api_apbp/pkg/pagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func PaginatePipeline(pipeline mongo.Pipeline, p pagination.Query) mongo.Pipeline {
	paginate := bson.D{
		primitive.E{Key: "$facet", Value: bson.D{
			primitive.E{Key: "data", Value: []bson.D{
				{primitive.E{Key: "$skip", Value: p.Skip}},
				{primitive.E{Key: "$limit", Value: p.Limit}},
			}},
			primitive.E{Key: "meta", Value: []bson.D{
				{primitive.E{Key: "$count", Value: "totalElements"}},
			}},
		}},
	}

	totalPage := bson.D{primitive.E{
		Key: "$addFields",
		Value: bson.D{primitive.E{
			Key: "meta.totalPages",
			Value: bson.D{primitive.E{
				Key: "$ceil",
				Value: bson.D{primitive.E{
					Key: "$divide",
					Value: bson.A{
						bson.D{primitive.E{
							Key: "$arrayElemAt",
							Value: bson.A{
								"$meta.totalElements",
								0,
							},
						}},
						p.Limit,
					},
				}},
			}},
		},
			primitive.E{
				Key: "meta.perPages",
				Value: bson.D{primitive.E{
					Key:   "$add",
					Value: p.Limit,
				}},
			}},
	}}

	pipeline = append(pipeline, paginate)
	pipeline = append(pipeline, totalPage)
	return pipeline
}
