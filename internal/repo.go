package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

type Repo interface {
	FindAll() ([]Item, error)
	Find(name string) (Item, error)
	Add(item Item) (Item, error)
	UpdateOrCreate(item Item) (Item, error)
	Delete(name string) error
	DeleteAll() error
}

type MapRepo struct {
	shoppingList map[string]Item
	mu           sync.Mutex
}

func NewMapRepo() Repo {
	b, err := os.ReadFile("shopping-list.json")
	if err != nil {
		slog.Debug("Error when reading shopping list file. Will start with empty list.", "error", err)
		return &MapRepo{
			shoppingList: make(map[string]Item),
			mu:           sync.Mutex{},
		}
	}
	repo := &MapRepo{
		shoppingList: make(map[string]Item),
		mu:           sync.Mutex{},
	}
	err = json.Unmarshal(b, &repo.shoppingList)
	if err != nil {
		slog.Debug("Error when Unmarshal shopping list file. Will start with empty list.", "error", err)
		return &MapRepo{
			shoppingList: make(map[string]Item),
			mu:           sync.Mutex{},
		}
	}
	return repo

}

func (r *MapRepo) FindAll() ([]Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := make([]Item, 0, len(r.shoppingList))
	for _, v := range r.shoppingList {
		items = append(items, v)
	}
	return items, nil
}

func (r *MapRepo) Find(name string) (Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.shoppingList[name]
	if ok {
		return item, nil
	}
	return Item{}, fmt.Errorf("item %s not found", name)
}

func (r *MapRepo) Add(item Item) (Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item.ID = len(r.shoppingList) + 1
	item.CreateDate = time.Now().Format(time.RFC3339)
	r.shoppingList[item.Name] = item
	return item, nil
}

func (r *MapRepo) UpdateOrCreate(item Item) (Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	oldItem, ok := r.shoppingList[item.Name]
	if ok {
		item.ID = oldItem.ID
		item.CreateDate = oldItem.CreateDate
		item.UpdateDate = time.Now().Format(time.RFC3339)
		r.shoppingList[item.Name] = item // just replace the item
	} else {
		item.ID = len(r.shoppingList) + 1
		item.CreateDate = time.Now().Format(time.RFC3339)
		r.shoppingList[item.Name] = item
	}
	return item, nil
}

func (r *MapRepo) Delete(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.shoppingList, name)
	return nil
}

func (r *MapRepo) DeleteAll() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.shoppingList = map[string]Item{}
	return nil
}

func (r *MapRepo) Flush() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, err := json.Marshal(r.shoppingList)
	if err != nil {
		return fmt.Errorf("error when marshaling shopping list: %v", err)
	}
	err = os.WriteFile("shopping-list.json", b, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error when flushing shopping list: %v", err)
	}
	return nil
}
