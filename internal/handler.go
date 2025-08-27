package internal

import (
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	ctx context.Context
}

func NewHandler(ctx context.Context) *Handler {
	return &Handler{
		ctx: ctx,
	}
}

func (h *Handler) GetQuote() func(echo.Context) error {
	return func(c echo.Context) error {
		quote := QuoteResponse{
			Quote: "Life is like a box of chocolates, you never know what you're going to get.",
		}

		return c.JSON(http.StatusOK, quote)
	}
}
