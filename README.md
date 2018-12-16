Redix
=======
> a very fast persistent pure key - value store, that uses the same [RESP](https://redis.io/topics/protocol) protocol and capable to store terabytes of data.
> Internally, I'm using [badgerdb](https://github.com/dgraph-io/badger) as storage engine 

Features
=========
- Fast on-disk store.
- [ACID Transactions](https://blog.dgraph.io/post/badger-txn/)
- [Crash Resilient](https://blog.dgraph.io/post/alice/)
- Multi Core
- No blocking commands
- Very easy and simple
- Very compatible with any `redis-client`

Why
===
> I built this software to learn more about data modeling, data structures and how to map any data to pure key value, I don't need to build a redis clone, but I need to build something with my own concepts in my own style.

Install
=======
- from source: `go get github.com/alash3al/redix`.
- from binaries: go [there](https://github.com/alash3al/redix/releases) and choose your platform based binary, then download and execute from the command line with `-h` flag to see the help text.

Client SDKs
===========
> you can use any redis client from `redis-cli` or [from here](https://redis.io/clients)

Supported Commands
===================
> `Redix` doesn't implement all redis commands, but instead it supports the core concepts that will help you to build any type of data models on top of it, there are more commands and features in all next releases.

## # Basic
- `PING`
- `QUIT`
- `SELECT`

## # Strings
- `SET <key> <value> [<TTL "millisecond">]`
- `MSET <key1> <value1> [<key2> <value2> ...]`
- `GET <key> [<default value>]`
- `MGET <key1> [<key2> ...]`
- `DEL <key1> [<key2> ...]`
- `EXISTS <key>`
- `INCR <key> [<by>]`

## # HASHES
> I enhanced the HASH MAP implementation and added some features like TTL per nested key,
> also you can check whether the hash map itself exists or not using `HEXISTS <hashmapname>` or a nested key 
> exists using `HEXISTS <hashmapname> <keyname>`.  
> I'm planning to support removing the map itself using `HDELALL` (**todo**).

- `HSET <HASHMAP> <KEY> <VALUE> [<TTL "millesecond">]`
- `HMSET <HASHMAP> <key1> <value1> [<key2> <value2> ...]`
- `HGET <HASHMAP> <KEY>`
- `HDEL <HASHMAP> <key1> [<key2> ...]`
- `HGETALL <HASHMAP>`
- `HMSET <HASHMAP> <key1> <val1> [<key2> <val2> ...]`
- `HEXISTS <HASHMAP> [<key>]`.
- `HINCR <HASHMAP> <key> [<by>]`

## # LIST
> I applied a new concept, you can push or push-unique values into the list,
>  based on that I don't need to implement two different data structures, 
> as well as, you can quickly iterate over a list in a high performance way,
> every push will return the internal offset of the value, also, the iterator `lrange`
> will tell you the next offset you can start from.  
> I'm also planning to remove a list using `LDELALL` (**todo**).

- `LPUSH <LIST> <val1> [<val2> ...]`
- `LPUSHU <LIST> <val1> [<val2> ...]` push unique
- `LGETALL <LIST> [<offset> <size>]`
- `LREM <LIST> <val> [<val> <val> ...]`
- `LCOUNT <LIST>`


TODO Commands
=============
- `HDELALL <HASHMAP>`
- `LDELLALL <LIST>`

