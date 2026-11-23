# MVP: Expense Tracker API

Status berdasarkan scan kode (bukan hasil menjalankan aplikasi). Tujuan MVP: user bisa mencatat pengeluaran/pemasukan, mengelompokkannya ke bucket budget, dan melihat sisa budget.

## Sudah ada (fitur MVP)

- [x] Registrasi, login, refresh token (rotasi), logout
- [x] Profil user
- [x] Category default (seed) dan category custom per user (CRUD, bulk create)
- [x] Expense/income CRUD dengan filter, sort, dan pagination
- [x] Expense template default (seed) dan custom (CRUD, bulk create)
- [x] Budget plan: template seed (50/30/20, 4/3/2/1), buat plan sendiri, update bucket (create/update/delete dalam satu transaksi), set active, hapus
- [x] Active budget plan dengan `max_allocation`, `total_spending`, `spending_percentage` per bucket
- [x] Dashboard: total spent, this week spent, left in budget, daily expenses
- [x] Swagger docs, health check, CORS

## Belum ada / kurang untuk rilis MVP

Wajib sebelum dipakai user nyata:
- [ ] Perbaiki SQL injection pada `sortBy` dan filter key (allowlist kolom)
- [ ] Tambah `c.Abort()` di `AuthMiddleware` saat user tidak ditemukan
- [ ] Perbaiki `GenerateToken`/`GenerateRefreshToken` agar mengembalikan error
- [ ] Lengkapi `.env.example` (`JWT_REFRESH_SECRET`, `JWT_REFRESH_EXPIRE_HOURS`) dan hapus fallback secret hardcoded (gagal start jika kosong)
- [ ] Batasi CORS ke origin frontend, turunkan log level GORM di produksi
- [ ] Perbaiki perhitungan dashboard (income dari category default, batas akhir hari, filter `deleted_at`)
- [ ] Validasi `BucketTypeID` pada budget plan, dan simpan saat update bucket
- [ ] Update `BucketTypeID` expense saat category berubah
- [ ] Cegah hapus category yang masih dipakai expense (atau tentukan perilakunya)

Sebaiknya ada:
- [ ] Test (minimal handler auth, expense, budget plan active, dan repository dashboard)
- [ ] Dockerfile + docker-compose (app + Postgres)
- [ ] Migrasi SQL versioned menggantikan `AutoMigrate` untuk produksi
- [ ] Status code konsisten (expense memakai 401 untuk bukan pemilik, seharusnya 403/404)
- [ ] Tipe uang: `float64` untuk `amount` rawan pembulatan, pertimbangkan integer (satuan terkecil) atau decimal
- [ ] CI sederhana (`go build`, `go vet`, `go test`)

## Di luar MVP (ide lanjutan)

- Tipe transaksi eksplisit (`income`/`expense`) terpisah dari tabel `expenses`
- Laporan per kategori (struct `CategorySplit` sudah ada tapi belum dipakai di endpoint)
- Transaksi berulang, multi-currency, export CSV
- Ubah password, hapus akun, revoke semua sesi (`RevokeAllByUserID` sudah ada di repository tapi belum dipakai)
- Rate limiting pada endpoint auth

## Definisi selesai MVP

1. Semua item "Wajib" di atas tertutup.
2. Alur utama lolos end-to-end: register -> pilih template budget -> set active -> catat income dan expense -> lihat active plan dan dashboard dengan angka benar.
3. Aplikasi bisa dijalankan dari nol dengan satu perintah (compose) dan `.env.example`.
