package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	ctx         context.Context
	repo        Repo
	notificator *Notificator
}

func NewHandler(ctx context.Context, repo Repo) *Handler {
	return &Handler{
		ctx:         ctx,
		repo:        repo,
		notificator: NewNotificator(),
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

func (h *Handler) GetItems() func(echo.Context) error {
	return func(c echo.Context) error {
		items, err := h.repo.FindAll()
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, items)
	}
}

func (h *Handler) AddItem() func(echo.Context) error {
	return func(c echo.Context) error {
		item := Item{}
		if err := c.Bind(&item); err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse(fmt.Errorf("error when binding item: %v", err)))
		}
		item, err := h.repo.Add(item)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, NewErrorResponse(fmt.Errorf("error when adding item: %v", err)))
		}
		h.notificator.Notify(item)
		return c.JSON(http.StatusCreated, item)
	}
}

func (h *Handler) UpdateItem() func(echo.Context) error {
	return func(c echo.Context) error {
		return nil
	}
}

func (h *Handler) DeleteItem() func(echo.Context) error {
	return func(c echo.Context) error {
		return nil
	}
}

var upgrader = websocket.Upgrader{}

func (h *Handler) Notifications() func(echo.Context) error {
	return func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			c.Logger().Error(fmt.Errorf("upgrader error: %v", err))
			return err
		}
		defer func() {
			h.notificator.RemoveClient(conn)
			conn.Close()
		}()
		h.notificator.AddClient(conn)
		fmt.Println("Client connected")

		go func() {
			defer c.Logger().Error("notification goroutine exited")
			for {
				select {
				case item := <-h.notificator.Out:
					b, err := json.Marshal(item)
					if err != nil {
						c.Logger().Error(fmt.Errorf("json marshal error: %v", err))
						continue
					}
					for cc := range h.notificator.GetClients() {
						if cc != conn {
							err = conn.WriteMessage(websocket.TextMessage, b)
							if err != nil {
								c.Logger().Error(fmt.Errorf("write message error: %v", err))
								return
							}
						}
					}
				}
			}
		}()

		for {
			select {
			case <-h.ctx.Done():
				c.Logger().Warn("Context canceled.")
				return nil
			}
		}

	}
}
