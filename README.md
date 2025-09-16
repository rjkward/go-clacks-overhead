# go-clacks-overhead

Drop-in solution to add the [X-Clacks-Overhead](http://www.gnuterrypratchett.com/) header value `GNU Terry Pratchett` to all http requests.

# install

```
go get github.com/rjkward/go-clacks-overhead
```

# server

```Go
r := mux.NewRouter()

r.Use(clacks.Middleware())

r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("OK"))
}).Methods(http.MethodGet)

log.Fatal(http.ListenAndServe(":8080", r))
```

# client

```Go
req, _ := http.NewRequest(http.MethodGet, "http://www.example.com", http.NoBody)

res, _ := clacks.DefaultClient.Do(req)
```

Alternatively use the transport in your custom client:

```Go
customClient := &http.Client{
    Transport: clacks.DefaultTransport,
}
```
