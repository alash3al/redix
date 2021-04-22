// engine represents the storage driver to be used
// engine = "postgres"
storage {
    driver = "postgres"
    
    connection {
        default = "postgres://postgres@localhost/tstdb"

        cluster {
            read = ["postgres://postgres@localhost/tstdb"]
            write = ["postgres://postgres@localhost/tstdb"]
        }
    }
}

// modules is an array of modules shared ".so" files that should be loaded.
// each module must follow the core redix module structure.
modules = []

// here we define configure our main servers
// currently there is only one server which is: redis
server {
    // redis related configs
    redis {
        listen = ":6380"
    }
}
