package main

import (
	"./gocache"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string {
	"Tom": "630",
	"Jack": "589",
	"Sam": "567",
}

func createCacheGroup() *gocache.Group {
	return gocache.NewGroup("scores", 2<<10, gocache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// cache Server 即为一个Peer
func startCacheServer(addr string, addrs []string, cache *gocache.Group) {
	peers := gocache.NewHTTPPool(addr)
	peers.SetPeer(addrs...)
	cache.RegisterPeers(peers)
	log.Println("gocache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

// 代理所有的CacheServer, 与用户进行交互。所有的CacheServer对外只暴露这一个API接口
func startAPIServer(apiAddr string, cache *gocache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := cache.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("frontend server is running at:", apiAddr)
	// apiAddr[7:] 去掉http://
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "GoCache Server Port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	cache := createCacheGroup()
	if api {
		go startAPIServer(apiAddr, cache)
	}
	startCacheServer(addrMap[port], []string(addrs), cache)
}
