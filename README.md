Redix v2 (Koka)
==================
> here is redix v2 (kokadb codename), it is the fully refactored, optimized and rich verion not only for redix v2, but also for redis itself!

Features
=========
- Multi core support.
- Uses Serializable Transactions.
- Optimizable for write-heavy workloads.
- Flixible Datastructure(s).

Supported Commands
==================
### SET
> `SET <key> <value> [<ttl_duration_string>]` 

**Examples**
- `SET key value`
- `SET key value 10s`
- `SET key value 10ms`
- `SET key value 22ns`
- `SET key value 1000us`
- `SET key value 2h30m`

#### GET
> `GET <key>`

#### DEL
> `DEL <key> [<key2>, ....]`

#### INCR
> `INCR <key> [<delta>] [<ttl_duration_string>]`

**Examples**
- `INCR key`
- `INCR key 5`
- `INCR key 8.24`
- `INCR key 1 1h30m10s`
