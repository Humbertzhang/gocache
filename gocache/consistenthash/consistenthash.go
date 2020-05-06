package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 映射 data to uin32
type Hash func(data []byte) uint32

// 一致性Hash主数据结构
type Map struct {
	// hash函数
	hash 		Hash
	// 虚拟节点倍数. 用于防止节点数量太少造成的大量key落在一个节点上的问题
	replicas 	int
	// 哈希环
	nodes 		[]int
	// 虚拟节点与真实节点映射表  虚拟节点哈希值:真实节点名
	hashMap		map[int]string
}


func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash: fn,
		hashMap: make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}


// 添加节点
// TODO: 添加节点后怎么处理之前节点中已有的数据？
func (m *Map) Add(nodeNames ...string) {
	for _, nodeName := range nodeNames {
		for i := 0; i < m.replicas; i++ {
			// 节点名为NodeName1 NodeName2 格式
			// 之后将节点名进行hash
			hash := int(m.hash( []byte(strconv.Itoa(i) + nodeName) ))
			// 哈希环添加节点
			m.nodes = append(m.nodes, hash)
			// 真实虚拟节点映射
			m.hashMap[hash] = nodeName
		}
	}
	sort.Ints(m.nodes)
}


// 选择节点. 返回节点名
// 此处key为想要存储的key名
func (m *Map) Get(key string) string{
	if len(m.nodes) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// binary search...
	idx := sort.Search(len(m.nodes), func(i int) bool {
		return m.nodes[i] >= hash
	})
	// 将node节点通过idx获取到
	return m.hashMap[m.nodes[idx % len(m.nodes)]]
}