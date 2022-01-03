Redix v5
========
> a tiny highly available key-value store that could be run on commodity servers.

TODOs
======
- [x] Default engine as boltdb
- [x] Basic engine contracts
- [x] Basic Redis Interface
- [x] Wal implementation on top of leveldb
- [x] Local state machine on top of leveldb
- [x] Expose wal scanning to main Redis interface
- [x] Only write to wal, then let any instance consume from it even the master
- [x] Expose snapshot api let any replica to fetch a snapshot
    - [x] Let new replicas to resync all dump as well as the current master offset
- [ ] Expose a state api to let the master detect how much of wal should be trimmed
