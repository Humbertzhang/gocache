package singleflight

import "sync"

type call struct {
	wg 		sync.WaitGroup
	val 	interface{}
	err 	error
}

// 防止同一时间，多个请求请求同一个没在缓存中的key
// 导致同时对数据库发起很多请求，造成缓存击穿或缓存穿透
type Throttler struct {
	mu 		sync.Mutex
	m 		map[string]*call
}

// 对于相同的key，无论Do被调用多少次，fn都只调用一次
func (t *Throttler) Do(key string, fn func() (interface{}, error) ) (interface{}, error) {
	// 加锁,防止重入
	t.mu.Lock()
	// 此处为对m的延迟初始化处理。使内存利用率更高
	if t.m == nil {
		t.m = make(map[string]*call)
	}

	// t.m[]可以获取到key，代表这个key有正在处理的call
	if c, ok := t.m[key]; ok {
		t.mu.Unlock()
		//等待该Call处理完毕
		c.wg.Wait()
		return c.val, c.err
	}

	// 如果t.m中没有，则创建该call
	c := new(call)
	c.wg.Add(1)
	t.m[key] = c
	// 解锁
	t.mu.Unlock()

	c.val, c.err = fn()
	// 调用Done时，如果有其他goroutine在等待完成，则会获得c的相应结果
	c.wg.Done()

	// 请求完毕后从map中删除.
	t.mu.Lock()
	delete(t.m, key)
	t.mu.Unlock()

	return c.val, c.err
}