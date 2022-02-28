package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash
	replicas int //虚拟结点与真实结点的倍数
	keys     []int
	hashMap  map[int]string //这里的hashMap是用int类型的hash值，找到对应的string 真实结点
}

func New(replicas int, fn Hash) *Map { //这里传入的Hash是 自己自定义的fn函数
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil { //如果这里传入的fn是nil的话，我们就默认使用 ，采取依赖注入的方式，允许用于替换成自定义的 Hash 函数，也方便测试时替换，默认为 `crc32.ChecksumIEEE` 算法。
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

//添加真实的结点进去
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

//获取顺时针最近的结点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	index := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	return m.hashMap[m.keys[index%len(m.keys)]]
}

/*
- 选择节点就非常简单了，第一步，计算 key 的哈希值。
- 第二步，顺时针找到第一个匹配的虚拟节点的下标 `index`，从 m.keys 中获取到对应的哈希值。如果 `index == len(m.keys)`，说明应选择 `m.keys[0]`，因为 `m.keys` 是一个环状结构，所以用取余数的方式来处理这种情况。
- 第三步，通过 `hashMap` 映射得到真实的节点。
*/
