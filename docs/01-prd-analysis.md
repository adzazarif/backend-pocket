# 01 - PRD Analysis & Backend Requirement

## Overview

Dokumen ini berisi hasil analisis PRD **Pocket App** dari sudut pandang backend. Tujuannya adalah menerjemahkan requirement product menjadi kebutuhan teknis backend sebelum masuk ke fase implementasi.

---

## 1. Domain Backend

Berdasarkan analisis PRD, terdapat **2 domain utama** yang perlu diimplementasikan backend:

### 1.1 Authentication Domain
Mengelola sesi user, login, logout, dan proteksi route. Pada MVP ini menggunakan mock authentication, namun arsitektur harus mendukung migrasi ke real auth di kemudian hari.

### 1.2 Pocket Item Domain
Core domain aplikasi. Mengelola seluruh lifecycle pocket item mulai dari create, read, update, archive, search, filter, sort, toggle favorite, hingga update reading status.

---

## 2. Fitur yang Membutuhkan API

| Feature | Method | Endpoint | Priority |
|---|---|---|---|
| Login | POST | `/api/auth/login` | Must Have |
| Logout | POST | `/api/auth/logout` | Must Have |
| Dashboard summary | GET | `/api/dashboard` | Should Have |
| List pocket items | GET | `/api/pockets` | Must Have |
| Create pocket item | POST | `/api/pockets` | Must Have |
| Get pocket detail | GET | `/api/pockets/:id` | Must Have |
| Update pocket item | PUT | `/api/pockets/:id` | Must Have |
| Archive pocket item | DELETE | `/api/pockets/:id` | Must Have |
| Update reading status | PATCH | `/api/pockets/:id/status` | Should Have |
| Toggle favorite | PATCH | `/api/pockets/:id/favorite` | Should Have |

Total: **10 endpoint** untuk MVP.

---

## 3. Data yang Perlu Disimpan

### 3.1 Tabel `users`
Menyimpan data user yang terautentikasi.

| Field | Tipe | Keterangan |
|---|---|---|
| id | VARCHAR / UUID | Primary key |
| name | VARCHAR | Nama user |
| email | VARCHAR | Email unik, dipakai untuk login |
| password | VARCHAR | Hashed password |
| avatar_url | VARCHAR | Nullable |
| created_at | DATETIME | Auto |
| updated_at | DATETIME | Auto |

### 3.2 Tabel `pocket_items`
Menyimpan semua pocket item milik user.

| Field | Tipe | Keterangan |
|---|---|---|
| id | VARCHAR / UUID | Primary key |
| user_id | VARCHAR | Foreign key ke `users` |
| title | VARCHAR(120) | Wajib |
| url | VARCHAR | Nullable, conditional wajib |
| description | TEXT | Nullable, max 500 char |
| content_type | ENUM | `article`, `video`, `document`, `note` |
| status | ENUM | `unread`, `reading`, `read`, `archived` |
| is_favorite | BOOLEAN | Default false |
| created_at | DATETIME | Auto |
| updated_at | DATETIME | Auto |

### 3.3 Tabel `pocket_item_tags`
Relasi many-to-many antara pocket item dan tags (atau JSON column sebagai alternatif — lihat section 7).

| Field | Tipe | Keterangan |
|---|---|---|
| id | INT | Primary key |
| pocket_item_id | VARCHAR | FK ke `pocket_items` |
| tag | VARCHAR(24) | Nama tag |

> **Keputusan teknis:** Karena tag bersifat per-item dan tidak shared antar user/item, pendekatan **JSON column** (`tags JSON`) pada tabel `pocket_items` lebih simpel untuk MVP. Namun pendekatan tabel terpisah lebih proper untuk full-text search dan indexing. Keduanya valid — akan diputuskan di database design doc.

---

## 4. Business Rules yang Perlu Divalidasi Backend

Business rule berikut **wajib divalidasi di layer backend**, tidak boleh hanya client-side:

| ID | Rule | Implementasi Backend |
|---|---|---|
| BR-003 | Title wajib diisi | Validation middleware: `required, min=3, max=120` |
| BR-004 | URL wajib jika content_type bukan note | Conditional validation di service layer |
| BR-005 | URL harus format valid | Regex atau `url` validation tag |
| BR-009 | Tag tidak boleh duplicate dalam satu item | Deduplicate sebelum insert |
| BR-010 | Status default item baru = `unread` | Set default di service/repository layer |
| BR-011 | Favorite default = `false` | Set default di service/repository layer |
| BR-012 | Archived item tidak tampil di pocket list utama | Filter `status != archived` di query list |
| BR-015 | Search mencakup title, URL, description, tags | LIKE query atau FULLTEXT index pada field terkait |
| BR-016 | Filter dapat dikombinasikan dengan search | Query builder harus support dynamic WHERE clause |
| BR-017 | Sort default = `created_at DESC` | Default order di repository layer |
| BR-018 | Tidak boleh submit berkali-kali | Idempotency key atau handled di frontend; backend cukup proses satu request |
| BR-019 | Error API tidak boleh menghapus input user | Backend harus return error response yang jelas, bukan 200 dengan pesan gagal |
| BR-021 | Unauthorized user diarahkan ke login | Middleware auth return 401 |
| BR-022 | External link dibuka aman di tab baru | Backend tidak memproses ini — frontend concern |

### Business Rule yang Perlu Perhatian Khusus

**BR-004 (Conditional URL):**
```
IF content_type IN (article, video, document) THEN url REQUIRED
IF content_type == note THEN url OPTIONAL
```
Harus divalidasi di service layer, bukan hanya struct tag validation.

**BR-012 & BR-013 (Archived item):**
- List `/api/pockets` → hanya tampilkan `status != 'archived'`
- List `/api/archive` → hanya tampilkan `status = 'archived'`
- Satu tabel `pocket_items`, dibedakan oleh nilai `status`

**BR-002 (User isolation):**
Setiap query ke `pocket_items` harus menyertakan `WHERE user_id = <authenticated_user_id>`. Tidak boleh ada endpoint yang bisa mengakses item milik user lain.

---

## 5. Search, Filter, dan Sort

### 5.1 Search
Search dilakukan terhadap 4 field: `title`, `url`, `description`, dan `tags`.

**Strategi untuk MySQL:**
- Gunakan `LIKE '%keyword%'` untuk pendekatan sederhana (cukup untuk MVP dengan data kecil)
- Atau gunakan `FULLTEXT INDEX` dengan `MATCH ... AGAINST` untuk performa lebih baik pada data besar
- Untuk field `tags` (jika JSON column): gunakan `JSON_CONTAINS` atau `JSON_SEARCH`

### 5.2 Filter
Filter yang tersedia:

| Parameter | Nilai Valid |
|---|---|
| `status` | `unread`, `reading`, `read` (archived dipisah) |
| `type` | `article`, `video`, `document`, `note` |
| `favorite` | `true` / `false` |

Filter ini dikombinasikan secara dinamis dengan AND condition.

### 5.3 Sort
| Parameter | Behavior |
|---|---|
| `createdAt:desc` | ORDER BY created_at DESC (default) |
| `createdAt:asc` | ORDER BY created_at ASC |
| `title:asc` | ORDER BY title ASC |
| `title:desc` | ORDER BY title DESC |
| `updatedAt:desc` | ORDER BY updated_at DESC |

### 5.4 Pagination
PRD menyebutkan pagination sebagai `Could Have`. Akan diimplementasikan dengan query parameter `page` dan `limit`, response mengikuti `PaginatedResponse<T>` schema dari PRD.

---

## 6. API Response Contract

Berdasarkan PRD section 24, semua response mengikuti wrapper berikut:

**Success (single):**
```json
{
  "data": { ... },
  "message": "..."
}
```

**Success (list/paginated):**
```json
{
  "data": [ ... ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 25,
    "totalPage": 3
  }
}
```

**Error:**
```json
{
  "code": "ERROR_CODE",
  "message": "Human readable message",
  "details": [
    { "field": "title", "message": "Title is required" }
  ]
}
```

Error code yang perlu didefinisikan:

| Code | HTTP Status | Situasi |
|---|---|---|
| `VALIDATION_ERROR` | 422 | Input tidak valid |
| `INVALID_CREDENTIAL` | 401 | Login gagal |
| `UNAUTHORIZED` | 401 | Token tidak ada / expired |
| `POCKET_NOT_FOUND` | 404 | Item tidak ditemukan |
| `FORBIDDEN` | 403 | Akses item bukan miliknya |
| `INTERNAL_ERROR` | 500 | Error tak terduga |

---

## 7. Risiko dan Ambiguitas Requirement

### 7.1 Ambiguitas yang Ditemukan

| No | Area | Ambiguitas | Asumsi yang Diambil |
|---|---|---|---|
| 1 | Auth | PRD menyebut "mock authentication" tapi tidak mendefinisikan mock credential | Gunakan hardcoded mock user di seed data dengan email/password yang bisa dikonfigurasi via env |
| 2 | Token | PRD tidak menjelaskan mekanisme token (JWT, session, API key) | Gunakan **JWT** dengan expiry singkat untuk simulasi real auth |
| 3 | Tags | PRD tidak menentukan apakah tags disimpan sebagai JSON atau tabel terpisah | Gunakan **JSON column** untuk MVP (simpler query, cukup untuk data kecil) |
| 4 | Delete | PRD section 46 menyebutkan delete = archive/soft delete, tapi juga ada `DELETE /api/pockets/:id` | `DELETE` endpoint akan mengubah status menjadi `archived`, bukan hard delete |
| 5 | Search | PRD tidak menentukan apakah search server-side atau client-side | Implementasikan **server-side search** via query parameter `search` |
| 6 | Pagination | Status `Could Have`, namun struktur response sudah didefinisikan di PRD | Implementasikan pagination dari awal, lebih mudah daripada refactor |
| 7 | User ownership | PRD menyebutkan user hanya lihat item miliknya, tapi tidak ada multi-user flow | Gunakan `user_id` dari JWT claim untuk filter semua query |
| 8 | Archive page | PRD mendefinisikan `/archive` endpoint di frontend tapi tidak ada dedicated API untuk archive list | Gunakan query parameter `status=archived` di GET `/api/pockets` untuk archive list |

### 7.2 Scope yang Tidak Diimplementasikan (Out of Scope)

Sesuai PRD section 10.2:
- Real OAuth login
- File upload
- AI summary / auto tagging
- Auto metadata extraction dari URL
- Team collaboration / public sharing
- Offline mode / push notification

---

## 8. Asumsi Teknis

| No | Asumsi |
|---|---|
| 1 | Mock auth menggunakan JWT — token berisi `user_id`, `email`, `name` |
| 2 | Mock credential seed: `user@example.com` / `password123` |
| 3 | Tags disimpan sebagai JSON column di tabel `pocket_items` |
| 4 | `DELETE /api/pockets/:id` = soft delete (set status = archived) |
| 5 | Archive list menggunakan GET `/api/pockets?status=archived` |
| 6 | Search dilakukan server-side dengan LIKE query |
| 7 | Pagination default: `page=1`, `limit=10` |
| 8 | UUID digunakan sebagai primary key untuk semua entitas |
| 9 | Semua timestamp menggunakan UTC |
| 10 | Password di-hash menggunakan bcrypt (meski mock, praktik security tetap dijaga) |

---

## 9. Ringkasan Kebutuhan Backend

```
Backend Pocket App MVP
│
├── Auth Module
│   ├── POST /api/auth/login     → validate credential, return JWT
│   └── POST /api/auth/logout    → invalidate token (optional untuk mock)
│
├── Dashboard Module
│   └── GET  /api/dashboard      → aggregate count by status + favorite
│
└── Pocket Module
    ├── GET    /api/pockets           → list (search + filter + sort + paginate)
    ├── POST   /api/pockets           → create item
    ├── GET    /api/pockets/:id       → get detail
    ├── PUT    /api/pockets/:id       → update item
    ├── DELETE /api/pockets/:id       → archive (soft delete)
    ├── PATCH  /api/pockets/:id/status   → update reading status
    └── PATCH  /api/pockets/:id/favorite → toggle favorite
```

**Semua endpoint kecuali login dilindungi middleware JWT auth.**

---

*Dokumen ini menjadi dasar untuk `02-technical-plan.md`, `03-database-design.md`, dan `04-payload-contract.md`.*
