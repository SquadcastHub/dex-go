# DEX for Go

DEX SDK is a simple middleware for HTTP routers which can record metrics like
status code, latency and memory on request completion and send the same to the Squadcast DEX servers.

## Example

The middleware function signature is very common and used across many popular routers for Go.
A simple example is included in the [example](example) folder


### [`go-chi`](https://github.com/go-chi/chi)

```go
r := chi.NewRouter()
d := dex.New("<Your API Key here>")
r.Use(d.Middleware)
```

### [`gorilla mux`](https://github.com/gorilla/mux)

```go
r := mux.NewRouter()
d := dex.New("<Your API Key here>")
r.Use(d.Middleware)
r.HandleFunc("/", handler)
```

## License

Apache-2.0 licensed. See the LICENSE file for details.