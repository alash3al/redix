server {
    redis {
        listen = ":6380"
        max_connections = 100
    }
}

engine "postgresql" {
    dsn = "postgresql://postgres@localhost/redix"
}