package lru

import (
	"reflect"
	"testing"
)

type String string

// 实现Len参数，以使String可以作为value
func (s String) Len() uint64 {
	return uint64(len(s))
}

func TestGet(t *testing.T) {
	lru := NewCache(uint64(0), nil)
	lru.Add("key1", String("1234"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}

	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	cap := len(k1 + k2 + v1 + v2)
	lru := NewCache(uint64(cap), nil)

	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	// Add k3时,会去除最旧的k1
	lru.Add(k3, String(v3))

	if _, ok := lru.Get(k1); ok || lru.CountItem() != 2 {
		t.Fatalf("Remove oldest key1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}

	lru := NewCache(uint64(10), callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("v2"))
	lru.Add("k3", String("v3"))
	lru.Add("k4", String("v4"))


	expect := []string{"key1", "k2"}
	// 深层次Equal中，slice array中底层数据也应该相等
	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed.")
	}
}

func TestAdd(t *testing.T)  {
	lru := NewCache(uint64(0), nil)
	lru.Add("key", String("1"))
	lru.Add("key", String("111"))

	if lru.usedBytes != uint64(len("key") + len("111")) {
		t.Fatalf("expected 6 used bytes but got %d", lru.usedBytes)
	}
}
