package api

import (
	"avito-shop/internal/api/handlers"
	"avito-shop/internal/api/middleware"
	"avito-shop/internal/service"
	"net/http"
)

type Router struct {
	services *service.Services
	mux      *http.ServeMux
}

func NewRouter(services *service.Services) *Router {
	return &Router{
		services: services,
		mux:      http.NewServeMux(),
	}
}

func (r *Router) Setup() http.Handler {
	r.mux.Handle("/api/auth", handlers.NewAuthHandler(r.services.Users))

	r.mux.Handle("/api/info", middleware.AuthMiddleware(r.services.TokenSecret)(
		handlers.NewInfoHandler(r.services.Info)))
	r.mux.Handle("/api/sendCoin", middleware.AuthMiddleware(r.services.TokenSecret)(
		handlers.NewTransferHandler(r.services.Users)))
	r.mux.Handle("/api/buy/", middleware.AuthMiddleware(r.services.TokenSecret)(
		handlers.NewBuyHandler(r.services.Merchandise)))

	return r.mux
}
