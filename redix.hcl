// NWEzMmZlNTktMDQ3OS00YTczLWEyNjItN2JhZTkwMTFkMzY3OmMzc24xZXA2MmpwcTQzb3FmcmMw

// any environment var here will be expanded first, e.x: `"${REDIX_DRIVER}"`
// engine represents the storage driver to be used
// engine = "postgres"
storage {
    driver = "postgres"
    
    connection {
        default = "postgres://postgres@localhost/redix"

        cluster {
            read = ["postgres://postgres@localhost/redix"]
            write = ["postgres://postgres@localhost/redix"]
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

    // TODO
    // used as a replacement for base64.encode
    // server_master_key = ""
}
