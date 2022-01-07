server {
    redis {
        listen = ":4000"
        max_connections = 100
    }
}

engine "postgresql" {
    dsn = "postgresql://postgres@localhost/redix"
}