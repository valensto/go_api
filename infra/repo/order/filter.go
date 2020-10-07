package order

import (
	"time"

	mongorepo "github.com/valensto/api_apbp/infra/repo/mongo"
	"github.com/valensto/api_apbp/pkg/filter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func listPipe(uid primitive.ObjectID, f filter.Query) mongo.Pipeline {
	var pipeline mongo.Pipeline

	if uid != primitive.NilObjectID {
		pipeline = append(pipeline, bson.D{primitive.E{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: uid}}}})
		pipeline = populatePipeline(pipeline, f.Populate)
		return pipeline
	}

	pipeline = searchTerm(pipeline, f.Term)

	// pipeline = recoveryPipeline(pipeline, f.Interval)
	pipeline = populatePipeline(pipeline, f.Populate)
	pipeline = mongorepo.PaginatePipeline(pipeline, f.Pagination)

	return pipeline
}

func forecastPipe(f filter.Query, confirm bool) mongo.Pipeline {
	var pipeline mongo.Pipeline

	if confirm {
		match := bson.D{primitive.E{
			Key:   "$match",
			Value: bson.M{"status": "confirm"},
		}}
		pipeline = append(pipeline, match)
	}

	rangeStage := bson.D{primitive.E{Key: "$match", Value: bson.M{
		"recovery_at": bson.M{
			"$gte": f.Range.Start,
			"$lte": f.Range.End,
		},
	}}}

	pipeline = append(pipeline, rangeStage)

	unwindStage := bson.D{primitive.E{
		Key:   "$unwind",
		Value: "$products",
	}}

	pipeline = append(pipeline, unwindStage)

	groupStage := bson.D{primitive.E{
		Key: "$group",
		Value: bson.D{
			primitive.E{
				Key: "_id",
				Value: bson.D{
					primitive.E{
						Key:   "ref",
						Value: "$products.ref",
					},
					primitive.E{
						Key:   "name",
						Value: "$products.name",
					},
				},
			},
			primitive.E{
				Key: "quantity",
				Value: bson.D{primitive.E{
					Key: "$sum",
					Value: bson.D{primitive.E{
						Key: "$ceil",
						Value: bson.D{primitive.E{
							Key: "$cond",
							Value: bson.D{
								primitive.E{
									Key: "if",
									Value: bson.D{primitive.E{
										Key: "$eq",
										Value: bson.A{
											"$products.unit",
											"gr",
										},
									}},
								},
								primitive.E{
									Key:   "then",
									Value: "$products.quantity",
								},
								primitive.E{
									Key: "else",
									Value: bson.D{primitive.E{
										Key: "$multiply",
										Value: bson.A{
											"$products.quantity",
											"$products.auw",
										},
									}},
								},
							},
						}},
					}},
				}},
			},
		},
	}}

	pipeline = append(pipeline, groupStage)

	sortStage := bson.D{primitive.E{
		Key: "$sort",
		Value: bson.D{primitive.E{
			Key:   "quantity",
			Value: -1,
		}},
	}}

	pipeline = append(pipeline, sortStage)

	return pipeline
}

func recoveryPipeline(pipeline mongo.Pipeline, i *filter.Range) mongo.Pipeline {
	if i == nil {
		return pipeline
	}

	return append(pipeline, bson.D{primitive.E{Key: "$match", Value: bson.M{
		"recovery_at": bson.M{
			"$gte": time.Now(),
			"$lte": i.End,
		},
	}}})
}

func populatePipeline(pipeline mongo.Pipeline, populate bool) mongo.Pipeline {
	if !populate {
		return pipeline
	}

	ds := []bson.D{
		{primitive.E{Key: "$lookup", Value: bson.D{primitive.E{Key: "from", Value: "users"}, primitive.E{Key: "localField", Value: "relationShip.customer"}, primitive.E{Key: "foreignField", Value: "_id"}, primitive.E{Key: "as", Value: "relationShip.included.customer"}}}},
		{primitive.E{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$relationShip.included.customer"}, primitive.E{Key: "preserveNullAndEmptyArrays", Value: false}}}},
		{primitive.E{Key: "$lookup", Value: bson.D{primitive.E{Key: "from", Value: "users"}, primitive.E{Key: "localField", Value: "relationShip.editor"}, primitive.E{Key: "foreignField", Value: "_id"}, primitive.E{Key: "as", Value: "relationShip.included.editor"}}}},
		{primitive.E{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$relationShip.included.editor"}, primitive.E{Key: "preserveNullAndEmptyArrays", Value: false}}}},
	}

	for _, d := range ds {
		pipeline = append(pipeline, d)
	}

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
						Key: "ref",
						Value: bson.M{
							"$regex":   str,
							"$options": "i",
						},
					}},
					bson.D{primitive.E{
						Key: "status",
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
