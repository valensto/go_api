package api

import (
	"net/http"
	"time"

	"github.com/valensto/api_apbp"
	"github.com/valensto/api_apbp/internal/formator"
	"github.com/valensto/api_apbp/internal/repo/user"
	"github.com/valensto/api_apbp/pkg/filter"
	"github.com/valensto/api_apbp/pkg/pagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) listUser(admin bool) http.HandlerFunc {
	type response struct {
		Meta  pagination.Meta     `json:"meta,omitempty"`
		Data  []formator.JsonData `json:"data"`
		Links map[string]string   `json:"links,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		f := filter.ParseQuery(r.URL.RequestURI())
		meta, usrs, err := s.Store.User().List(f, admin)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "listing-user", err)
			return
		}

		var jsonUsrs = make([]formator.JsonData, len(usrs))
		for i, u := range usrs {
			jsonUsrs[i] = formator.NewJSONData("users", u.ID.Hex(), api_apbp.MapUserToJSON(u))
		}

		resp := response{
			Meta:  meta,
			Data:  jsonUsrs,
			Links: f.Pagination.GetLinks(r.URL.RequestURI(), meta.TotalElements),
		}

		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *Server) searchUser() http.HandlerFunc {
	type response struct {
		Meta  pagination.Meta     `json:"meta,omitempty"`
		Data  []formator.JsonData `json:"data"`
		Links map[string]string   `json:"links,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		f := filter.ParseQuery(r.URL.RequestURI())
		meta, usrs, err := s.Store.User().List(f, false)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "listing-user", err)
			return
		}

		var jsonUsrs = make([]formator.JsonData, len(usrs))
		for i, u := range usrs {
			jsonUsrs[i] = formator.NewJSONData("users", u.ID.Hex(), api_apbp.MapUserToJSON(u))
		}

		resp := response{
			Meta:  meta,
			Data:  jsonUsrs,
			Links: f.Pagination.GetLinks(r.URL.RequestURI(), meta.TotalElements),
		}

		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *Server) getUser() http.HandlerFunc {

	type response struct {
		Data formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uid := s.getParam(r, "id")

		usr, err := s.Store.User().Read(uid)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "reading-user", err)
			return
		}

		resp := response{
			Data: formator.NewJSONData("users", usr.ID.Hex(), api_apbp.MapUserToJSON(usr)),
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *Server) createUser() http.HandlerFunc {

	type request struct {
		Lastname  string            `json:"lastname,omitempty" validate:"required"`
		Firstname string            `json:"firstname,omitempty" validate:"required"`
		Phone     string            `json:"phone,omitempty" validate:"required,phone"`
		Email     string            `json:"email,omitempty" validate:"rfe=Role:admin,omitempty,email"`
		Password  string            `json:"password,omitempty" validate:"rfe=Role:admin,omitempty,pwd"`
		Address   api_apbp.JsonAddr `json:"address,omitempty"`
		Role      string            `json:"role,omitempty" validate:"required,oneof=admin customer"`
	}

	type response struct {
		Data formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := request{}
		err := s.decode(w, r, &req)
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "decoding-user", err)
			return
		}

		fmtErrs, err := s.validateStruct(r, req)
		if len(fmtErrs) > 0 {
			s.respondErr(w, r, http.StatusBadRequest, "user-json-validation", fmtErrs)
			return
		}
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "user-json-validation", err)
			return
		}

		addr := user.Addr{
			StreetName: req.Address.StreetName,
			Number:     req.Address.Number,
			Postcode:   req.Address.Postcode,
			City:       req.Address.City,
		}

		u := user.User{
			ID:        primitive.NewObjectID(),
			CreatedAt: time.Now(),
			Lastname:  req.Lastname,
			Firstname: req.Firstname,
			Phone:     req.Phone,
			Email:     req.Email,
			Password:  req.Password,
			Address:   &addr,
			Role:      req.Role,
		}

		err = s.Store.User().Create(u)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "creating-user", err)
			return
		}

		resp := response{
			Data: formator.NewJSONData("users", u.ID.Hex(), api_apbp.MapUserToJSON(u)),
		}
		s.respond(w, r, http.StatusCreated, resp)
	}
}

func (s *Server) updateUser() http.HandlerFunc {
	type request struct {
		Lastname  string `json:"lastname" validate:"required"`
		Firstname string `json:"firstname" validate:"required"`
		Phone     string `json:"phone" validate:"required,phone"`
		Email     string `json:"email" validate:"required,email"`
	}

	type response struct {
		Data formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uid := s.getParam(r, "id")
		us := s.Store.User()

		req := request{}
		err := s.decode(w, r, &req)
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "decoding-user", err)
			return
		}

		fmtErrs, err := s.validateStruct(r, req)
		if len(fmtErrs) > 0 {
			s.respondErr(w, r, http.StatusBadRequest, "user-json-validation", fmtErrs)
			return
		}
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "user-json-validation", err)
			return
		}

		u, err := us.UpdateFields(uid, req)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "updating-user", err)
			return
		}

		resp := response{
			Data: formator.NewJSONData("users", u.ID.Hex(), api_apbp.MapUserToJSON(u)),
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *Server) updateAddress() http.HandlerFunc {
	type request struct {
		StreetName string `json:"streetName" validate:"required"`
		Number     string `json:"number,omitempty"`
		Postcode   string `json:"postcode" validate:"required"`
		City       string `json:"city" validate:"required"`
	}

	type response struct {
		Data formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uid := s.getParam(r, "id")
		us := s.Store.User()

		req := request{}
		err := s.decode(w, r, &req)
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "decoding-user", err)
			return
		}

		fmtErrs, err := s.validateStruct(r, req)
		if len(fmtErrs) > 0 {
			s.respondErr(w, r, http.StatusBadRequest, "user-json-validation", fmtErrs)
			return
		}
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "user-json-validation", err)
			return
		}

		addr := user.Addr{
			StreetName: req.StreetName,
			Number:     req.Number,
			Postcode:   req.Postcode,
			City:       req.City,
		}

		u, err := us.UpdateField(uid, "address", addr)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "updating-user", err)
			return
		}

		resp := response{
			Data: formator.NewJSONData("users", u.ID.Hex(), api_apbp.MapUserToJSON(u)),
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *Server) deleteUser() http.HandlerFunc {
	type response struct {
		Data formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {

		uid := s.getParam(r, "id")

		err := s.Store.User().Delete(uid)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "deleting-user", err)
			return
		}

		resp := response{
			Data: formator.NewJSONData("users", uid, nil),
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *Server) updatePwd() http.HandlerFunc {
	type request struct {
		OldPwd   string `json:"oldpwd" validate:"required"`
		Password string `json:"newpwd" validate:"required,pwd"`
		VerifPwd string `json:"verifpwd" validate:"required,eqfield=Password"`
	}

	type response struct {
		Data formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uid := s.getParam(r, "id")
		us := s.Store.User()

		var req request

		err := s.decode(w, r, &req)
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "decoding-user", err)
			return
		}

		fmtErrs, err := s.validateStruct(r, req)
		if len(fmtErrs) > 0 {
			s.respondErr(w, r, http.StatusBadRequest, "user-json-validation", fmtErrs)
			return
		}
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "user-json-validation", err)
			return
		}

		u, err := us.Read(uid)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "updating-user", err)
			return
		}

		if u.Password != "" {
			err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.OldPwd))
			if err != nil {
				s.respondErr(w, r, http.StatusBadRequest, "matching-password", err)
				return
			}
		}

		pwd, err := user.HashPassword(req.Password)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "hashing-password", err)
			return
		}

		u, err = us.UpdateField(uid, "password", pwd)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "updating-user", err)
			return
		}

		resp := response{
			Data: formator.NewJSONData("users", u.ID.Hex(), api_apbp.MapUserToJSON(u)),
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *Server) updateRole() http.HandlerFunc {
	type request struct {
		Email string `json:"email,omitempty" validate:"rfe=Role:admin,omitempty,email"`
		Role  string `json:"role,omitempty" validate:"required,oneof=admin customer"`
	}

	type response struct {
		Data formator.JsonData `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uid := s.getParam(r, "id")
		us := s.Store.User()

		var req request

		err := s.decode(w, r, &req)
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "decoding-user", err)
			return
		}

		fmtErrs, err := s.validateStruct(r, req)
		if len(fmtErrs) > 0 {
			s.respondErr(w, r, http.StatusBadRequest, "user-json-validation", fmtErrs)
			return
		}
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "user-json-validation", err)
			return
		}

		u, err := us.Read(uid)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "updating-user", err)
			return
		}

		u.Role = req.Role
		u.Email = req.Email

		u, err = us.UpdateFields(uid, u)
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "updating-user", err)
			return
		}

		resp := response{
			Data: formator.NewJSONData("users", u.ID.Hex(), api_apbp.MapUserToJSON(u)),
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}
