package lru

import (
	"container/list"
)

type Value interface { //Value接口 实现了 Value 接口的任意类型
	Len() int //记录value所占用的内存大小
}
type entry struct {
	key   string
	value Value
}
type Cache struct {
	maxBytes int64                    //允许的最大占内存数
	nBytes   int64                    //已占内存数
	ll       *list.List               //双向链表 中 队首的元素优先删除
	cache    map[string]*list.Element //cache是整个LRU的数据结构，
	//其中Element是这个双向链表的容器封装的结点

	OnEvicted func(key string, value Value) //回调函数
	//OnEvicted暂时不懂 （OnEvicted是某条记录被移除时的回调函数，可以为 nil）
}

//为了方便测试，我们实现 Len() 用来获取添加了多少条数据。
func (c *Cache) Len() int {
	return c.ll.Len()
}

//为了实例化Cache 先来实现一个New() 函数
func New(maxBytes int64, OnEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: OnEvicted, //OnEvicted暂时不懂 （OnEvicted是某条记录被移除时的回调函数，可以为 nil）
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)    //将链表中的节点 ele 移动到队尾（双向链表作为队列，队首队尾是相对的，在这里定 约front 为队尾）
		kv := ele.Value.(*entry) //ele.Value是一个接口 后面的(*entry)是实例化这个结点
		return kv.value, ok
	}
	return
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back() //取出双向链表队首的结点
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry) //实例化结点
		delete(c.cache, kv.key)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value) //如果还有回调函数的话，执行一下再退出
		}
	}
}
func (c *Cache) Add(key string, value Value) {
	ele := c.cache[key]
	if ele != nil { //如果已经存在的话，就更新数据
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nBytes = c.nBytes - int64(kv.value.Len()) + int64(value.Len())
		kv.value = value
	} else { //不存在的话，就push这个数据
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nBytes = c.maxBytes + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}
