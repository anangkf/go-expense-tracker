# Tech Stack & Commands

- Go 1.25.1, Gin v1.11, GORM v1.31 + driver Postgres (pgx), `golang-jwt/jwt/v5`, `go-playground/validator/v10`, `bcrypt` (x/crypto), `godotenv`, `google/uuid`, swaggo (swag + gin-swagger + files), `gin-contrib/cors`.
- Database: PostgreSQL. Schema dibuat oleh `AutoMigrate` saat startup, lalu `Seed()` (idempotent via `FirstOrCreate`). Tidak ada file migrasi SQL.
- Belum ada test, Dockerfile, Makefile, atau CI.

## Konfigurasi (`.env`, contoh di `.env.example`; `.env` ada di `.gitignore`, jangan di-commit atau dicetak)

`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`, `JWT_SECRET`, `JWT_EXPIRE_HOURS` (24), `JWT_REFRESH_SECRET`, `JWT_REFRESH_EXPIRE_HOURS` (168), `SERVER_PORT` (8080), `SERVER_MODE` (debug).

`.env.example` belum memuat `JWT_REFRESH_SECRET` dan `JWT_REFRESH_EXPIRE_HOURS`. Tanpa itu, fallback `your-refresh-secret-key` dipakai (tidak aman untuk produksi). Fallback `JWT_SECRET` juga hardcoded (`your-secret-key`).

## Perintah

```bash
go run .                       # jalankan server (butuh Postgres aktif)
go build ./...                 # build semua package
go vet ./...
swag init                      # regenerate docs/ setelah mengubah anotasi swagger (butuh CLI swag)
```

`docs/` (docs.go, swagger.json, swagger.yaml) adalah hasil generate. Jangan edit manual.
