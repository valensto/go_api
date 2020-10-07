package api

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) login() http.HandlerFunc {

	type request struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	type token struct {
		Token string `json:"token"`
	}

	type response struct {
		Data token `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := request{}
		err := s.decode(w, r, &req)
		if err != nil {
			s.respondErr(w, r, http.StatusBadRequest, "decoding-request", err)
			return
		}

		respErr, err := s.validateStruct(r, req)
		if len(respErr) > 0 {
			s.respondErr(w, r, http.StatusBadRequest, "auth-validation-json", respErr)
			return
		}
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "auth-validation-json", err)
		}

		u, err := s.Store.User().FindByCredential(req.Email)
		if err != nil {
			s.respond(w, r, http.StatusNotFound, nil)
			return
		}

		// TODO not proud
		if u.Role != "admin" {
			s.respond(w, r, http.StatusNotFound, nil)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password))
		if err != nil {
			s.respond(w, r, http.StatusNotFound, nil)
			return
		}

		// TODO isAdmin
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"userID":  u.ID.Hex(),
			"isAdmin": true,
			"exp":     time.Now().Add(time.Hour * time.Duration(24)).Unix(),
			"iat":     time.Now().Unix(),
		})

		tokenStr, err := token.SignedString([]byte(s.Conf.JWTSecret))
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, "generate-jwt", err)
			return
		}

		w.Header().Add("x-auth-token", tokenStr)
		w.Header().Add("Access-Control-Expose-Headers", "x-auth-token")
		s.respond(w, r, http.StatusNoContent, nil)
	}
}
