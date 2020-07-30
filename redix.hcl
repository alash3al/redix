database 0 {
    dsn = "goleveldb://./"

    async {
        enabled = true
        queueSize = 1000
    }
}
