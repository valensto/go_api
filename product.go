package api_apbp

import (
	"time"

	"github.com/valensto/api_apbp/infra/repo/product"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JsonProduct struct {
	ID          primitive.ObjectID `json:"-"`
	CreatedAt   time.Time          `json:"created_at,omitempty"`
	ModifiedAt  time.Time          `json:"modified_at,omitempty"`
	Ref         string             `json:"ref" validate:"required,len=8,ref"`
	Name        string             `json:"name" validate:"required"`
	Category    string             `json:"category,omitempty"`
	Description string             `json:"description,omitempty"`
	AUW         float32            `json:"average_unit_weight" validate:"required,numeric"`
}

type JsonCategory struct {
	Category map[string]string `json:"category"`
	Products []JsonProduct     `json:"product"`
}

func MapProductToJSON(p *product.Product) JsonProduct {
	return JsonProduct{
		ID:          p.ID,
		CreatedAt:   p.CreatedAt,
		ModifiedAt:  p.ModifiedAt,
		Ref:         p.Ref,
		Name:        p.Name,
		Category:    p.Category,
		Description: p.Description,
		AUW:         p.AUW,
	}
}

func MapCategoryToJSON(p *product.Category) JsonCategory {
	products := make([]JsonProduct, len(p.Products))
	for i, pr := range p.Products {
		product := JsonProduct{
			ID:          pr.ID,
			CreatedAt:   pr.CreatedAt,
			ModifiedAt:  pr.ModifiedAt,
			Ref:         pr.Ref,
			Name:        pr.Name,
			Description: pr.Description,
			AUW:         pr.AUW,
		}
		products[i] = product
	}

	return JsonCategory{
		Category: p.Category,
		Products: products,
	}
}
