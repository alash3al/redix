Redix v5
========
> a tiny highly available key-value store that could be run on commodity servers.

TODOs
======
[x]- Default engine as boltdb
[x]- Basic engine contracts
[x]- Basic Redis Interface
[x]- Wal implementation on top of leveldb
[x]- Local state machine on top of leveldb
[x]- Expose wal scanning to main Redis interface
[x]- Only write to wal, then let any instance consume from it even the master
[ ]- Expose snapshot command let any replica to fetch a snapshot
[ ]- Expose a state command to let the master detect how much of wal should be trimmed