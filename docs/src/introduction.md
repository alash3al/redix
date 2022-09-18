# Introduction
> **redix** is a hashtable service with multiple storage engines as backend.  
> 
> **redix** uses RESP protocol (Redis serialization protocol), you can learn more about it 
> [from here](https://redis.io/docs/reference/protocol-spec/), you can learn more about redis 
> [from here](https://redis.io/).
> 
> **redix** won't implement all redis features, but the core generic features 
> (the basic flat key-value model) that enables anyone to implement redis like features.
> 
> **redix** server tries to use all available cores, so it can provide the maximum machine 
> utilization as possible as it can to give you the best performance.

# Why Redix?
> [redis](https://redis.io/), is an in-memory data store, it is perfect
> as well, forces you to map everything to `key - value` which sometimes simplifies
> the thinking regarding the job you're doing, by the time, redis is abused by us
> due to that simplicity then we started to face new issues especially the
> memory and the single-threaded model.  
> 
> How could we store data on disk (read/write from/to disk)?
> How could we utilize the existing storage solutions but with simple redis interface?
> for sure there are many questions out there, most of them aren't in redis scope.  
> 
> That is why I created redix, which exposes redis-like interface but
> the datasource is different which could be anything (for now it is postgresql and filesystem).
> 
