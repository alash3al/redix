# Installation

#### 1)- Binary installation

1. Goto the [releases page](https://github.com/alash3al/redix/releases).
2. Select the release the matches your OS (Linux or Mac).
3. Extract the binary file from the downloaded archive.
4. Rename the extracted binary file as `redix` and copy/move it to your `/usr/local/bin` or any folder included in your `$PATH` env-var.


#### 2)- Docker

1. `docker pull ghcr.io/alash3al/redix:latest`

# Running
> assuming you followed the corresponding installation instructions above,
> and created a configurations file called `redix.hcl` in the current working directory
> containing the configurations content found [here](https://github.com/alash3al/redix/blob/master/redix.hcl)
> and edited to match your preferences, for more info about configurations [click here](./configurations.md).

#### 1)- Binary
> assuming that you have a postgresql server running on the local machine contains a database called `redix`.
```bash
$ redix ./redix.hcl
```

#### 2)- Docker
> assuming that you have a postgresql container named `redixdb` having
> a database called `redix`, and the engine part configurations in `redix.hcl` as follows:
```hcl
engine "postgresql" {
     dsn = "postgresql://postgres@redixdb/redix"
}
```

```bash
docker run -v $(pwd)/redix.hcl:/etc/redix/redix.hcl --link redixdb -p 6380:6380 ghcr.io/alash3al/redix
```

# Connecting
> you can connect to redix using any redis client/library, here we will use the `redis-cli` command:

```bash
redis-cli -p 6380
```