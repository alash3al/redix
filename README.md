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

## # Flat
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

- `HSET <HASHMAP> <KEY> <VALUE> [<TTL "millesecond">]`
- `HMSET <HASHMAP> <key1> <value1> [<key2> <value2> ...]`
- `HGET <HASHMAP> <KEY>`
- `HDEL <HASHMAP> [<key1> <key2> ...]` (deletes the map itself or keys in the map)
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

- `LPUSH <LIST> <val1> [<val2> ...]` (push the item into the list "it doesn't check for uniqueness, it will append anyway (duplicate)")
- `LPUSHU <LIST> <val1> [<val2> ...]` (push the item into the list only if it isn't exists)
- `LGETALL <LIST> [<offset> <size>]`
- `LREM <LIST> [<val1> <val2> <val3> ...]` (deletes the list itself or values in the list)
- `LCOUNT <LIST>` (get the list members count)
- `LSUM <LIST>` (sum the members of the list "in case they were numbers")
- `LAVG <LIST>` (get the avg of the members of the list "in case they were numbers")
- `LMIN <LIST>` (get the minimum of the members of the list "in case they were numbers")
- `LMAX <LIST>` (get the maximum of the members of the list "in case they were numbers")
- `LSRCH <LIST> <NEEDLE>` (text-search using (string search or regex) in the list)
- `LSRCHCOUNT <LIST> <NEEDLE>` (size of text-search result using (string search or regex) in the list)

## # Pub/Sub
> `Redix` has very simple pub/sub functionality, you can subscribe to internal logs on the `*` channel or any custom defined channel, and publish to any custom channel.

- `SUBSCRIBE [<channel1> <channel2>]`, if there is no channel specified, it will be set to `*`
- `PUBLISH <channel> <payload>`

## # Utils
> a helpers commands

- `ENCODE <method> <payload>`, encode the specified `<payload>` using the specified `<method>` (`md5`, `sha1`, `sha256`, `sha512`, `hex`)
- `UUIDV4`, generates a uuid-v4 string, i.e `0b98aa17-eb06-42b8-b39f-fd7ba6aba7cd`.
- `UNIQID`, generates a unique string.
- `RANDSTR [<size>, default size is 10]`, generates a random string using the specified length. 
- `RANDINT <min> <max>`, generates a random string between the specified `<min>` and `<max>`.
- `TIME`, returns the current time in `utc`, `seconds` and `nanoseconds`

TODO
=====
- [x] Basic Commands
- [x] Strings Commands
- [x] Hashmap Commands
- [x] List Commands
- [x] PubSub Commands
- [x] Utils Commands
- [ ] Document/JSON Commands
- [ ] GIS Commands