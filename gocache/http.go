package gocache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// 基于http， 提供被其他节点获取的能力


const defaultBasePath = "/_gocache/"


type HTTPPool struct {
	// host
	self 		string
	basePath 	string
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self: self,
		basePath: defaultBasePath,
	}
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
