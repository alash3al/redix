<p align="center"> 
      <img src="https://via.placeholder.com/800x200/fff/000/?text=RedixDB" />
</p>

<p align="center">
      <a style="display: inline-block" align="center" href="https://travis-ci.com/alash3al/redix"><img alt="Build Status" src="https://travis-ci.com/alash3al/redix.svg?branch=master" /></a>
      <a style="display: inline-block" align="center" href="https://github.com/alash3al/redix/blob/master/LICENSE"><img alt="License" src="https://img.shields.io/hexpm/l/plug.svg" /></a>
      <a style="display: inline-block" align="center" href="https://cloud.docker.com/u/alash3al/repository/docker/alash3al/redix"><img alt="Docker" src="https://img.shields.io/docker/pulls/alash3al/redix.svg" /></a>
      <a style="display: inline-block" align="center" href="https://github.com/alash3al/redix/graphs/contributors"><img alt="Contributors" src="https://img.shields.io/github/contributors/alash3al/redix.svg" /></a>
</p>

<blockquote align="center">

a fast NoSQL DB, that uses the same <a href="https://redis.io/topics/protocol">RESP</a> protocol and capable to store terabytes of data, also it integrates with your mobile/web apps to add real-time features, soon you can use it as a document store cause it should become a multi-model db. `Redix` is used in production, you can use it in your apps with no worries.

</blockquote>

Features
=========
- Core data structure: `KV`, `List`, `Hashmap` with advanced implementations.
- Advanced Publish/Subscribe using webhook and websocket!
- Pluggable Storage Engine (`badgerdb`, `boltdb`, `leveldb`, `null`, `sqlite`)
- Very compatible with any `redis client` including `redis-cli`
- Standalone with no external dependencies
- Helpers commands for `Time`, `Encode <hex|md5|sha1|sha256|sha512> <payload>`, `RANDINT`, `RANDSTR`
- Implements `RATELIMIT` helpers natively.

Why
===
> I started this software to learn more about data modeling, data structures and how to map any data to pure key value, I don't need to build a redis clone, but I need to build something with my own concepts in my own style. I decided to use RESP (redis protocol) so you can use `Redix` with any redis client out there.

Install
=======
- Using Homebrew:
  - Add Homebrew Tap `brew tap alash3al/redix https://github.com/alash3al/redix`
  - Install Redix `brew install alash3al/redix/redix`
- From Binaries: go [there](https://github.com/alash3al/redix/releases) and choose your platform based binary, then download and execute from the command line with `-h` flag to see the help text.
- Using Docker: `docker run -P -v /path/to/redix-data:/root/redix-data alash3al/redix`
- From Source: `go get github.com/alash3al/redix`.

Configurations
============
> It is so easy to configure `Redix`, there is no configuration files, it is all about running `./redix` after you download it from the [releases](https://github.com/alash3al/redix/releases), if you downloaded i.e 'redix_linux_amd64' and unziped it.

```bash
$ ./redix_linux_amd64 -h

  -engine string
        the storage engine to be used, available (default "badger")
  -http-addr string
        the address of the http server (default ":7090")
  -resp-addr string
        the address of resp server (default ":6380")
  -storage string
        the storage directory (default "./redix-data")
  -verbose
        verbose or not
  -workers int
        the default workers number (default ...)
```

Examples
=========

```bash

# i.e: $mykey1 = "this is my value"
$ redis-cli -p 6380 set mykey1 "this is my value"

# i.e: $mykey1 = "this is my value" and expire it after 10 seconds
$ redis-cli -p 6380 set mykey1 "this is my value" 10000

# i.e: echo $mykey1
$ redis-cli -p 6380 get mykey1

# i.e: $mymap1[x] = y
$ redis-cli -p 6380 hset mymap1 x y

# i.e: $mymap1[x] = y and expires it after 10 seconds
$ redis-cli -p 6380 hset mymap1 x y 10000

# i.e: sha512 of "test"
$ redis-cli -p 6380 encode sha512 test

# you want to notify an endpoint i.e: "http://localhost:800/new-data" that there is new data available, in other words, you want to subscribe a webhook to channel updates.
$ redis-cli -p 6380 webhookset testchan http://localhost:800/new-data

# add data to a list
# i.e: [].push(....)
$ redis-cli -p 6380 lpush mylist1 "I'm Mohammed" "I like to Go using Go" "I love coding"

# search in the list
$ redis-cli -p 6380 lsrch mylist1 "mo(.*)"

```

DB Engines
===========
- `Redix` supports two engines called `badger` and `bolt`
- `badger` is the default, it is inspired by Facebook [`RocksDB`](https://rocksdb.org/), it meant to be fast on-disk engine, [read more](https://github.com/dgraph-io/badger)
- `bolt` is our alternate engine, it is inspired by [`LMDB`](http://symas.com/mdb/), [read more](https://github.com/etcd-io/bbolt)

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
- `TTL <key>` returns `-1` if key will never expire, `-2` if it doesn't exists (expired), otherwise will returns the `seconds` remain before the key will expire.
- `KEYS [<regexp-pattern>]`


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
- `HTTL <HASHMAP> <key>`, the same as `TTL` but for `HASHMAP`
- `HKEYS <HASHMAP>`
- `HLEN <HASHMAP>`

## # LIST
> I applied a new concept, you can push or push-unique values into the list,
>  based on that I don't need to implement two different data structures, 
> as well as, you can quickly iterate over a list in a high performance way,
> every push will return the internal offset of the value, also, the iterator `lrange`
> will tell you the next offset you can start from.  

- `LPUSH <LIST> <val1> [<val2> ...]` (push the item into the list "it doesn't check for uniqueness, it will append anyway (duplicate)")
- `LPUSHU <LIST> <val1> [<val2> ...]` (push the item into the list only if it isn't exists)
- `LRANGE <LIST> [<offset> <size>]`
- `LREM <LIST> [<val1> <val2> <val3> ...]` (deletes the list itself or values in the list)
- `LCOUNT <LIST>` (get the list members count)
- `LCARD <LIST>` (alias of `LCOUNT`)
- `LSUM <LIST>` (sum the members of the list "in case they were numbers")
- `LAVG <LIST>` (get the avg of the members of the list "in case they were numbers")
- `LMIN <LIST>` (get the minimum of the members of the list "in case they were numbers")
- `LMAX <LIST>` (get the maximum of the members of the list "in case they were numbers")
- `LSRCH <LIST> <NEEDLE>` (text-search using (string search or regex) in the list)
- `LSRCHCOUNT <LIST> <NEEDLE>` (size of text-search result using (string search or regex) in the list)

## # SET
- `SADD <LIST> <val1> [<val2> ...]` (alias of `LUPUSH`)
- `SMEMBERS <LIST> [<offset> <size>]` (alias of `LRANGE`)
- `SSCAN <LIST> [<offset> <size>]` (alias of `LRANGE`)
- `SCARD <LIST>` (aliad of `LCOUNT`)
- `SREM <LIST> [<val1> <val2> <val3> ...]` (alias of `LREM`)

## # Pub/Sub
> `Redix` has very simple pub/sub functionality, you can subscribe to internal logs on the `*` channel or any custom defined channel, and publish to any custom channel.

- `SUBSCRIBE [<channel1> <channel2>]`, if there is no channel specified, it will be set to `*`
- `PUBLISH <channel> <payload>`
- `WEBHOOKSET <channel> <httpurl>`, register a http endpoint so it can be notified through `JSON POST` request with the channel updates, this command will return a reference ID so you can manage it later.
- `WEBHOOKDEL <ID>`, stops listening on a channel using the above reference ID.
- `WEBSOCKETOPEN <channel>`, opens a websocket endpoint and returns its id, so you can receive updates through `ws://server.address:port/stream/ws/{generated_id_here}`
- `WEBSOCKETCLOSE <ID>`, closes the specified websocket endpoint using the above generated id. 

## # Ratelimit
- `RATELIMITSET <bucket> <limit> <seconds>`, create a new `$bucket` that accepts num of `$limit` of actions per the specified num of `$seconds`, it will returns `1` for success.
- `RATELIMITTAKE <bucket>`, do an action in the specified `bucket` and take an item from it, it will return `-1` if the bucket not exists or it has unlimited actions `$limit < 1`, `0` if there are no more actions to be done right now, `reminder` of actions on success.
- `RATELIMITGET <bucket>`, returns array [`$limit`, `$seconds`, `$remaining_time`, `$counter`] information for the specified bucket

## # Utils
> some useful utils that you can use within your app to remove some hassle from it.

- `ENCODE <method> <payload>`, encode the specified `<payload>` using the specified `<method>` (`md5`, `sha1`, `sha256`, `sha512`, `hex`)
- `UUIDV4`, generates a uuid-v4 string, i.e `0b98aa17-eb06-42b8-b39f-fd7ba6aba7cd`.
- `UNIQID`, generates a unique string.
- `RANDSTR [<size>, default size is 10]`, generates a random string using the specified length. 
- `RANDINT <min> <max>`, generates a random string between the specified `<min>` and `<max>`.
- `TIME`, returns the current time in `utc`, `seconds` and `nanoseconds`
- `DBSIZE`, returns the database size in bytes.
- `GC`, runs the Garbage Collector.
- `ECHO [<arg1> <arg2> ...]`
- `INFO`

TODO
=====
- [x] Basic Commands
- [x] Strings Commands
- [x] Hashmap Commands
- [x] List Commands
- [x] PubSub Commands
- [x] Utils Commands
- [x] Adding BoltDB engine
- [x] Adding LevelDB engine
- [x] Adding Null engine
- [x] Adding SQLite engine
- [ ] Adding TiKV engine
- [ ] Adding RAM engine
