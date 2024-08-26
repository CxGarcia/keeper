package httpjson

import (
	"fmt"
	"keeper/services/keeper"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

// Handler defines the httpjson handler
type Handler struct {
	echo   *echo.Echo
	keeper *keeper.Service
}

func New(keeper *keeper.Service) *Handler {
	e := echo.New()
	h := &Handler{
		echo:   e,
		keeper: keeper,
	}

	h.initMiddleware()

	return h
}

func (h *Handler) initMiddleware() {
	h.echo.Logger.SetLevel(log.DEBUG)
	h.echo.Use(middleware.Logger())

	// enrich the request with user settings
	h.echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			settings, err := h.keeper.GetUserSettings(c.Request().Context(), 1)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user settings")
			}

			c.Set("settings", settings)

			return next(c)
		}
	})

	// add the api key to the request header
	h.echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			settings, ok := c.Get("settings").(keeper.UserSettings)
			if !ok {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user settings")
			}

			c.Request().Header.Set("Authorization", fmt.Sprintf("Bearer %s", *settings.ApiKey))

			return next(c)
		}
	})

	h.echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			settings, ok := c.Get("settings").(keeper.UserSettings)
			if !ok {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user settings")
			}

			targetURL, err := url.Parse(settings.BaseURL)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "invalid target URL")
			}

			h.echo.Logger.Debugf("Proxying to %s", targetURL)

			proxyConfig := middleware.ProxyConfig{
				Balancer: middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
					{
						URL: targetURL,
					},
				}),
				Rewrite: map[string]string{
					"^/*": "/$1",
				},
				Transport: &http.Transport{
					Proxy: http.ProxyFromEnvironment,
				},
			}

			proxyMiddleware := middleware.ProxyWithConfig(proxyConfig)

			return proxyMiddleware(next)(c)
		}
	})
}

func (h *Handler) Start(addr string) {
	h.echo.Logger.Debug("Starting server")

	if err := h.echo.Start(addr); err != nil && err != http.ErrServerClosed {
		h.echo.Logger.Fatal("shutting down the server: ", err)
	}
}

func (h *Handler) Stop() error {
	if err := h.echo.Shutdown(nil); err != nil {
		return err
	}

	return nil
}
