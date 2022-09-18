# Configurations
> redix is using a configuration language called [hcl](https://github.com/hashicorp/hcl),
> here is a configurations example named as (redix.hcl):
```c
// server block contains the configurations related to redix server
server {
  // redix is modular, "we can have multiple interfaces not only redis interface"
  // currently the supported interface is redis interface.
  redis {
    // which [address]:portNumber to let the server listen on
    listen = ":6380"

    // maximum number of connections allowed to the server instance in the same time
    max_connections = 100

    // whether to let the writes be async (done in background) or not?
    async = false
  }
}

// redix is modular, "we can have multiple storage engines to store the data"
// currently the supported engines are "postgresql" and "filesystem".
// in case you want to connect to "postgresql":
engine "postgresql" {
  // data-source-name regarding postgresql server configurations
  dsn = "postgresql://postgres@localhost/redix"
}

// in case you want "filesystem" to be your backend:
//engine "filesystem" {
//  // data-source-name: the directory to store the data files in
//  dsn = "./data/"
//}
```