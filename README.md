Redix v5
========
> `redix` is a very simple `key => value` storage engine that speaks redis and even more simpler and flexible.

Why did I build this?
======================
> `redis` is very simple, sometimes we abuse it, so I decided to build a pure `key-value` storage system that introduces the core utilities for building any data structure you want, that why redix has no specific commands for hashmaps, lists, sets, ... etc, it is all about keys & value nothing else!

Redix isn't
=============
- Full redix drop-in replacement
- Very fast data writes
- Caching Datastore, but you can use it as caching engine if you want

Redix is
==========
- Simple `key => value` storage that speaks redis protocol.
- A database, you can store any size of data till your postgres db be down, or till your disk free size is about to be zero.
- Ready to be abused.
- Nested Large Hash-table.
- 

Core Commands
=============

- `PING`
- `QUIT`
- `FLUSHALL`
- `FLUSHDB`
- `SELECT <DB index>`
- `SET <key> <value> [EX seconds | KEEPTTL] [NX]`
- `TTL <key>`
- `GET <key>`
- `INCR <key> [<delta>]`
- `DEL key [key ...]`
- `HGETALL <prefix>`
    > Fetches the whole data under the specified prefix as a hashmap result
    ```bash
        127.0.0.1:6380> set /users/u1 USER_1
        OK
        127.0.0.1:6380> set /users/u2 USER_2
        OK
        127.0.0.1:6380> set /users/u3 USER_3
        OK
        127.0.0.1:6380> hgetall /users/
        1) "u1"
        2) "USER_1"
        3) "u2"
        4) "USER_2"
        5) "u3"
        6) "USER_3"
    ```

