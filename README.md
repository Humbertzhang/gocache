# gocache

学习 [7days-golang](https://github.com/geektutu/7days-golang) 中模仿 [golang/groupcache](https://github.com/golang/groupcache)而写的分布式缓存。


### 主要流程

```
                            是
接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
                |  否                         是
                |-----> 是否应当从远程节点(peer)获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
                            |  否
                            |-----> 调用`回调函数`，从数据源(通常为数据库)获取值并添加到缓存 --> 返回缓存值 ⑶
```

### 其他特性

* 使用一致性Hash避免了一台机器宕机导致其他机器上缓存不可使用，避免了缓存雪崩.
* singleflight模块保证了在同一时刻请求同一个key时，apiserver仅会向其他分布式节点或数据源请求一次，避免了缓存击穿或缓存穿透。

```
缓存雪崩：缓存在同一时刻全部失效，造成瞬时DB请求量大、压力骤增，引起雪崩。缓存雪崩通常因为缓存服务器宕机、缓存的 key 设置了相同的过期时间等引起。

缓存击穿：一个存在的key，在缓存过期的一刻，同时有大量的请求，这些请求都会击穿到 DB ，造成瞬时DB请求量大、压力骤增。

缓存穿透：查询一个不存在的数据，因为不存在则不会写到缓存中，所以每次都会去请求 DB，如果瞬间流量过大，穿透到 DB，导致宕机。
```
