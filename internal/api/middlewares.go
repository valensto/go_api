package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/valensto/api_apbp/internal/session"
)

func (s *Server) commonMW() {
	s.Router.Use(cors.Handler(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	s.Router.Use(middleware.RequestID)
	s.Router.Use(middleware.RealIP)
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Recoverer)
	s.Router.Use(middleware.Timeout(10 * time.Second))
	s.Router.Use(httprate.LimitByIP(100, 1*time.Minute))
}

func (s *Server) restricted(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if bearer := r.Header.Get("Authorization"); bearer != "" {

			var tk string
			tk = strings.Replace(bearer, "Bearer ", "", 1)
			token, _ := jwt.Parse(tk, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}

				return []byte(s.Conf.JWTSecret), nil
			})

			if token == nil {
				s.respond(w, r, http.StatusUnauthorized, nil)
				return
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				ctx := session.WithUserID(r.Context(), claims["userID"].(string))
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		s.respond(w, r, http.StatusUnauthorized, nil)
	}
}
