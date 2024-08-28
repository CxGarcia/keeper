package proxy

import (
	"context"
	"fmt"
	log "keeper/internal/logger"
	"keeper/services/keeper"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Service defines the proxy handler
type Service struct {
	server *http.Server
	keeper *keeper.Repository
}

func New(keeper *keeper.Repository) *Service {
	h := &Service{
		keeper: keeper,
	}

	return h.init()
}

func (h *Service) init() *Service {
	mux := http.NewServeMux()

	h.server = &http.Server{
		Handler: h.logMiddleware(h.userSettingsMiddleware(h.apiKeyMiddleware(h.proxyMiddleware(mux)))),
	}

	return h
}

func (h *Service) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("Request: %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)
	})
}

func (h *Service) userSettingsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		settings, err := h.keeper.GetUserSettings(ctx, 1)
		if err != nil {
			http.Error(w, "failed to get user settings", http.StatusInternalServerError)

			log.Errorf("failed to get user settings: %v", err)

			return
		}

		next.ServeHTTP(w, r.WithContext(
			context.WithValue(ctx, "settings", settings),
		))
	})
}

func (h *Service) apiKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		settings, ok := r.Context().Value("settings").(keeper.UserSettings)
		if !ok {
			http.Error(w, "failed to get user settings", http.StatusInternalServerError)
			return
		}

		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", settings.Secret))

		next.ServeHTTP(w, r)
	})
}

func (h *Service) proxyMiddleware(_ http.Handler) http.Handler {
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

		log.Debugf("Proxying to %s", targetURL)

		proxy := &httputil.ReverseProxy{
			Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(targetURL)
			},
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				http.Error(w, "failed to proxy request", http.StatusInternalServerError)
				log.Errorf("failed to proxy request: %v", err)
			},
		}

		proxy.ServeHTTP(w, r)
	})
}

func (h *Service) Start(addr string) error {
	log.Infof("Starting server on %s", addr)

	h.server.Addr = addr
	if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (h *Service) Stop() error {
	return h.server.Shutdown(context.Background())
}
