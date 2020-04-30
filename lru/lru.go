package lru

import "container/list"

type Cache struct {
	// 最大可用内存
	// maxBytes为0时表示无限制
	maxBytes 		uint64
	// 已用内存
	usedBytes		uint64
	// 存储数据的链表
	ll 				*list.List
	// 映射key和链表中的节点. 字典下面的数据都在链表中. 当进行增删时会操作链表. 通过key进行存取时不需要遍历链表了.
	cache 			map[string]*list.Element
	// 当节点被驱逐的回调函数
	OnEvicted		func(key string, value Value)
}

// 链表节点的数据类型，即为list.Element
// 在这里其实cache中的key和此处的key为冗余的.
type entry struct {
	key 		string
	value 		Value
}


type Value interface {
	Len() 			uint64
}


func NewCache(maxBytes uint64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes: maxBytes,
		usedBytes: 0,
		ll: list.New(),
		cache: make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		// LRU 算法将访问过的(最不应该被删除的)放到链表头部.
		c.ll.MoveToFront(ele)
		// 强制类型转换
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	// 不写直接返回零值，其中ok为false
	return
}

// 缓存淘汰，将最久未被使用的东西进行删除
func (c *Cache) RemoveOldest() {
	// 最久未使用的在链表尾部
	ele := c.ll.Back()
	if ele != nil {
		// 1. 从链表中删除对应的value
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		// 2.从map中删除key对应的value
		// 在entry中存储key的原因就是可以直接使用kv中的key进行删除
		delete(c.cache, kv.key)
		// 处理cache结构中的东西
		c.usedBytes -= uint64(len(kv.key)) + kv.value.Len()
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// 增加; 在已经有key的时候进行替换
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.usedBytes = c.usedBytes - kv.value.Len() + value.Len()
		// 因为在c.cache中键值对的值为指针，即kv为entry指针，所以此处可以直接替换.
		kv.value = value
	} else {
		ele = c.ll.PushFront(&entry{key:key, value:value})
		c.cache[key] = ele
		c.usedBytes = c.usedBytes + uint64(len(key)) + value.Len()
	}

	// LRU 算法去除最旧的那个
	for c.maxBytes != 0 && c.usedBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) CountItem() int {
	return c.ll.Len()
}
