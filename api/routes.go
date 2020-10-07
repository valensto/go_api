package api

import (
	"github.com/go-chi/chi"
)

func (s *Server) routes() {
	s.commonMW()

	s.Router.Route("/v1", func(r chi.Router) {

		r.Route("/users", func(r chi.Router) {
			r.Get("/", s.restricted(s.listUser(false)))
			r.Get("/search", s.restricted(s.searchUser()))
			r.Get("/admin", s.restricted(s.listUser(true)))

			r.Post("/", s.restricted(s.createUser()))

			r.Route("/{id}", func(r chi.Router) {
				r.Put("/", s.restricted(s.updateUser()))
				r.Put("/password", s.restricted(s.updatePwd()))
				r.Put("/address", s.restricted(s.updateAddress()))
				r.Put("/role", s.restricted(s.updateRole()))

				r.Get("/", s.restricted(s.getUser()))
				r.Delete("/", s.restricted(s.deleteUser()))

				r.Post("/password", s.restricted(s.updatePwd()))
			})
		})

		r.Route("/products", func(r chi.Router) {
			r.Get("/", s.restricted(s.listProduct()))
			r.Get("/search", s.restricted(s.listProduct()))

			r.Post("/", s.restricted(s.createProduct()))

			r.Route("/{id}", func(r chi.Router) {
				r.Put("/", s.restricted(s.updateProduct()))
				r.Get("/", s.restricted(s.getProduct()))
				r.Delete("/", s.restricted(s.deleteProduct()))
			})
		})

		r.Route("/categories", func(r chi.Router) {
			r.Get("/", s.restricted(s.listCategory()))
		})

		r.Route("/orders", func(r chi.Router) {
			r.Get("/", s.restricted(s.listOrder()))
			r.Get("/search", s.restricted(s.listOrder()))
			r.Get("/forecast/confirm", s.restricted(s.forecast(true)))
			r.Get("/forecast", s.restricted(s.forecast(false)))

			r.Post("/", s.restricted(s.createOrder()))

			r.Route("/{id}", func(r chi.Router) {
				r.Put("/status", s.restricted(s.updateOrderStatus()))
				r.Put("/products", s.restricted(s.updateOrderProducts()))
				r.Put("/recovery", s.restricted(s.updateOrderRecovery()))
				r.Get("/", s.restricted(s.getOrder()))
				r.Delete("/", s.restricted(s.deleteOrder()))
			})
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", s.login())
			r.Post("/logout", s.login())
		})

	})
}
