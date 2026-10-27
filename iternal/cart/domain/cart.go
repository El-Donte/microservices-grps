package domain

import (
	"sort"
	"sync"
)

type Cart struct {
	mu    sync.Mutex
	items map[uint64]*Item
}

func NewCart() *Cart {
	return &Cart{items: make(map[uint64]*Item)}
}

func (c *Cart) AddItem(sku uint64, name string, count uint32, price uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, ok := c.items[sku]; ok {
		item.Count += count
		return
	}

	c.items[sku] = &Item{sku, name, count, price}
}

func (c *Cart) IsEmpty() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.items) == 0
}

func (c *Cart) Items() []Item {
	c.mu.Lock()
	defer c.mu.Unlock()

	list := make([]Item, 0, len(c.items))
	for _, item := range c.items {
		list = append(list, *item)
	}

	sort.Slice(list, func(i, j int) bool { return list[i].SKU < list[j].SKU })
	return list
}
