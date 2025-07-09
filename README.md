# go-clacks-overhead

Drop-in solution to add the [X-Clacks-Overhead](http://www.gnuterrypratchett.com/) header value `GNU Terry Pratchett` to all http requests.

# client

```Go
req := http.NewRequest(http.MethodGet, "www.example.com", http.NoBody)

res, err := clacks.DefaultClient.Do(req)
```

# server

```Go
r := mux.NewRouter()

r.Use(clacks.Middleware())

r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("OK"))
}).Methods("GET")

log.Fatal(http.ListenAndServe(":8080", r))
```
