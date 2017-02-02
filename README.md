# go-redistest 

go-redistest controls Redis server instance to be used in unit tests for Golang.
You don't need to start instances outside the unit test code manually.

### Features

* Replication
* Sentinels

### Requirement

[Redis](https://github.com/antirez/redis) must be installed and ```redis-server``` needs to be in your ```$PATH```.

### Install

Install redistest package:

```bash
go get github.com/masatomizuta/go-redistest
```

Import it in your application:

```go
import "github.com/masatomizuta/go-redistest
```

## Usage

### Run a master server

```go
master, err := redistest.RunServer(6379)
if err != nil {
    panic(err)
}
defer master.Stop()
// Do your tests using Redis server
```

### Run a slave server

```go
slave, err := master.RunSlaveServer(6380)
if err != nil {
    log.Fatal(err)
}
defer slave.Stop()
```

### Run a sentinel server

```go
sentinel, err := master.RunSentinelServer(26379, "mymaster")
if err != nil {
    log.Fatal(err)
}
defer sentinel.Stop()
```

## Todo

* Wipe DB and conf file
* More settings

## License

[MIT License](LICENSE)
