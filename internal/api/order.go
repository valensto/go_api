package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/valensto/api_apbp"
	"github.com/valensto/api_apbp/internal/formator"
	"github.com/valensto/api_apbp/internal/repo/order"
	"github.com/valensto/api_apbp/internal/session"
	"github.com/valensto/api_apbp/pkg/filter"
	"github.com/valensto/api_apbp/pkg/pagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Server) listOrder() http.HandlerFunc {
	type response struct {
		Meta  pagination.Meta     `json:"meta"`
		Data  []formator.JsonData `json:"data"`
		Links map[string]string   `json:"links"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		f := filter.ParseQuery(r.URL.RequestURI())

		meta, orders, err := s.Store.Order().List(f)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "listing-order", err)
			return
		}

		var jsonOrders = make([]formator.JsonData, len(orders))
		for i, o := range orders {
			order := api_apbp.MapOrderToJSON(o)
			jsonOrders[i] = formator.NewJSONData("orders", o.ID.Hex(), order)
		}

		resp := response{
			Meta:  meta,
			Data:  jsonOrders,
			Links: f.Pagination.GetLinks(r.URL.RequestURI(), meta.TotalElements),
		}

		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *Server) forecast(confirm bool) http.HandlerFunc {
	type response struct {
		Meta map[string]time.Time    `json:"meta"`
		Data []api_apbp.JsonForecast `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		f := filter.ParseQuery(r.URL.RequestURI())

		t := time.Time{}
		if f.Range.End == t {
			s.respondErr(w, r, http.StatusBadRequest, "parsing-params", fmt.Errorf("end param is a required param to ended request range, ?end=2006-01-02T15:04:05.000Z"))
			return
		}

		fs, err := s.Store.Order().Forecast(f, confirm)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "listing-forecast", err)
			return
		}

		var jsonForecasts = make([]api_apbp.JsonForecast, len(fs))
		for i, f := range fs {
			jsonForecasts[i] = api_apbp.MapForecastToJson(f)
		}

		meta := map[string]time.Time{
			"from": f.Range.Start,
			"to":   f.Range.End,
		}
		resp := response{
			Meta: meta,
			Data: jsonForecasts,
		}

		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *Server) getOrder() http.HandlerFunc {

	type response struct {
		Data formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		id := s.getParam(r, "id")
		populate := r.URL.Query().Get("populate") == "1"

		order, err := s.Store.Order().Read(id, populate)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "reading-order", err)
			return
		}

		data := api_apbp.MapOrderToJSON(order)

		resp := response{
			Data: formator.NewJSONData("orders", order.ID.Hex(), data),
		}

		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *Server) createOrder() http.HandlerFunc {

	type reqProductLine struct {
		Quantity  float32            `json:"quantity,omitempty" validate:"required"`
		Unit      string             `json:"unit,omitempty" validate:"required,oneof=gr p"`
		ProductID primitive.ObjectID `json:"product_id,omitempty" validate:"required"`
	}

	type request struct {
		RecoveryAt    time.Time          `json:"recovery_at,omitempty" validate:"required"`
		Customer      primitive.ObjectID `json:"customer,omitempty" validate:"required"`
		ProductsLines []reqProductLine   `json:"products,omitempty" validate:"required,unique,min=1,dive,required"`
		Status        string             `json:"status,omitempty" validate:"required,oneof=waiting confirm ready delivered"`
	}

	type response struct {
		Data formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uIDstr, err := session.GetUserID(r.Context())
		if err != nil {
			s.respondErr(w, r, http.StatusUnauthorized, "", err)
			return
		}

		uid, err := primitive.ObjectIDFromHex(uIDstr)
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "decoding-editor", err)
			return
		}

		req := request{}
		err = s.decode(w, r, &req)
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "decoding-order", err)
			return
		}

		fmtErrs, err := s.validateStruct(r, req)
		if len(fmtErrs) > 0 {
			s.respondErr(w, r, http.StatusBadRequest, "order-json-validation", fmtErrs)
			return
		}
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "order-json-validation", err)
			return
		}

		customer, err := s.Store.User().Read(req.Customer.Hex())
		if err != nil {
			s.respondErr(w, r, 0, "", err)
			return
		}

		productLines := make([]order.ProductLine, len(req.ProductsLines))
		for i, pl := range req.ProductsLines {
			product, err := s.Store.Product().Read(pl.ProductID.Hex())
			if err != nil {
				s.respondErr(w, r, http.StatusBadRequest, "product-not-found", err)
				return
			}
			productLines[i] = order.ProductLine{
				Quantity: pl.Quantity,
				Unit:     pl.Unit,
				Ref:      product.Ref,
				Name:     product.Name,
				AUW:      product.AUW,
			}
		}

		o := order.Order{
			ID:         primitive.NewObjectID(),
			Ref:        api_apbp.GenerateRef(),
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
			RecoveryAt: req.RecoveryAt,
			RelationShip: order.RelationShip{
				Customer: customer.ID,
				Editor:   uid,
			},
			ProductsLines: productLines,
			Status:        req.Status,
		}

		err = s.Store.Order().Create(o)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "creating-order", err)
			return
		}

		order := api_apbp.MapOrderToJSON(o)

		go func() {
			if customer.Email != "" {
				mail := order.NewOrderMail(customer)
				if err := s.Mailer.Send(mail); err != nil {
					log.Println(err)
				}
			}
		}()

		resp := response{
			Data: formator.NewJSONData("orders", o.ID.Hex(), order),
		}
		s.respond(w, r, http.StatusCreated, resp)
	}
}

func (s *Server) updateOrderStatus() http.HandlerFunc {
	type request struct {
		Status string `json:"status,omitempty" validate:"required,oneof=waiting confirm ready delivered"`
	}

	type response struct {
		Data formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		id := s.getParam(r, "id")
		os := s.Store.Order()
		req := request{}

		err := s.decode(w, r, &req)
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "decoding-json", err)
			return
		}

		fmtErrs, err := s.validateStruct(r, req)
		if len(fmtErrs) > 0 {
			s.respondErr(w, r, http.StatusBadRequest, "json-validation", fmtErrs)
			return
		}
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "order-json-validation", err)
			return
		}

		o, err := os.UpdateField(id, "status", req.Status)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "updating-order", err)
			return
		}

		order := api_apbp.MapOrderToJSON(o)

		go func() {
			if order.RelationShip.Included.Customer.Email != "" && req.Status == "ready" {
				mail := order.NewStatusMail()
				if err := s.Mailer.Send(mail); err != nil {
					log.Println(err)
				}
			}
		}()

		resp := response{
			Data: formator.NewJSONData("orders", id, req),
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *Server) updateOrderRecovery() http.HandlerFunc {
	type request struct {
		Recovery time.Time `json:"recovery_at" validate:"required"`
	}

	type response struct {
		Data formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		id := s.getParam(r, "id")
		os := s.Store.Order()
		req := request{}

		err := s.decode(w, r, &req)
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "decoding-json", err)
			return
		}

		fmtErrs, err := s.validateStruct(r, req)
		if len(fmtErrs) > 0 {
			s.respondErr(w, r, http.StatusBadRequest, "json-validation", fmtErrs)
			return
		}
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "order-json-validation", err)
			return
		}

		_, err = os.UpdateField(id, "recovery_at", req.Recovery)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "updating-order", err)
			return
		}

		// order := api_apbp.MapOrderToJSON(o)

		// go func() {
		// 	if order.RelationShip.Included.Customer.Email != "" && req.Status == "ready" {
		// 		mail := order.NewStatusMail()
		// 		if err := s.Mailer.Send(mail); err != nil {
		// 			log.Println(err)
		// 		}
		// 	}
		// }()

		resp := response{
			Data: formator.NewJSONData("orders", id, req),
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *Server) updateOrderProducts() http.HandlerFunc {
	type reqProductLine struct {
		Quantity  float32            `json:"quantity,omitempty" validate:"required"`
		Unit      string             `json:"unit,omitempty" validate:"required,oneof=gr p"`
		ProductID primitive.ObjectID `json:"product_id,omitempty" validate:"required"`
	}

	type productLine struct {
		Quantity float32 `json:"quantity,omitempty"`
		Unit     string  `json:"unit,omitempty"`
		Ref      string  `json:"ref,omitempty"`
		Name     string  `json:"name,omitempty"`
		AUW      float32 `json:"auw,omitempty"`
	}

	type upd struct {
		Products []productLine
	}

	type request struct {
		ProductsLines []reqProductLine `json:"products,omitempty" validate:"required,unique,min=1,dive,required"`
	}

	type response struct {
		Data formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uid := s.getParam(r, "id")
		os := s.Store.Order()
		req := request{}

		err := s.decode(w, r, &req)
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "decoding-json", err)
			return
		}

		fmtErrs, err := s.validateStruct(r, req)
		if len(fmtErrs) > 0 {
			s.respondErr(w, r, http.StatusBadRequest, "json-validation", fmtErrs)
			return
		}
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "order-json-validation", err)
			return
		}

		productLines := make([]productLine, len(req.ProductsLines))
		for i, pl := range req.ProductsLines {
			product, err := s.Store.Product().Read(pl.ProductID.Hex())
			if err != nil {
				s.respondErr(w, r, http.StatusBadRequest, "product-not-found", err)
				return
			}
			productLines[i] = productLine{
				Quantity: pl.Quantity,
				Unit:     pl.Unit,
				Ref:      product.Ref,
				Name:     product.Name,
				AUW:      product.AUW,
			}
		}

		o := upd{
			Products: productLines,
		}

		_, err = os.UpdateField(uid, "products", o.Products)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "updating-order", err)
			return
		}

		resp := response{
			Data: formator.NewJSONData("orders", uid, req),
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *Server) deleteOrder() http.HandlerFunc {
	type response struct {
		Data formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uid := s.getParam(r, "id")

		err := s.Store.Order().Delete(uid)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "deleting-order", err)
			return
		}

		resp := response{
			Data: formator.NewJSONData("orders", uid, nil),
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}
