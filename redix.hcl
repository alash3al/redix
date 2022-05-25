server {
    redis {
        listen = ":6380"
        max_connections = 100
        async = false
    }
}

// engine "postgresql" {
//     dsn = "postgresql://postgres@localhost/redix"
// }

engine "filesystem" {
    dsn = "./data/"
}
