package mycache

import (
	"mycache/lru"
	"sync"
)

type cache struct {
	mu         sync.Mutex //互斥锁
	lru        *lru.Cache //实例化lru
	cacheBytes int64      //cache的大小
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil) //这里的nil代表了回调函数
	}
	c.lru.Add(key, value)
}
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	//return c.lru.Get(key) 这里不能这样写，因为我们要封装到ByteView里面去
	if val, ok := c.lru.Get(key); ok { //因为Get()返回的是Value 接口类型 然后用ByteView 去实例化它
		return val.(ByteView), ok
	}
	return
}
