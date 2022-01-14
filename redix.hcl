server {
    redis {
        listen = ":6380"
        max_connections = 100
        async = true
    }
}

engine "postgresql" {
    dsn = "postgresql://postgres@localhost/redix"
}