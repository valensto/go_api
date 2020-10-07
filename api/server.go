package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/valensto/api_apbp/api/formator"
	config "github.com/valensto/api_apbp/configs"
	"github.com/valensto/api_apbp/infra/repo"
	"github.com/valensto/api_apbp/infra/store"
	"github.com/valensto/api_apbp/pkg/mailer"
	validator "github.com/valensto/api_apbp/pkg/validator"
)

// Server is a struct representation of a app server
type Server struct {
	Router    *chi.Mux
	Store     store.Store
	Validator *validator.Valider
	Conf      config.App
	Mailer    mailer.Sender
}

// NewServer is a struct of app server
func NewServer(conf config.App) (*Server, error) {
	s := &Server{
		Router:    chi.NewRouter(),
		Validator: validator.NewValider(),
		Conf:      conf,
	}

	s.routes()

	return s, nil
}

// InitStructValidator init validator
func (s *Server) InitStructValidator() error {
	return s.Validator.RegisterValidator()
}

func (s *Server) getParam(r *http.Request, k string) string {
	return chi.URLParam(r, k)
}

func (s *Server) respond(w http.ResponseWriter, _ *http.Request, status int, data interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	if data == nil {
		return
	}

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("cannot format json. err=%v\n", err)
	}
}

func (s *Server) respondErr(w http.ResponseWriter, r *http.Request, status int, title string, data interface{}) {
	if data == nil {
		return
	}

	if err, ok := data.(error); ok {
		var errRepo repo.ErrRepoOp
		if errors.As(err, &errRepo) {
			data = formator.NewRespErr(r, errRepo.Code, errRepo.Op, errRepo.Err.Error())
			status = errRepo.Code
		} else {
			data = formator.NewRespErr(r, status, title, err.Error())
		}
	}

	type responseErr struct {
		Errors interface{} `json:"errors"`
	}

	jsonErr := responseErr{
		Errors: data,
	}

	s.respond(w, r, status, jsonErr)
}

func (s *Server) validateStruct(r *http.Request, data interface{}) ([]formator.RespErr, error) {
	var fmtErrs []formator.RespErr

	if data == nil {
		return fmtErrs, errors.New("no data to validate")
	}

	strErrs, err := s.Validator.ValidateStruct(data)
	if len(strErrs) > 0 {
		for _, e := range strErrs {
			fmtErrs = append(
				fmtErrs,
				formator.NewRespErr(r, http.StatusBadRequest, validator.ErrInvalidAttribute.Error(), e),
			)
		}
		return fmtErrs, nil
	}

	if err != nil {
		return fmtErrs, err
	}

	return fmtErrs, nil
}

// Decode is an helper func to decode request body
func (s *Server) decode(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if r.ContentLength == 0 {
		return nil
	}
	return json.NewDecoder(r.Body).Decode(v)
}
