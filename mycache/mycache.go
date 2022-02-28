package mycache

import (
	"fmt"
	"log"
	"sync"
)

/*
函数类型实现某一个接口，称之为接口型函数，
方便使用者在调用时既能够传入函数作为参数，也能够传入实现了该接口的结构体作为参数。
*/

//这里不太明白

type Getter interface {
	Get(key string) ([]byte, error)
}
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

/*
	- 一个 Group 可以认为是一个缓存的命名空间，每个 Group 拥有一个唯一的名称 `name`。比如可以创建三个 Group，缓存学生的成绩命名为 scores，缓存学生信息的命名为 info，缓存学生课程的命名为 courses。
	- 第二个属性是 `getter Getter`，即缓存未命中时获取源数据的回调(callback)。
	- 第三个属性是 `mainCache cache`，即一开始实现的并发缓存。
	- 构建函数 NewGroup 用来实例化 Group，并且将 group 存储在全局变量 `groups` 中。
	- GetGroup 用来特定名称的 Group，这里使用了只读锁 `RLock()`，因为不涉及任何冲突变量的写操作。
*/
var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

type Group struct {
	name      string
	getter    Getter
	mainCache cache

	peers PeerPicker
}

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter********")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

//最重要的Group的Get方法
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required ***********")
	}
	if v, ok := g.mainCache.get(key); ok { //ok是true的话 就返回获取到的值啊！
		log.Println("[myCache] Hit********** ")
		return v, nil
	}
	return g.load(key) //没命中的话，就老老实实执行回调函数吧
}
func (g *Group) load(key string) (value ByteView, err error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if value, err = g.getFromPeer(peer, key); err == nil {
				return value, nil
			}
			log.Println("[mycache] Failed to get from peer", err)
		}
	}
	return g.getLocally(key)
}
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil { //有错误就执行
		return ByteView{}, err
	}
	//正常就返回 源数据的只读拷贝
	value := ByteView{b: cloneBytes(bytes)}
	g.popalateCache(key, value) //并且把源数据 迁移到缓存里面去
	return value, nil
}
func (g *Group) popalateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once 这个远程结点已经注册好了")
	}
	g.peers = peers
}
