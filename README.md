# go-http-kv

Simple HTTP server / key-value store written in Go.

## To build

go build

## To run as HTTP server

From your root directory:

```
go-http-kv -mode ws -listen 8080 -index index.html -root /path/to/root
```

This will run an HTTP server with root at current working directory.

## To run as KV store

```
go-http-kv -mode kv -listen 8080 -root /path/to/root
```

* To write "key"

```
curl -s -XPUT -d"value" http://localhost:8080/key
```

* To read "key"

```
curl -s http://localhost:8080/key
```

* To delete "key"

```
curl -s -XDELETE http://localhost:8080/key
```

