package lru

import (
	"container/list"
)

// 为了通用性，值可以是实现了 Value 接口的任意类型，
// Value接口只包含了一个方法 Len() int，用于返回值所占用的内存大小
type Value interface {
	Len() int
}

type entry struct {
	key   string
	value Value
}

type Cache struct { // LRU cache
	maxBytes  int                           // maxBytes 是允许使用的最大内存
	nbytes    int                           // nbytes 是当前已使用的内存
	l1        *list.List                    //	基于双向链表实现的缓存
	cache     map[string]*list.Element      // 键是字符串，值是双向链表中对应节点的指针。
	OnEvicted func(key string, value Value) // OnEvicted 是某条记录被移除时的回调函数，可以为 nil
}

// Constructor of Cache
func New(maxBytes int, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		l1:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get value(use key)
//第一步是从字典中找到对应的双向链表的节点，第二步，将该节点移动到队尾
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.l1.MoveToFront(ele)    // 移动到队尾
		kv := ele.Value.(*entry) // 取出节点的值
		return kv.value, true    // 返回值和是否找到的标志
	}
	return
}

// 移除最近最少访问的节点（队首）,即缓存淘汰
func (c *Cache) RemoveOldest() {
	ele := c.l1.Back() //取到队首节点
	if ele != nil {
		c.l1.Remove(ele) // 删除队首节点
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)                  //从字典中 c.cache 删除该节点的映射关系。
		c.nbytes -= len(kv.key) + kv.value.Len() // 删除该节点后，更新内存使用量
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value) // 如果有回调函数（不为nil），则调用回调函数
		}
	}
}

// 新增或更新缓存
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok { //如果键存在，则更新对应节点的值，并将该节点移到队尾。
		c.l1.MoveToFront(ele)                    // 移动到队尾
		kv := ele.Value.(*entry)                 // 取出节点的值
		c.nbytes += value.Len() - kv.value.Len() // 更新内存使用量
		kv.value = value
	} else {
		ele := c.l1.PushFront(&entry{key, value}) // 如果键不存在，则新增节点，并将该节点移到队尾。
		c.cache[key] = ele
		c.nbytes += len(key) + value.Len()
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes { // 如果内存使用量超过了最大值，则移除最近最少访问的节点。
		c.RemoveOldest()
	}
}

// 获取缓存条目数
func (c *Cache) Len() int {
	return c.l1.Len()
}
