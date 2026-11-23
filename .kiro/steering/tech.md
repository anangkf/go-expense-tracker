# Tech Stack & Commands

- Go 1.25.1, Gin v1.11, GORM v1.31 + driver Postgres (pgx), `golang-jwt/jwt/v5`, `go-playground/validator/v10`, `bcrypt` (x/crypto), `godotenv`, `google/uuid`, swaggo (swag + gin-swagger + files), `gin-contrib/cors`.
- Database: PostgreSQL. Schema dibuat oleh `AutoMigrate` saat startup, lalu `Seed()` (idempotent via `FirstOrCreate`). Tidak ada file migrasi SQL.
- Sudah ada test untuk config, jwt, auth middleware, utils, model bucket type, dan repository (test DB dilewati tanpa `TEST_DATABASE_DSN`). Belum ada Dockerfile, Makefile, atau CI.

## Konfigurasi (`.env`, contoh di `.env.example`; `.env` ada di `.gitignore`, jangan di-commit atau dicetak)

`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`, `JWT_SECRET` (wajib), `JWT_EXPIRE_HOURS` (24), `JWT_REFRESH_SECRET` (wajib), `JWT_REFRESH_EXPIRE_HOURS` (168), `SERVER_PORT` (8080), `SERVER_MODE` (debug), `CORS_ALLOWED_ORIGINS` (daftar origin dipisah koma).

`.env.example` sudah memuat semua variabel di atas. Tidak ada fallback secret: `Config.Validate()` dipanggil di `main.go` dan app gagal start jika `JWT_SECRET` atau `JWT_REFRESH_SECRET` kosong. Jika `CORS_ALLOWED_ORIGINS` kosong, mode `debug` memakai `http://localhost:3000`, mode `release` gagal start. Level log GORM: `Info` di mode `debug`, `Warn` selain itu.

## Perintah

```bash
go run .                       # jalankan server (butuh Postgres aktif)
go build ./...                 # build semua package
go vet ./...
swag init                      # regenerate docs/ setelah mengubah anotasi swagger (butuh CLI swag)
```

`docs/` (docs.go, swagger.json, swagger.yaml) adalah hasil generate. Jangan edit manual.
