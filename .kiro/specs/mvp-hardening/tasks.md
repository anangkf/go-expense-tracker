# Tasks: MVP Hardening

Urutan dibuat dari yang paling kritis dan independen. Setiap task diakhiri `go build ./... && go vet ./...`.

- [x] 1. Auth dan token (R2, R3)
  - [x] 1.1 Tambah `c.Abort()` di `middleware/auth.go` pada cabang user tidak ditemukan
  - [x] 1.2 `services/jwt_services.go`: kembalikan `err` dari `SignedString`; pakai `jwt.WithValidMethods([]string{"HS256"})` pada kedua parser
  - [x] 1.3 `handlers/auth.go`: tangani error Generate di Register/Login/Refresh dengan 500; hapus log `RefreshExpireHours`
  - [x] 1.4 Test `services/jwt_services_test.go`: token valid lolos, token `alg=none` dan HS512 ditolak, secret salah ditolak
  - [x] 1.5 Test `middleware/auth_test.go` (httptest): user tidak ada -> 401 dan handler berikutnya tidak dipanggil

- [x] 2. Konfigurasi, CORS, log (R4, R5)
  - [x] 2.1 `config/config.go`: hapus fallback secret, tambah `CORSAllowedOrigins`, tambah `Validate()`
  - [x] 2.2 `main.go`: panggil `Validate()` dengan `log.Fatal`; pakai origin dari config; fallback `http://localhost:3000` di debug, gagal di release bila kosong
  - [x] 2.3 `database/connection.go`: level log GORM sesuai `SERVER_MODE`
  - [x] 2.4 Perbarui `.env.example` (`JWT_REFRESH_SECRET`, `JWT_REFRESH_EXPIRE_HOURS`, `CORS_ALLOWED_ORIGINS`)
  - [x] 2.5 Test `config/config_test.go` untuk `Validate()` (secret kosong gagal, terisi lolos)
  - [ ] 2.6 Verifikasi manual: app gagal start tanpa `JWT_SECRET`, start normal dengan `.env` lengkap

- [ ] 3. Allowlist sort/filter (R1)
  - [ ] 3.1 Buat `utils/sanitize.go` (`SafeSort`, `SafeFilters`) dan `utils/sanitize_test.go` (termasuk payload injeksi)
  - [ ] 3.2 Terapkan di `expense_repository.go` (list dan `Order`) dengan allowlist kolom expenses
  - [ ] 3.3 Terapkan di `expense_template_repository.go`
  - [ ] 3.4 Terapkan di `category_repository.go` (dua fungsi list: `GetByUserID`, `GetByUserIDWithTotal`)
  - [ ] 3.5 Terapkan di `budget_plan_repository.go`
  - [ ] 3.6 Verifikasi manual: `?sortBy=id;DROP TABLE users--` dan `?foo=bar` pada keempat endpoint list tidak error dan hasil tetap terpaginasi

- [ ] 4. Perhitungan dashboard dan agregat (R6)
  - [ ] 4.1 Tambah helper tanggal di `utils` (`StartOfDay`, `EndOfDayExclusive`) beserta test
  - [ ] 4.2 `dashboard_repository.go`: income dari semua category income, `deleted_at`, rentang, `this_week_spent`; daily expenses ikut aturan sama
  - [ ] 4.3 `budget_plan_repository.go` `GetTotalSpendingByBucket`: `e.deleted_at IS NULL` dan rentang
  - [ ] 4.4 `expense_repository.go` `GetTotalIncomeAndExpenses`: `deleted_at` dan rentang
  - [ ] 4.5 Audit pemanggil (handler dashboard, budget plan active) agar tidak double-add hari
  - [ ] 4.6 Test repository (skip tanpa `TEST_DATABASE_DSN`): income dari category default terhitung, expense di hari terakhir terhitung, expense soft-deleted tidak terhitung

- [ ] 5. Budget plan BucketTypeID (R7)
  - [ ] 5.1 Konstanta `ValidBucketTypeIDs` di `models` dan fungsi validasi
  - [ ] 5.2 Terapkan validasi di create dan update budget plan (400 bila tidak valid)
  - [ ] 5.3 Simpan `BucketTypeID` pada bucket yang dibuat/diupdate di transaksi update
  - [ ] 5.4 Update anotasi swagger endpoint terkait
  - [ ] 5.5 Test: validasi menolak ID 0 dan 5; update menyimpan `BucketTypeID` (DB test, skip tanpa DSN)

- [ ] 6. Expense dan category (R8, R9)
  - [ ] 6.1 `UpdateExpense`: set `BucketTypeID` sesuai category baru (NULL untuk income); pastikan repository menyimpan nilai NULL
  - [ ] 6.2 `CategoryRepository.CountExpensesByCategoryID` dan pengecekan di `DeleteCategory` (409)
  - [ ] 6.3 Update anotasi swagger (`@Failure 409`)
  - [ ] 6.4 Test (DB test): ganti category expense->expense, expense->income, dan hapus category terpakai/tidak terpakai

- [ ] 7. Penutup
  - [ ] 7.1 Jalankan `swag init` (docs jangan diedit manual)
  - [ ] 7.2 Jalankan `go build ./... && go vet ./... && go test ./...`
  - [ ] 7.3 Uji alur end-to-end manual: register -> pilih template -> set active -> catat income dan expense -> cek `/budget-plans/active` dan `/dashboard`
  - [ ] 7.4 Centang item "Wajib" di `mvp.md` dan perbarui bagian "Hal yang perlu diwaspadai" di steering `structure.md`/`tech.md` yang sudah tertangani
