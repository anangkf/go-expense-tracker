# Product: Expense Tracker API

REST API (Go) untuk pencatatan pengeluaran/pemasukan pribadi dengan budget plan berbasis "bucket" (mis. aturan 50/30/20). Module Go: `go-expense-tracker-api`. Klien utamanya frontend web (CORS dibuka `*`).

## Konsep domain

- **User**: punya email unik, password bcrypt, dan `ActiveBudgetPlanID` (nullable).
- **Category**: `type` = `expense` atau `income`. Default category (`is_default=true`, `user_id NULL`) di-seed dan dipakai semua user. Category custom milik satu user. Category expense punya `BucketTypeID`.
- **BucketType** (seed, ID tetap): 1 Needs, 2 Wants, 3 Savings, 4 Investments.
- **Expense**: tabel `expenses` menyimpan pengeluaran DAN pemasukan. Pembeda ada di `category.type`. `BucketTypeID` disalin dari category hanya saat create dan hanya jika category bertipe `expense`.
- **ExpenseTemplate**: preset cepat (nama, kategori, icon). Default di-seed (Makan, Parkir, Internet, Ojol, Bayar Kos, Kopi, Belanja, Menabung).
- **BudgetPlan**: berisi `Buckets` (nama, `BucketTypeID`, `Percentage`). Total persentase harus tepat 100. `IsTemplate=true` + `user_id NULL` untuk template seed ("50/30/20 Rule", "4/3/2/1 Rule").
- **RefreshToken**: disimpan di DB per `jti`, bisa di-revoke.

## Aturan bisnis penting

- Auth: access token (HS256, default 24 jam, `JWT_SECRET`) dan refresh token (HS256, default 168 jam, `JWT_REFRESH_SECRET`) memakai `jti` yang sama. Refresh bersifat rotasi: token lama di-revoke, pasangan baru dibuat. Logout me-revoke refresh token berdasarkan `jti` dari access token.
- Active budget plan: `GET /budget-plans/active` menghitung `max_allocation = percentage/100 * total_income` per bucket, dan `total_spending` per bucket dari expense yang `bucket_type_id`-nya cocok. Default rentang tanggal = bulan berjalan; override dengan `expense_start_date` dan `expense_end_date`.
- Dashboard: `overview` (total_spent, this_week_spent, left_in_budget = income - spent) dan `daily_expenses`. Query `start_date`/`end_date` format `YYYY-MM-DD`, default bulan berjalan.
- Kepemilikan data: expense/template/category/budget plan hanya bisa diakses pemiliknya. Category default boleh dipakai semua user tapi tidak bisa diubah/dihapus user.
- Soft delete (`gorm.DeletedAt`) dipakai di semua model utama.

## Endpoint (base `/api/v1`, Swagger di `/api/v1/swagger/index.html`, health di `/health`)

| Grup | Route |
|---|---|
| auth (publik) | `POST /auth/register`, `/auth/login`, `/auth/refresh-token`; `POST /auth/logout` (butuh token) |
| user | `GET /user/profile` |
| budget-plans | `GET /templates`, `GET ""`, `GET /active`, `POST ""`, `POST /:id/set-active`, `GET/PUT/DELETE /:id` |
| categories | `GET ""`, `GET /default`, `POST ""`, `POST /multiple`, `GET/PUT/DELETE /:id` |
| expenses | `GET ""`, `POST ""`, `GET/PUT/DELETE /:id` |
| expense-templates | `GET ""`, `GET /default`, `POST ""`, `POST /multiple`, `GET/PUT/DELETE /:id` |
| dashboard | `GET /dashboard` |

Response seragam: `{success, message, data, error}`. List terpaginasi memakai `data: {data, total, page, limit, total_pages}`.
