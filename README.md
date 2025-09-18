# go-clacks-overhead

[![codecov](https://codecov.io/github/rjkward/go-clacks-overhead/graph/badge.svg?token=2LH2EIXNLC)](https://codecov.io/github/rjkward/go-clacks-overhead)
[![Go Report Card](https://goreportcard.com/badge/github.com/rjkward/go-clacks-overhead)](https://goreportcard.com/report/github.com/rjkward/go-clacks-overhead)
[![Release](https://img.shields.io/github/v/release/rjkward/go-clacks-overhead.svg)](https://github.com/rjkward/go-clacks-overhead/releases)

Drop-in solution to add the [X-Clacks-Overhead](http://www.gnuterrypratchett.com/) header value `GNU Terry Pratchett` to all http requests.

# install

```bash
go get github.com/rjkward/go-clacks-overhead
```

# server

```go
r := mux.NewRouter()

r.Use(clacks.Middleware())

r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("OK"))
}).Methods(http.MethodGet)

log.Fatal(http.ListenAndServe(":8080", r))
```

# client

```go
req, _ := http.NewRequest(http.MethodGet, "http://www.example.com", http.NoBody)

res, _ := clacks.DefaultClient.Do(req)
```

Alternatively use the transport in your custom client:

```go
customClient := &http.Client{
    Transport: clacks.DefaultTransport,
}
```
