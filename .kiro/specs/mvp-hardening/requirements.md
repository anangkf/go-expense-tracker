# Requirements: MVP Hardening

Sumber: bagian "Belum ada / kurang untuk rilis MVP > Wajib sebelum dipakai user nyata" di `mvp.md`. Cakupan hanya item "Wajib". Item "Sebaiknya ada" dan "Di luar MVP" tidak termasuk.

## Keputusan yang diasumsikan (ubah jika tidak sesuai)

- Hapus category yang masih dipakai expense: **ditolak** dengan `409 Conflict`.
- Origin CORS dibaca dari env `CORS_ALLOWED_ORIGINS` (daftar dipisah koma).
- Level log GORM: `Info` bila `SERVER_MODE=debug`, selain itu `Warn`.
- Income pada dashboard dihitung dari semua category bertipe `income` (default + milik user), hanya untuk expense milik user.
- Tidak ada perubahan skema database. Tidak ada migrasi.

## Requirement

### R1. Cegah SQL injection pada sort dan filter
User story: sebagai pemilik API, saya ingin parameter `sortBy` dan filter key tidak bisa disisipi SQL.

Acceptance criteria:
1. WHEN `sortBy` tidak ada di allowlist endpoint tersebut THEN sistem memakai default `created_at`.
2. WHEN key filter tidak ada di allowlist endpoint tersebut THEN key itu diabaikan.
3. Allowlist ada per endpoint: expenses, expense-templates, categories, budget-plans.
4. Nama kolom yang masuk ke SQL hanya berasal dari allowlist, tidak pernah dari input mentah.
5. Request dengan `sortBy=id;DROP TABLE users--` tidak menyebabkan error SQL dan tidak mengubah data.

### R2. AuthMiddleware berhenti saat user tidak ditemukan
1. WHEN user dari token tidak ada THEN response 401 dan handler berikutnya tidak berjalan (`c.Abort()`).

### R3. Generate token mengembalikan error
1. WHEN `SignedString` gagal THEN `GenerateToken` dan `GenerateRefreshToken` mengembalikan error aslinya.
2. Pemanggil di `handlers/auth.go` menangani error itu dengan 500 dan tidak menyimpan refresh token.
3. `ValidateToken` dan `ValidateRefreshToken` menolak token yang signing method-nya bukan HS256.

### R4. Konfigurasi secret aman
1. `.env.example` memuat `JWT_REFRESH_SECRET`, `JWT_REFRESH_EXPIRE_HOURS`, dan `CORS_ALLOWED_ORIGINS`.
2. Fallback hardcoded untuk `JWT_SECRET` dan `JWT_REFRESH_SECRET` dihapus.
3. WHEN salah satu secret kosong THEN aplikasi gagal start dengan pesan jelas.
4. Log di `handlers/auth.go` yang mencetak `RefreshExpireHours` dihapus.

### R5. CORS dan log GORM untuk produksi
1. CORS hanya mengizinkan origin dari `CORS_ALLOWED_ORIGINS`. WHEN kosong THEN gagal start bila `SERVER_MODE=release`, dan memakai `http://localhost:3000` bila `debug`.
2. Level log GORM mengikuti keputusan di atas.

### R6. Perhitungan dashboard dan agregat benar
1. Income di `GetOverview` mencakup category default dan milik user.
2. Batas akhir rentang tanggal mencakup seluruh hari terakhir (`< endDate + 1 hari`), di dashboard, `GetTotalIncomeAndExpenses`, dan `GetTotalSpendingByBucket`.
3. Semua query yang memakai join manual memfilter `expenses.deleted_at IS NULL`. Join ke `categories` juga memfilter `categories.deleted_at IS NULL` kecuali ada alasan untuk menyertakan category terhapus (catat jika ada).
4. `this_week_spent` memakai awal hari (00:00) untuk awal minggu dan akhir minggu eksklusif (`< start + 7 hari`).

### R7. BucketTypeID pada budget plan valid
1. WHEN create atau update budget plan berisi `BucketTypeID` di luar 1-4 THEN 400.
2. WHEN update membuat atau mengubah bucket THEN `BucketTypeID` tersimpan.
3. Total persentase tetap harus 100.

### R8. BucketTypeID expense mengikuti category
1. WHEN `UpdateExpense` mengganti category ke category `expense` THEN `BucketTypeID` disalin dari category baru.
2. WHEN category baru bertipe `income` THEN `BucketTypeID` di-set `NULL`.

### R9. Category yang dipakai tidak bisa dihapus
1. WHEN masih ada expense (tidak terhapus) yang memakai category THEN `DELETE /categories/:id` mengembalikan 409 dan tidak menghapus.
2. WHEN tidak ada pemakaian THEN perilaku lama berlaku.

## Definisi selesai

- `go build ./...` dan `go vet ./...` bersih.
- Test baru untuk R1, R3, R6, R8, R9 lulus (lihat tasks).
- Anotasi swagger diperbarui untuk endpoint yang berubah (409 pada delete category, 400 pada budget plan), lalu `swag init` dijalankan.
- Checklist "Wajib" di `mvp.md` dicentang untuk item yang selesai.
