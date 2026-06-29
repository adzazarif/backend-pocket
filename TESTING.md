# Pocket App - Testing Documentation

Dokumen ini menjelaskan strategi pengujian otomatis (_Automated Testing_) pada backend Pocket App. Sistem pengujian kami mencakup pengujian pada lapisan _Service_ (Unit Test), lapisan _Handler_ (Unit Test), serta pengujian API secara keseluruhan (End-to-End Integration Test).

## Cakupan Testing (_Test Coverage_)

Sistem testing ini dibagi ke dalam beberapa fase pengujian:

1. **Unit Test (Service Layer):** Menguji langsung logika bisnis dengan melakukan *mock* pada *Database Repository*.
2. **Unit Test (Handler Layer):** Menguji pengolahan *Request/Response* HTTP (menggunakan `gofiber`) dan *Error Handling*, dengan memotong ketergantungan pada *Service*.
3. **Integration Test (End-to-End):** Menguji alur hidup *Request* dari awal ke akhir menggunakan _In-Memory Database_ (SQLite) tanpa memerlukan server database asli.

---

## 🚀 Cara Menjalankan Test

Pastikan Anda berada di direktori `backend/` sebelum menjalankan perintah-perintah berikut.

### 1. Menjalankan Semua Test
Untuk mengeksekusi seluruh pengujian (Unit dan Integrasi) sekaligus, jalankan:
```bash
go test ./... -v
```

### 2. Melihat *Coverage* Keseluruhan
Untuk melihat persentase baris kode yang sudah di-cover oleh pengujian:
```bash
go test ./... -cover
```

### 3. Membuat Laporan *Coverage* Visual (HTML)
Jika Anda ingin melihat secara visual baris kode mana yang belum tertutup pengujian:
```bash
# Hasilkan file profile coverage
go test ./... -coverprofile=coverage.out

# Buka file profile di browser (HTML)
go tool cover -html=coverage.out
```

---

## 🏗 Struktur Testing

### 1. Mocking (*testify/mock*)
Setiap dependensi di-*mock* agar unit test dapat berjalan secara terisolasi. Mock ini berada pada direktori domain masing-masing:
- `internal/domain/auth/repository_mock.go`
- `internal/domain/pocket/repository_mock.go`
- `internal/domain/dashboard/repository_mock.go`

### 2. Service Layer Unit Tests (`service_test.go`)
Testing di layer ini memvalidasi *business rules* seperti:
- Pengembalian `apperror.NotFound` jika item tidak ditemukan.
- Keberhasilan mengumpulkan data (seperti penjumlahan *Dashboard Summary*).
- Kalkulasi dan manipulasi entitas sebelum diteruskan ke Repository (contoh: fungsi `Archive` mengubah _status_ menjadi `archived`).

### 3. Handler Layer Unit Tests (`handler_test.go`)
Testing di layer ini mengisolasi *routing* Fiber:
- **Validasi Input:** Memastikan _payload_ tidak valid akan dicegat dan mengembalikan kode status HTTP `422 Unprocessable Entity` secara otomatis via mekanisme fiber error.
- **Isolasi Rute:** Setiap `t.Run` akan membuat _instance_ `fiber.App` baru. Ini mencegah adanya tabrakan rute (Error: `404 Cannot GET /`) yang biasa terjadi pada _global setup_ di framework Fiber.

### 4. Integration Tests (`tests/integration_test.go`)
Integration Test (*End-to-End API Test*) kami menggunakan in-memory database **SQLite**.
- **Penting:** Kami menggunakan versi driver SQLite murni golang (`github.com/glebarez/sqlite`) untuk menghindari masalah dependensi kompiler CGO (khususnya di OS Windows).
- **Kompatibilitas:** Tipe data kolom *database* dirancang agnostik (contoh: mengganti pemakaian fungsi `ENUM` dan memanipulasi *casting* array JSON kustom untuk `StringArray`) agar seluruh sintaks bekerja secara paralel di MySQL (Produksi) dan SQLite (Testing).

#### Skenario Integrasi (Simulasi Alur Pengguna)
Test ini akan menyimulasikan kejadian berurutan berikut di dalam database in-memory:
1. Menjalankan *Auto-Migrate* tabel.
2. Membuat pengguna (`Seed User`).
3. Melakukan *Login* (`POST /api/auth/login`) dan mengambil _JWT Token_.
4. Menambahkan item saku baru (`POST /api/pockets`) dengan melampirkan _Header Authorization_ (`Bearer <token>`).
5. Menguji pengiriman _payload_ kosong/salah untuk mendapatkan respons validasi HTTP 422.
6. Membaca *Dashboard* (`GET /api/dashboard`) dan memastikan statistik baru sudah ter-*update*.
7. Menguji *Error Handling* Unauthorized (mengakses rute *protected* tanpa token).

---

## 🗂 Daftar Skenario Testing per Endpoint

Berikut adalah daftar keseluruhan endpoint yang diuji, skenario spesifik untuk *Unit Test*, dan ekspektasi kode status (HTTP Status Code).

### 1. Auth & Authentication Endpoints

| Endpoint | Method | Skenario Uji (Test Case) | Ekspektasi Status | Keterangan |
| :--- | :--- | :--- | :--- | :--- |
| `/api/auth/login` | POST | Success Login | `200 OK` | Kredensial valid, token JWT diterbitkan. |
| `/api/auth/login` | POST | Validation Error | `422 Unprocessable Entity` | Format email salah atau field kosong. |
| `/api/auth/login` | POST | Invalid Credential | `401 Unauthorized` | Email/password tidak cocok atau salah. |
| `/api/auth/logout` | POST | Success Logout | `200 OK` | Bearer Token valid, logout sukses. |

### 2. Dashboard Endpoints

| Endpoint | Method | Skenario Uji (Test Case) | Ekspektasi Status | Keterangan |
| :--- | :--- | :--- | :--- | :--- |
| `/api/dashboard` | GET | Success Get Summary | `200 OK` | Mengembalikan ringkasan jumlah status pocket item dan 5 item terbaru. |
| `/api/dashboard` | GET | Unauthorized | `401 Unauthorized` | Request tanpa menyertakan JWT di *Header*. |

### 3. Pocket Items Endpoints

| Endpoint | Method | Skenario Uji (Test Case) | Ekspektasi Status | Keterangan |
| :--- | :--- | :--- | :--- | :--- |
| `/api/pockets` | POST | Success Create | `201 Created` | Berhasil membuat pocket item baru. |
| `/api/pockets` | POST | Validation Error | `422 Unprocessable Entity` | Title kosong atau Content Type tidak valid. |
| `/api/pockets` | GET | Success List | `200 OK` | Menampilkan paginasi daftar item dengan metadata. |
| `/api/pockets/:id` | GET | Success Get Detail | `200 OK` | Mengambil detail 1 pocket item spesifik. |
| `/api/pockets/:id` | GET | Not Found | `404 Not Found` | ID tidak ditemukan atau item bukan milik pengguna (`POCKET_NOT_FOUND`). |
| `/api/pockets/:id` | PUT | Success Update | `200 OK` | Memperbarui isi metadata/url pocket item. |
| `/api/pockets/:id` | DELETE | Success Archive | `200 OK` | Mengubah *status* item menjadi `archived`. |
| `/api/pockets/:id/status` | PATCH | Success Update Status | `200 OK` | Mengubah status item (misal ke `reading` atau `read`). |
| `/api/pockets/:id/favorite` | PATCH| Success Toggle Favorite | `200 OK` | Mengubah *flag* `is_favorite` (true/false). |

---

## ✅ Indikator Keberhasilan

Ketika pengujian berhasil (seperti pada saat penulisan dokumen ini), _output console_ akan menghasilkan hal-hal berikut:

```text
=== RUN   TestAPIIntegration
=== RUN   TestAPIIntegration/1._Login_Success
=== RUN   TestAPIIntegration/2._Create_Pocket_Item_Success
=== RUN   TestAPIIntegration/3._Validation_Error_(Create_Pocket_without_Title)
=== RUN   TestAPIIntegration/4._Check_Dashboard_Summary
=== RUN   TestAPIIntegration/5._Error_Handling_-_Unauthorized_(No_Token)
--- PASS: TestAPIIntegration (0.25s)
...
PASS
coverage: 75.0% of statements
ok      pocket-app/internal/domain/auth 5.166s
...
```

Tidak boleh ada satupun pesan `FAIL` atau `panic` pada keseluruhan laporan. Kegagalan dapat menandakan ada kontrak antarmuka API (_Payload Contract_) yang rusak atau inkonsistensi struktur database.
