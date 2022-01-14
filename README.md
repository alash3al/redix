Redix v5
========
> `redix` is a very simple `key => value` storage engine that speaks redis and even more simpler and flexible.

Why did I build this?
======================
> `redis` is very simple, sometimes we abuse it, so I decided to build a pure `key-value` storage system that introduces the core utilities for building any data structure you want based on the `key => value` model that is because I think that everything could be modeled easily using that model, so I decided to not to follow redis and all of its commands, you won't find `lpush`, `hset`, `sadd`, ... etc you will find a new way to do the same job but more easier and flexable, i.e, the well-known `hset key field value` command could be replaced with `set key/field value`, but sometimes you need to return a specific hashmap as `key => value`, but you run `hget key field` to get the key's value and also it could be replaced with `get key/field`, but how could we replace `hgetall key`? I will say "it is easy", let's make the `hget` command work as a prefix scanner that scan the whole database using the specified prefix and return all `key => value` pairs as redis hashmap response!, so `HGETALL` in redix means scan and return the result as `hashmap`

Features
==========
- A really simple `key => value` store that speaks `redis` protocol but with our rules!.
- A real system that you can abuse! it isn't intedented for cache only but a "database system".
- `Async` (all writes happen in the background), or `Sync` it won't respond to the client before writing to the internal datastore.
- Pluggable storage engines, currently it supports `postgresql`, and there nay be more engines introduces in the upcomning releases.


Core Commands
=============
- `PING`
- `QUIT`
- `FLUSHALL`
- `FLUSHDB`
- `SELECT <DB index>`
- `SET <key> <value> [EX seconds | KEEPTTL] [NX]`
- `TTL <key>`
- `GET <key> [DELETE]`, it has an alias for backward compatibility reasons called `GETDEL <key>`
- `INCR <key> [<delta>]`, it has an alias for backward compatibility reasons called `INCRBY`
- `DEL key [key ...]`
- `HGETALL <prefix>`
    > Fetches the whole data under the specified prefix as a hashmap result
    ```bash
        $ 127.0.0.1:6380> set /users/u1 USER_1
        OK

        $ 127.0.0.1:6380> set /users/u2 USER_2
        OK

        $ 127.0.0.1:6380> set /users/u3 USER_3
        OK

        $ 127.0.0.1:6380> hgetall /users/
        1) "u1"
        2) "USER_1"
        3) "u2"
        4) "USER_2"
        5) "u3"
        6) "USER_3"
        ## in the hgetall response, redix removed the prefix you specified `/users/`
    ```
**NOTE**
- Any command starts with `#` will be just as if it were a comment, and `redix` will response with "OK"