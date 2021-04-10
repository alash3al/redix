// engine represents the storage driver to be used
engine = "postgres"

// here we define configure our main servers
// currently there is only one server which is: redis
server {
    redis {
        listen = ":6380"
    }
}

// connection block contains the read/write connection definations to be used
// redix uses the DSN style for connections.
// You must have at least one read & write dsn.
connection {
    read = [
        "postgres://postgres@localhost/tstdb"
    ]

    write = [
        "postgres://postgres@localhost/tstdb"
    ]
}
