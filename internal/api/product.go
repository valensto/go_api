package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/valensto/api_apbp"
	"github.com/valensto/api_apbp/internal/formator"
	"github.com/valensto/api_apbp/internal/repo/product"
	"github.com/valensto/api_apbp/pkg/filter"
	"github.com/valensto/api_apbp/pkg/pagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Server) listProduct() http.HandlerFunc {
	type response struct {
		Meta  pagination.Meta     `json:"meta"`
		Data  []formator.JsonData `json:"data"`
		Links map[string]string   `json:"links"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		f := filter.ParseQuery(r.URL.RequestURI())
		meta, products, err := s.Store.Product().List(f)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "listing-products", err)
			return
		}

		var jsonProducts = make([]formator.JsonData, len(products))
		for i, p := range products {
			jsonProducts[i] = formator.NewJSONData("products", p.ID.Hex(), api_apbp.MapProductToJSON(&p))
		}

		resp := response{
			Meta:  meta,
			Data:  jsonProducts,
			Links: f.Pagination.GetLinks(r.URL.RequestURI(), meta.TotalElements),
		}

		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *Server) listCategory() http.HandlerFunc {
	type response struct {
		Data []formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		f := filter.ParseQuery(r.URL.RequestURI())

		categories, err := s.Store.Product().Categories(f)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "listing-categories", err)
			return
		}

		var jsonCategories = make([]formator.JsonData, len(categories))
		for i, c := range categories {
			jsonCategories[i] = formator.NewJSONData("category", "", api_apbp.MapCategoryToJSON(&c))
		}

		resp := response{
			Data: jsonCategories,
		}

		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *Server) createProduct() http.HandlerFunc {
	type response struct {
		Data formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := api_apbp.JsonProduct{}
		err := s.decode(w, r, &req)
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "decoding-product", err)
			return
		}

		fmtErrs, err := s.validateStruct(r, req)
		if len(fmtErrs) > 0 {
			s.respondErr(w, r, http.StatusBadRequest, "product-json-validation", fmtErrs)
			return
		}
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "product-json-validation", err)
			return
		}

		p := product.Product{
			ID:          primitive.NewObjectID(),
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
			Ref:         strings.ToUpper(req.Ref),
			Name:        req.Name,
			Category:    req.Category,
			Description: req.Description,
			AUW:         req.AUW,
		}

		err = s.Store.Product().Create(p)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "creating-product", err)
			return
		}

		resp := response{
			Data: formator.NewJSONData("products", p.ID.Hex(), api_apbp.MapProductToJSON(&p)),
		}
		s.respond(w, r, http.StatusCreated, resp)
	}
}

func (s *Server) updateProduct() http.HandlerFunc {
	type request struct {
		Ref         string  `json:"ref" validate:"required,len=8,ref"`
		Name        string  `json:"name" validate:"required"`
		Category    string  `json:"category,omitempty"`
		Description string  `json:"description,omitempty"`
		AUW         float32 `json:"average_unit_weight" validate:"required,numeric"`
	}

	type response struct {
		Data formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uid := s.getParam(r, "id")
		req := request{}

		ps := s.Store.Product()

		err := s.decode(w, r, &req)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "decoding-product", err)
			return
		}

		fmtErrs, err := s.validateStruct(r, req)
		if len(fmtErrs) > 0 {
			s.respondErr(w, r, http.StatusBadRequest, "product-json-validation", fmtErrs)
			return
		}
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "product-json-validation", err)
			return
		}

		p := product.Product{
			Ref:         strings.ToUpper(req.Ref),
			Name:        req.Name,
			Category:    req.Category,
			Description: req.Description,
			AUW:         req.AUW,
		}

		up, err := ps.UpdateFields(uid, p)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "updating-product", err)
			return
		}

		resp := response{
			Data: formator.NewJSONData("products", up.ID.Hex(), api_apbp.MapProductToJSON(&up)),
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *Server) getProduct() http.HandlerFunc {

	type response struct {
		Data formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uid := s.getParam(r, "id")

		p, err := s.Store.Product().Read(uid)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "reading-product", err)
			return
		}

		resp := response{
			Data: formator.NewJSONData("products", p.ID.Hex(), api_apbp.MapProductToJSON(&p)),
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *Server) deleteProduct() http.HandlerFunc {
	type response struct {
		Data formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uid := s.getParam(r, "id")

		err := s.Store.Product().Delete(uid)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "deleting-product", err)
			return
		}

		resp := response{
			Data: formator.NewJSONData("products", uid, nil),
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}
