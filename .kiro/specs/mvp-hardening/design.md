# Design: MVP Hardening

Mengikuti konvensi repo: package flat, tanpa interface, DI lewat constructor, komentar section huruf kapital.

## R1. Allowlist sort/filter

Pendekatan yang dipilih: allowlist dipegang repository (yang tahu nama kolom), bukan middleware. Middleware tetap generik.

- Tambah `utils/sanitize.go`:
  - `func SafeSort(sortBy string, allowed map[string]string, fallback string) string` mengembalikan ekspresi kolom dari map (key = nama query, value = kolom SQL ber-prefix tabel), atau `fallback`.
  - `func SafeFilters(filters map[string]string, allowed map[string]string) map[string]string` hanya mengembalikan key yang ada di map, dengan key diganti kolom SQL.
- Setiap repository mendefinisikan map allowlist sebagai variabel package, mis. di `expense_repository.go`:
  - expenses: `id, amount, description, created_at, updated_at, category_id, bucket_type_id` (prefix `expenses.`). `category_name` dan `category_type` tetap ditangani khusus seperti sekarang.
  - expense_templates, categories, budget_plans: kolom serupa sesuai model.
- `order` sudah divalidasi di middleware (asc/desc).
- Alternatif yang ditolak: allowlist di middleware. Middleware tidak tahu endpoint mana, sehingga butuh konfigurasi per route dan lebih rumit.

## R2, R3

- `middleware/auth.go`: tambah `c.Abort()` sebelum `return` pada cabang user tidak ditemukan.
- `services/jwt_services.go`: `return "", err`. Di parser, validasi `token.Method` dengan `*jwt.SigningMethodHMAC` dan `Alg() == "HS256"`, atau pakai `jwt.WithValidMethods([]string{"HS256"})`.
- `handlers/auth.go`: periksa error dari kedua Generate (Register, Login, Refresh), balas 500.

## R4, R5

- `config.LoadConfig()` tetap mengembalikan `*Config`. Tambah `func (c *Config) Validate() error`. `main.go` memanggil dan `log.Fatal` bila error. Validasi: `JWT.Secret` dan `JWT.RefreshSecret` tidak kosong, DB password tidak diperiksa.
- `getEnv` untuk dua secret dipanggil dengan default kosong.
- Tambah `ServerConfig.CORSAllowedOrigins []string` dari `CORS_ALLOWED_ORIGINS` (split koma, trim).
- `main.go`: `AllowOrigins: cfg.Server.CORSAllowedOrigins`.
- `database/connection.go`: pilih `logger.Info` bila mode debug, selain itu `logger.Warn`. Fungsi perlu menerima mode (dari `cfg.Server.Mode`).
- `.env.example` diperbarui. `.env` lokal tidak disentuh dan tidak dicetak.

## R6. Dashboard dan agregat

- Helper di `utils`: `EndOfDayExclusive(t time.Time) time.Time` (= awal hari t + 1 hari) dan `StartOfDay(t)`. Query memakai `created_at >= ? AND created_at < ?`.
- Titik pemanggilan yang sudah memberi `endDate` sebagai tanggal: handler dashboard, budget plan active, expense totals. Konversi dilakukan di repository agar semua pemanggil terlindungi. Periksa tiap pemanggil agar tidak menambah hari dua kali.
- `DashboardRepository.GetOverview`:
  - Hapus langkah `Pluck` category IDs. Income: join `categories` dengan `categories.type = 'income'`, tanpa filter kepemilikan category (kepemilikan dijaga lewat `expenses.user_id`).
  - Tambah `expenses.deleted_at IS NULL` pada semua query.
  - `startOfWeek` = awal hari (Minggu, sesuai kode saat ini; jangan ubah awal minggu), `endOfWeek` eksklusif = start + 7 hari.
- `GetTotalSpendingByBucket` (budget_plan_repository, join `e` di `LEFT JOIN`): tambah `e.deleted_at IS NULL` ke kondisi ON dan ubah rentang.
- `GetTotalIncomeAndExpenses` (expense_repository): tambah `expenses.deleted_at IS NULL`, ubah rentang.
- Query daily expenses di dashboard: periksa dan terapkan aturan yang sama.

## R7. Budget plan

- Di handler create/update budget plan, validasi tiap bucket: `BucketTypeID` ada di 1..4 (konstanta di `models`, mis. `ValidBucketTypeIDs`). Tidak perlu query DB karena ID seed tetap.
- Di update (transaksi create/update/delete bucket), sertakan `BucketTypeID` pada struct yang disimpan. Jika memakai `Updates(map)`/`Select`, tambahkan kolom `bucket_type_id`.
- Gunakan `binding`/`validate` tag `min=1,max=4` pada DTO bucket jika DTO-nya bisa menampung, sebagai tambahan.

## R8. Update expense

- Di `UpdateExpense`, ketika `CategoryID` berubah: ambil category baru (sudah dilakukan untuk cek kepemilikan), lalu set `expense.BucketTypeID = category.BucketTypeID` bila `category.Type == "expense"`, selain itu `nil`.
- Pastikan repository `Update` menyimpan kolom `bucket_type_id` termasuk nilai NULL (`Select("*")` atau `Select("bucket_type_id", ...)` karena `Save`/`Updates` struct melewati nilai nol/nil).

## R9. Hapus category

- Tambah `CategoryRepository.CountExpensesByCategoryID(categoryID uint) (int64, error)` (`expenses.deleted_at IS NULL` otomatis lewat GORM `Model(&models.Expense{})`).
- `DeleteCategory`: setelah cek kepemilikan, bila count > 0 balas `409` dengan pesan jelas. Anotasi swagger `@Failure 409`.
- Race antara cek dan hapus diterima untuk MVP (tidak ada FK constraint baru).

## Pengujian

Repo belum punya test dan Postgres test belum ada. Strategi:
- Unit test murni (tanpa DB): `utils/sanitize_test.go`, `services/jwt_services_test.go` (error signing, alg `none`/HS512 ditolak), `config` validate, helper tanggal.
- Test repository membutuhkan Postgres. Gunakan env `TEST_DATABASE_DSN`; test di-skip bila kosong (`t.Skip`). Cakup dashboard (income default, hari terakhir, deleted_at) dan hapus category.
- Handler/middleware: `httptest` + Gin untuk R2 (handler berikutnya tidak berjalan).
- Tidak menambah dependency baru, selain library test standar.

## Risiko

- Perubahan rentang tanggal mengubah angka yang sudah dilihat user. Ini koreksi yang disengaja.
- Gagal start tanpa secret memutus setup lokal yang mengandalkan fallback. Mitigasi: `.env.example` lengkap.
- Mengabaikan filter key tidak dikenal diam-diam mengubah hasil list bagi klien yang memakainya. Dapat diterima karena sebelumnya key tak valid menghasilkan error SQL.
