package gocache

import (
	"fmt"
	"log"
	"sync"
)

/*
接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
                |  否                         是
                |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
                            |  否
                            |-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶
*/

// 控制缓存存储和获取

// getter负责当缓存中不存在时，从数据源获取数据
type Getter interface {
	Get(key string) ([]byte, error)
}

// Getter是一个实现了Getter接口的函数
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}


type Group struct {
	// 缓存组别名字. 缓存的namespace
	name 		string
	// 缓存未命中时数据回源获取回调函数
	getter 		Getter
	// 带并发的cache
	mainCache 	cache
	// 流程2时从peer中获取其他节点缓存的数据
	peers 		PeerPicker
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("Can't register peers twice.")
	}
	g.peers = peers
}

// 此处实现流程1， 先尝试从本地缓存获取数据
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Printf("[Cache] %s hit", key)
		return v, nil
	}

	// 尝试从其他地方获取数据
	return g.load(key)
}


func (g *Group) load(key string) (v ByteView, err error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if v, err = g.getFromPeer(peer, key); err == nil {
				return v, nil
			}
			log.Println("[GoCache] Failed get from peer:", err)
		}
	}
	// 尝试从数据源获取数据
	return g.getLocal(key)
}

// 本地缓存未命中，尝试在peer中获取缓存
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

// 当数据在缓存中没有命中时，从本地数据源中获取å
func (g *Group) getLocal(key string) (v ByteView, err error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}


// 将从数据源中获取的数据加入到数据源
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}


var (
	// 此处 mu 用来保护并发获取groups
	mu 			sync.RWMutex
	groups  = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes uint64, getter Getter) *Group {
	if getter == nil {
		panic("Getter is nil")
	}

	g := &Group{
		name: name,
		getter: getter,
		mainCache: cache{cacheBytes:cacheBytes},
	}

	mu.Lock()
	defer mu.Unlock()

	groups[name] = g
	return g
}


func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}



