package httpjson

import (
	"context"
	"fmt"
	"keeper/services/keeper"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Handler defines the httpjson handler
type Handler struct {
	server *http.Server
	keeper *keeper.Service
}

func New(keeper *keeper.Service) *Handler {
	h := &Handler{
		keeper: keeper,
	}

	return h.init()
}

func (h *Handler) init() *Handler {
	mux := http.NewServeMux()

	h.server = &http.Server{
		Handler: h.logMiddleware(h.userSettingsMiddleware(h.apiKeyMiddleware(h.proxyMiddleware(mux)))),
	}

	return h
}

func (h *Handler) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) userSettingsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		settings, err := h.keeper.GetUserSettings(ctx, 1)
		if err != nil {
			http.Error(w, "failed to get user settings", http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r.WithContext(
			context.WithValue(ctx, "settings", settings),
		))
	})
}

func (h *Handler) apiKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		settings, ok := r.Context().Value("settings").(keeper.UserSettings)
		if !ok {
			http.Error(w, "failed to get user settings", http.StatusInternalServerError)
			return
		}

		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *settings.ApiKey))

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) proxyMiddleware(_ http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		settings, ok := r.Context().Value("settings").(keeper.UserSettings)
		if !ok {
			http.Error(w, "failed to get user settings", http.StatusInternalServerError)
			return
		}

		var targetURL *url.URL
		if r.URL.Query().Get("debug") == "true" {
			targetURL = &url.URL{
				Scheme: "http",
				Host:   "localhost:3000",
			}
		} else {
			url, err := url.Parse(settings.BaseURL)
			if err != nil {
				http.Error(w, "invalid target URL", http.StatusInternalServerError)
				return
			}
			targetURL = url
		}

		log.Printf("Proxying to %s", targetURL)

		proxy := &httputil.ReverseProxy{
			Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(targetURL)
			},
		}

		proxy.ServeHTTP(w, r)
	})
}

func (h *Handler) Start(addr string) {
	log.Printf("Starting server on %s", addr)
	h.server.Addr = addr
	if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("shutting down the server: %v", err)
	}
}

func (h *Handler) Stop() error {
	return h.server.Shutdown(context.Background())
}
