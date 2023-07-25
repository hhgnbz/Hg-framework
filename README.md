# Hg-framework

Self study project

Refer to : [geektutu - 7天用Go从零实现系列](https://geektutu.com/post/gee.html)

# Hint

A Gin like web framework

### Features

- Dynamic router (Trie base)
- Routes grouping
- Middlewares support (Default Crash-free and Logger)
- Panic handle (Crash-free)
- Static templates support

# HintCache

A Memcached(Groupcache) like distributed cache framework

### Features

- LRU cache strategy
- Lock (Mutex base)
- Load balance (Consistent hash base)
- Optimized binary communication (protobuf base)

# HintRPC

A RPC framework based on "net/rpc" package

### Features

- Proto exchange
- Registry
- Service discovery
- HTTP protocol support
- Load balance
- Timeout processing