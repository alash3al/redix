Redix v5
========
> `redis` is very simple, sometimes we abuse it, so decided to build a pure `key-value` storage system that introduces the core utilities for building any data structure you want, that why it has no specific commands for hashmaps, lists, sets, ... etc, it is all about keys & value nothing else!

Core Commands
=============

- PING
- QUIT
- FLUSHALL
- FLUSHDB
- SELECT <DB index>
- SET <key> <value> [EX seconds | KEEPTTL] [NX]
- TTL <key>
- GET <key>
- INCR <key> [<delta>]
- DEL key [key ...]
- HGETALL <prefix>
    > our magic command that treats with the whole database as if it were a single nested large hash-table




