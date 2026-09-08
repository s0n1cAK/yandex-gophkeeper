package http

import (
	"time"

	"yandex-gophkeeper/internal/transport/http/middleware"

	"github.com/go-chi/chi/v5"
	mw "github.com/go-chi/chi/v5/middleware"
)

func buildRouter(d Deps) *chi.Mux {
	r := chi.NewRouter()

	r.Use(mw.RequestID)
	r.Use(mw.RealIP)
	r.Use(mw.Recoverer)
	r.Use(mw.StripSlashes)
	r.Use(mw.Timeout(60 * time.Second))

	r.Use(middleware.Logging(d.Logger))
	r.Use(middleware.GzipCompession())

	r.Get("/ping", d.Ping.Ping)

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", d.Auth.Register)
			r.Post("/login", d.Auth.Login)

			r.Group(func(r chi.Router) {
				r.Use(d.AuthMW)

				r.Route("/secrets", func(r chi.Router) {
					r.Get("/", d.Secrets.List)
					r.Post("/", d.Secrets.Create)

					r.Route("/{id}", func(r chi.Router) {
						r.Get("/", d.Secrets.Get)
						r.Put("/", d.Secrets.Update)
						r.Delete("/", d.Secrets.Delete)
					})
				})
			})
		})
	})

	return r
}
