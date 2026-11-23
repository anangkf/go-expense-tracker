# Struktur & Konvensi Kode

## Layout (layered, flat packages)

```
main.go            Wiring manual: config -> DB -> repos -> handlers -> routes. Anotasi swagger global.
config/            LoadConfig() dari env.
database/          connection.go (DSN, pool, memanggil AutoMigrate+Seed), migrate.go, seeds.go
models/            Struct GORM + DTO request/response (validate tag) per entitas
repositories/      Akses data GORM. Satu struct per entitas, constructor NewXRepository(db)
handlers/          Handler Gin. Satu struct per entitas, constructor NewXHandler(repos...)
middleware/        auth.go (JWT), pagination_filter.go (parse page/limit/sort/filter)
services/          jwt_services.go (generate/validasi token)
utils/             response.go (SuccessResponse/ErrorResponse), request.go (QueryParams, GetUserIDFromContext), hash.go
docs/              Output swag (generated)
```

Alur request: route -> `AuthMiddleware` (set `user_id`, `user_email`, `jti` di context) -> `PaginationAndFilter` (set `queryParams`) -> handler -> repository.

## Konvensi

- Tidak ada interface; dependency injection lewat constructor dengan tipe konkret.
- Handler: ambil `user_id` dari context, `ShouldBindJSON`, `validator.Struct(req)`, cek kepemilikan, panggil repo, balas lewat `utils.SuccessResponse`/`ErrorResponse`.
- Setiap handler diberi anotasi swaggo (`@Summary`, `@Router`, `@Security BearerAuth`, dst). Update anotasi saat mengubah endpoint, lalu jalankan `swag init`.
- Komentar section memakai huruf kapital (`// APPLY FILTERS`).
- Nama field JSON snake_case. Model dengan soft delete memakai `json:"-"` untuk `DeletedAt`.
- List endpoint: query `page` (1), `limit` (10), `sortBy` (`created_at`), `order` (`desc`), `start_date`, `end_date`, `expense_start_date`, `expense_end_date`. Query param lain dianggap filter kolom (`ILIKE %value%`). Khusus expenses/templates: `category_name`, `category_type`.
- Seed memakai ID tetap untuk BucketType (1-4) dan default Category (1-9). Seed expense template dan budget plan merujuk ID tersebut, jadi jangan ubah ID.

## Hal yang perlu diwaspadai (sisa temuan scan, belum diperbaiki)

Sudah ditangani oleh spec `mvp-hardening`: allowlist sort/filter (`utils.SafeSort`/`SafeFilters`), `c.Abort()` di `AuthMiddleware`, error `SignedString` dikembalikan, parser JWT hanya menerima HS256, log `RefreshExpireHours` dihapus, CORS dari `CORS_ALLOWED_ORIGINS`, log GORM mengikuti `SERVER_MODE`, perhitungan dashboard/agregat (income dari category default, rentang `>= start AND < akhir hari + 1`, filter `deleted_at`), validasi dan penyimpanan `BucketTypeID` budget plan, update `BucketTypeID` expense saat category berubah, dan hapus category terpakai (409). `go.mod` juga sudah tidak menandai `gin-contrib/cors` dan `google/uuid` sebagai indirect.

Keamanan:
- Access token tidak dicek terhadap status revoke; hanya refresh token yang disimpan dan di-revoke.
- Refresh token disimpan plaintext di DB.

Kebenaran data:
- `Expense`/`ExpenseTemplate` tag `gorm:"primaryKey, autoIncrement"` memakai spasi setelah koma (tag tidak valid sempurna). Handler mengembalikan 401 untuk "bukan milik user" pada expense (category/template memakai 403).
- Seed: default category Savings (ID 7) memakai `bucket_type_id` 1 (Needs), padahal seharusnya 3 (Savings). Template "Menabung" memakai bucket 3, jadi tidak konsisten dengan category-nya.
