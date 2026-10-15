package domain

type Cart struct {
	items map[uint64]*Item
}

func NewCart() *Cart {
	return &Cart{items: make(map[uint64]*Item)}
}

func (c *Cart) AddItem(sku uint64, name string, count uint32, price uint32) {
	if item, ok := c.items[sku]; ok {
		item.Count += count
		return
	}

	c.items[sku] = &Item{sku, name, count, price}
}

func (c *Cart) IsEmpty() bool {
	return len(c.items) == 0
}

func (c *Cart) Items() []Item {
	list := make([]Item, 0, len(c.items))
	for _, item := range c.items {
		list = append(list, *item)
	}

	return list
}
