package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"maps"
	"net/http"
	"slices"
)

type Handler struct {
	ctx         context.Context
	fridge      *Fridge
	repo        Repo
	notificator *Notificator
}

func NewHandler(ctx context.Context, repo Repo, fridge *Fridge) *Handler {
	return &Handler{
		ctx:         ctx,
		fridge:      fridge,
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
		if _, err := h.repo.Find(item.Name); err == nil {
			return c.JSON(http.StatusConflict, NewErrorResponse(fmt.Errorf("item %s already exists", item.Name)))
		}
		item, err := h.repo.Add(item)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, NewErrorResponse(fmt.Errorf("error when adding item: %v", err)))
		}
		h.notificator.Notify(item)
		return c.JSON(http.StatusCreated, item)
	}
}

func (h *Handler) UpdateOrCreateItem() func(echo.Context) error {
	return func(c echo.Context) error {
		item := Item{}
		if err := c.Bind(&item); err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse(fmt.Errorf("error when binding item: %v", err)))
		}
		found := false
		_, err := h.repo.Find(item.Name)
		if err != nil {
			found = true
		}
		item, err = h.repo.UpdateOrCreate(item)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, NewErrorResponse(fmt.Errorf("error when updating item: %v", err)))
		}
		h.notificator.Notify(item)
		if found {
			return c.JSON(http.StatusOK, item)
		}
		return c.JSON(http.StatusCreated, item)
	}
}

func (h *Handler) DeleteItem() func(echo.Context) error {
	return func(c echo.Context) error {
		name := c.Param("name")
		if err := h.repo.Delete(name); err != nil {
			return c.JSON(http.StatusInternalServerError, NewErrorResponse(fmt.Errorf("error when deleting item: %v", err)))
		}
		return c.JSON(http.StatusNoContent, nil)
	}
}

var upgrader = websocket.Upgrader{}

func (h *Handler) Notifications() func(echo.Context) error {
	return func(c echo.Context) error {
		wsCtx, wsCancel := context.WithCancel(h.ctx)
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			c.Logger().Error(fmt.Errorf("upgrader error: %v", err))
			wsCancel()
			return err
		}
		defer func() {
			h.notificator.RemoveClient(conn)
			fmt.Println("Client disconnected")
			conn.Close()
			wsCancel()
		}()
		h.notificator.AddClient(conn)
		fmt.Println("Client connected")

		go func() {
			defer func() {
				c.Logger().Error("notification goroutine exited")
				wsCancel()
			}()
			for {
				select {
				case item := <-h.notificator.Out:
					b, err := json.Marshal(item)
					if err != nil {
						c.Logger().Error(fmt.Errorf("json marshal error: %v", err))
						continue
					}
					for cc := range h.notificator.GetClients() {
						err = cc.WriteMessage(websocket.TextMessage, b)
						if err != nil {
							c.Logger().Error(fmt.Errorf("write message error: %v", err))
							return
						}
					}
				case <-wsCtx.Done():
					c.Logger().Warn("Ws context canceled.")
					return
				}
			}
		}()

		go func() {
			defer c.Logger().Error("read message goroutine exited")
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
						c.Logger().Warn(fmt.Errorf("read message error: %v", err))
					}
					wsCancel()
					return
				}
			}
		}()

		for {
			select {
			case <-wsCtx.Done():
				c.Logger().Warn("Websocket context canceled.")
				return nil
			case <-h.ctx.Done():
				c.Logger().Warn("Handler context canceled.")
				return nil
			}
		}

	}
}

func (h *Handler) GetFridgeItems() echo.HandlerFunc {
	return func(c echo.Context) error {
		// for now, we do not care about path param - only one fridge
		items := h.fridge.GetItems()
		basic := h.fridge.GetBasic()

		//return c.JSON(http.StatusOK, slices.Collect(maps.Values(items)))
		return c.JSON(http.StatusOK, map[string][]Item{
			"basic": slices.Collect(maps.Values(basic)),
			"items": slices.Collect(maps.Values(items)),
		})
	}
}

func (h *Handler) FillFridge() echo.HandlerFunc {
	return func(c echo.Context) error {
		shoppingList, err := h.repo.FindAll()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, NewErrorResponse(fmt.Errorf("error when retrieving shopping list: %v", err)))
		}
		err = h.fridge.Merge(shoppingList)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, NewErrorResponse(fmt.Errorf("error when filling a frige: %v", err)))
		}
		// todo add deleting shopping list after fill
		return c.JSON(http.StatusNoContent, nil)
	}
}

func (h *Handler) EmptyFridge() echo.HandlerFunc {
	return func(c echo.Context) error {
		_ = h.fridge.Empty()
		return c.JSON(http.StatusNoContent, nil)
	}
}
