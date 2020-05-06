package gocache

import (
	"fmt"
	"log"
	"testing"
)

// 使用map模拟数据库
var db = map[string]string {
	"Tom": "630",
	"Jack": "589",
	"Sam": "567",
}

func TestGet(t *testing.T) {
	// 统计从db中直接拿数据的次数, 大于1次代表缓存没命中
	loadCounts := make(map[string]int, len(db))

	// 将一个匿名函数类型转换为GetterFunc..
	getfunc := GetterFunc(func(key string) ([]byte, error) {
		log.Printf("[DB] get %s from db", key)
		if v, ok := db[key]; ok {
			// 统计获取数据次数
			if _, ok := loadCounts[key]; !ok {
				loadCounts[key] = 0
			}
			loadCounts[key] += 1
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist in db", key)
	})

	cache := NewGroup("scores", 2<<10, getfunc)

	// 测试数据获取和缓存
	for k, v := range db {
		if view, err := cache.Get(k); err != nil || view.String() != v {
			t.Fatalf("failed to get value of %s", k)
		}
		// 第二次get缓存未命中 或 调用load函数多次
		if _, err := cache.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}


	if view, err := cache.Get("OfCourseNotExist"); err == nil {
		t.Fatalf("get value should be empty, got %s", view)
	}
}