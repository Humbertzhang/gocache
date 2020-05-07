package gocache

import (
	"./consistenthash"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// 基于http， 提供被其他节点获取的能力


const (
	defaultBasePath = "/_gocache/"
	defaultReplicas = 50
)

// HTTPPool implements PeerPicker.
// httpGetter implements PeerGetter
type HTTPPool struct {
	// host
	self 		string
	basePath 	string
	//  guards peers and httpGetters
	mu 			sync.Mutex
	// 一致性哈希Map
	peers 		*consistenthash.Map
	httpGetters map[string]*httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self: self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) SetPeer(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// p.peers.Get 返回节点名
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		// 返回一个httpGetter对象，httpGetter为PeerGetter的实现
		// 其可以直接使用Get获取该peer上存储的数据
		return p.httpGetters[peer], true
	}
	return nil, false
}


func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}


func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		errStr := "path not found:" + r.URL.Path + ", try " + p.basePath
		http.Error(w, errStr,  http.StatusNotFound)
		return
	}

	p.Log("%s %s", r.Method, r.URL.Path)

	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	// 需要如下模式的字符串作为请求URL的Path /<basepath>/<groupname>/<key>
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// 以下为从本地的group中取数据的操作
	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:" + groupName, http.StatusNotFound)
		return
	}

	// TODO: 在这里err为NotFound的情况下仍为500
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}


// 客户端

type httpGetter struct {
	baseURL 		string
}

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", h.baseURL,
										url.QueryEscape(group),
										url.QueryEscape(key),)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

// var _ struct名 = (*Interface名)(nil) 可以确定struct是否实现了interface
var _ PeerPicker = (*HTTPPool)(nil)
var _ PeerGetter = (*httpGetter)(nil)


