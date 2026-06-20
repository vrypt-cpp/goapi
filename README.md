# goapi

Starter Go API: plugin-based router + auto-generated OpenAPI dari struct Go via reflection. Tanpa framework eksternal, hanya `net/http` (Go 1.22+).

## Run

```
go run ./cmd/server
```

Server di `:8080`. Spec OpenAPI di `GET /openapi.json`.

## Struktur

```
core/        router, schema generator, openapi types, response helpers
plugins/     satu folder = satu plugin = satu domain
cmd/server   wiring plugin ke app
```

## Bikin plugin baru

1. Buat folder di `plugins/<nama>`.
2. Implement interface `core.Plugin`:

```go
type Plugin interface {
    Name() string
    Routes() []core.Route
}
```

3. Tiap `core.Route` definisikan method, path, request/response struct (untuk schema OpenAPI), dan handler `http.HandlerFunc`.
4. Daftarkan di `cmd/server/main.go` lewat `app.Use(yourplugin.New())`.

Request/response struct otomatis jadi schema di `/openapi.json`. Tag yang didukung:

- `json:"name,omitempty"` — nama field & optional
- `doc:"deskripsi"` — description
- `enum:"a,b,c"` — enum values
- `required:"false"` — paksa field jadi optional walau tanpa omitempty

Path param pakai syntax native Go 1.22 `{id}`, ambil value lewat `r.PathValue("id")`.

## Nambah middleware

`core.App.Handler()` return `http.Handler`, tinggal wrap biasa:

```go
http.ListenAndServe(":8080", loggingMiddleware(app.Handler()))
```
