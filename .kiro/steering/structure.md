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

## Hal yang perlu diwaspadai (temuan dari scan, belum diperbaiki)

Keamanan:
- `repositories/*`: `queryParams.SortBy` dan key filter dari query string disambung langsung ke SQL (`Order(...)`, `Where(key+" ILIKE ?")`). Rentan SQL injection dan error pada kolom tak valid. Perlu allowlist kolom.
- `middleware/auth.go`: jika user tidak ditemukan, response 401 dikirim tanpa `c.Abort()`, sehingga handler tetap berjalan.
- `services/jwt_services.go`: saat `SignedString` gagal, fungsi mengembalikan `"", nil` (error hilang).
- Access token tidak dicek terhadap status revoke; hanya refresh token yang disimpan dan di-revoke. `ValidateToken` tidak memeriksa signing method.
- Refresh token disimpan plaintext di DB. `handlers/auth.go` Login mencetak `RefreshExpireHours` ke log.
- CORS `AllowOrigins: *`, dan GORM log level `Info` (mencetak semua SQL).

Kebenaran data:
- `BudgetPlan` create: `BucketTypeID` bucket tidak divalidasi. Pada update, `BucketTypeID` tidak ikut disimpan untuk bucket yang dibuat/diupdate, dan bucket dicocokkan ke pengeluaran berdasarkan `bucket_type_id`.
- `GetTotalSpendingByBucket` dan `GetTotalIncomeAndExpenses` memakai `BETWEEN tanggal AND tanggal` tanpa jam akhir hari, sehingga data di hari terakhir bisa terlewat, dan tidak memfilter `expenses.deleted_at`. Beberapa query dashboard juga tidak memfilter `deleted_at` (Joins manual pada GORM tidak otomatis memfilter tabel join).
- `DashboardRepository.GetOverview`: income dihitung hanya dari category milik user (`user_id = ?`), jadi income dari category default (mis. Salary) tidak terhitung. `this_week_spent` memakai `time.Now()` apa adanya (bukan awal hari), dan `endOfWeek` = start + 6 hari.
- `UpdateExpense` tidak memperbarui `BucketTypeID` saat category berubah.
- `Expense`/`ExpenseTemplate` tag `gorm:"primaryKey, autoIncrement"` memakai spasi setelah koma (tag tidak valid sempurna). Handler mengembalikan 401 untuk "bukan milik user" pada expense (category/template memakai 403).
- `Category.Delete` tidak mengecek expense yang memakai category tersebut.
- `go.mod` menandai `gin-contrib/cors` dan `google/uuid` sebagai `// indirect` padahal dipakai langsung.
