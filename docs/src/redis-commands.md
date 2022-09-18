# Redis Commands
> we don't support all redis commands, we create our own redis by 
> implementing abstract commands only.  


- `PING`
- `QUIT`
- `FLUSHALL`
- `FLUSHDB`
- `SELECT <DB index>`
- `SET <key> <value> [EX seconds | KEEPTTL] [NX]`
- `TTL <key>` **(not supported while using `filesystem` engine)**
- `GET <key> [DELETE]`, it has an alias for backward compatibility reasons called `GETDEL <key>`
- `INCR <key> [<delta>]`, it has an alias for backward compatibility reasons called `INCRBY` **(not supported while using `filesystem` engine)**
- `DEL key [key ...]`
- `HGETALL <prefix>`, Fetches the whole data under the specified prefix as a hashmap
    ```bash
    $ 127.0.0.1:6380> set /users/u1 USER_1
    OK

    $ 127.0.0.1:6380> set /users/u2 USER_2
    OK

    $ 127.0.0.1:6380> set /users/u3 USER_3
    OK

    $ 127.0.0.1:6380> hgetall /users/
    1) "u1"
    2) "USER_1"
    3) "u2"
    4) "USER_2"
    5) "u3"
    6) "USER_3"
    ## in the hgetall response, redix removed the prefix you specified `/users/`
    ```
- `PUBLISH <channel|topic|anyword> <message here>`  **(not supported while using `filesystem` engine)**
- `SUBSCRIBE <channel|topic|anyword>`  **(not supported while using `filesystem` engine)**