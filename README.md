Redix v3
========
> Redix v3 is an key/value store that speaks redis protocol, but it moved from cloning redis commands to solving real life problems.

Where is v2?
============
> When I started redix, I was trying to clone redis but using disk storage engines instead of memory only, after a long time, there were some use cases where I wanted to use some distributed sotrage engines like postgres/mysql/cassandra, so I had to refactor some parts that may introduce a major update. But to be honest each day I was asking myself "Do we really need a redis clone or we need to solve some critical"

Thoughts
========
- Core key/value operations like: put/delete/scan.
- Abbility to subscribe to data changes.
- Should speak redis protocol (at least for now).
- Don't accept key's value if it is not changed.
- Abbility to extend/add custom redis commands at least for basic operations without core change, may be via (js, lua, ... etc).
- Abbility to have multiple storage drivers.
- Must implement an in-memory driver to store data in memory.
- Must provide the abbility to choose the default driver and pipe everything to secondary drivers.
- Imagine each redix instance as a BIG HASH TABLE, there is no namespace/dbs, why? 
    to force you as a developer to separate the concerns of your apps.
- 