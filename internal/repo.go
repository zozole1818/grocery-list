package internal

import (
	"sync"
	"time"
)

type Repo interface {
	FindAll() ([]Item, error)
	Add(item Item) (Item, error)
	Update(item Item) error
	Delete(id int) error
}

type MapRepo struct {
	db map[int]Item
	mu sync.Mutex
}

func NewRepo() Repo {
	return &MapRepo{
		db: make(map[int]Item),
		mu: sync.Mutex{},
	}
}

func (r *MapRepo) FindAll() ([]Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := make([]Item, 0, len(r.db))
	for _, v := range r.db {
		items = append(items, v)
	}
	return items, nil
}

func (r *MapRepo) Add(item Item) (Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item.ID = len(r.db) + 1
	item.CreateDate = time.Now().Format(time.RFC3339)
	r.db[item.ID] = item
	return item, nil
}

func (r *MapRepo) Update(item Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.db[item.ID] = item // just replace the item
	return nil
}

func (r *MapRepo) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.db, id)
	return nil
}
