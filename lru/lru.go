package lru

import "container/list"

type Cache struct {
	maxBytes  int64
	nbytes    int64
	ll        list.List
	cache     map[string]*list.Element
	onEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvict func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        *list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvict,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	ele, ok := c.cache[key]
	if ok {
		//移动到队尾
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, ok
	}
	return
}
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		//删除对头元素
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		//删除map的映射关系
		delete(c.cache, kv.key)
		//调整容量
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.onEvicted != nil {
			//如果删除的回调函数不为空则调用函数
			c.onEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	ele, ok := c.cache[key]
	if ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}
