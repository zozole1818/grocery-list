package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

type Fridge struct {
	basicItems map[string]Item
	items      map[string]Item
	mu         sync.Mutex
}

type FridgeStorage struct {
	storage map[string]map[string]Item
}

var basicItems = []string{"mleko", "chleb", "jaja", "woda3L"}

func initBasicItems() map[string]Item {
	items := make(map[string]Item)
	for i, name := range basicItems {
		items[name] = Item{
			ID:         i + 1,
			Name:       name,
			Count:      0,
			CreateDate: time.Now().Format(time.RFC3339),
		}
	}
	return items
}

func NewFridge() *Fridge {
	emptyFridge := &Fridge{
		basicItems: initBasicItems(),
		items:      make(map[string]Item),
		mu:         sync.Mutex{},
	}
	b, err := os.ReadFile("fridge.json")
	if err != nil {
		slog.Debug("Error when reading fridge file. Will start with empty list.", "error", err)
		return emptyFridge
	}
	fs := FridgeStorage{}
	err = json.Unmarshal(b, &fs.storage)
	if err != nil {
		slog.Debug("Error when Unmarshal shopping list file. Will start with empty list.", "error", err)
		return emptyFridge
	}
	return &Fridge{
		basicItems: fs.storage["basic"],
		items:      fs.storage["items"],
		mu:         sync.Mutex{},
	}
}

func (f *Fridge) GetItems() map[string]Item {
	f.mu.Lock()
	defer f.mu.Unlock()
	items := make(map[string]Item, len(f.items))
	for k, v := range f.items {
		items[k] = v
	}
	return items
}

func (f *Fridge) GetBasic() map[string]Item {
	f.mu.Lock()
	defer f.mu.Unlock()
	items := make(map[string]Item, len(f.items))
	for k, v := range f.basicItems {
		items[k] = v
	}
	return items
}

func (f *Fridge) Merge(shoppingList []Item) error {
	fridgeItems := f.GetItems()
	basic := f.GetBasic()
	for _, item := range shoppingList {
		name := item.Name
		if _, ok := basic[name]; ok {
			b := basic[name]
			b.Count += item.Count
			basic[name] = b
		} else if _, ok := fridgeItems[name]; ok {
			fi := fridgeItems[name]
			fi.Count += item.Count
			fridgeItems[name] = fi
		} else {
			fridgeItems[name] = item
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.items = fridgeItems
	f.basicItems = basic
	return nil
}

func (f *Fridge) Empty() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.items = make(map[string]Item)
	return nil
}

func (f *Fridge) Flush() error {
	slog.Debug("Flushing fridge")
	f.mu.Lock()
	defer f.mu.Unlock()
	fs := FridgeStorage{
		storage: map[string]map[string]Item{
			"basic": f.basicItems,
			"items": f.items,
		},
	}
	slog.Debug("Flushing fridge", "fridge", fs)
	b, err := json.Marshal(fs.storage)
	if err != nil {
		return fmt.Errorf("error when marshaling fridge items: %v", err)
	}
	err = os.WriteFile("fridge.json", b, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error when flushing fridge: %v", err)
	}
	return nil
}
