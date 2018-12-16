Redix
======
> a very fast persistent pure key - value store, that uses the same [RESP](https://redis.io/topics/protocol) protocol and capable to store terabytes of data.
> Internally, I'm using [badgerdb](https://github.com/dgraph-io/badger) as storage 

Supported Commands
===================

##### Basic
- `PING`
- `QUIT`
- `SELECT`

##### Strings
- `SET <key> <value> [<TTL "millisecond">]`
- `MSET <key1> <value1> [<key2> <value2> ...]`
- `GET <key> [<default value>]`
- `MGET <key1> [<key2> ...]`
- `DEL <key1> [<key2> ...]`
- `EXISTS <key>`
- `INCR <key> [<by>]`

##### HASHES
- `HSET <HASHMAP> <KEY> <VALUE> <TTL>`
- `HMSET <HASHMAP> <key1> <value1> [<key2> <value2> ...]`
- `HGET <HASHMAP> <KEY>`
- `HDEL <HASHMAP> <key1> [<key2> ...]`
- `HGETALL <HASHMAP>`
- `HMSET <HASHMAP> <key1> <val1> [<key2> <val2> ...]`
- `HEXISTS <HASHMAP> [<key>]`, you can just check if the map is exists or not, or a key in the map exists or not.
- `HINCR <HASHMAP> <key> [<by>]`

##### LIST
- `LPUSH <LIST> <val1> [<val2> ...]`
- `LPUSHU <LIST> <val1> [<val2> ...]` push unique
- `LGETALL <LIST> [<offset> <size>]`
- `LREM <LIST> <val> [<val> <val> ...]`
- `LCOUNT <LIST>`

Install
=======
- from source: `go get github.com/alash3al/redix`
- from binaries: go [there](https://github.com/alash3al/redix/releases) and choose your platform based binary

Client SDKs
===========
> you can use any redis client from `redis-cli` or [from here](https://redis.io/clients)

Why
===
> I built this software to lear more about data modeling, data structrues and how to map any data to pure key value.

WHO AM I
========
> I'm Mohamed Al Ashaal, a software engineer, team leader and now I'm the CTO of [uflare](https://uflare.io)