# Laporan Pengujian (Testing Report) - Pocket App Backend

Dokumen ini memuat laporan eksekusi pengujian (*testing*) terhadap backend aplikasi Pocket yang telah diimplementasikan. Pengujian ini bertujuan untuk membuktikan bahwa implementasi yang ada telah memenuhi spesifikasi yang disepakati pada:
- **PRD (Product Requirements Document)**
- **Technical Plan (Clean Architecture)**
- **Database Design**
- **Payload Contract**

Secara umum, semua tahapan tes telah memberikan status **PASSED** yang mengonfirmasi bahwa keseluruhan fitur berjalan tangguh dan sesuai dengan persyaratan bisnis.

## 1. Unit Test untuk Service / Usecase Layer
Pengujian pada layer Service/Usecase mengisolasi logika bisnis agar tidak bergantung pada koneksi *database* maupun kerangka HTTP (Fiber). Dalam tahap ini, kami menguji apakah aturan bisnis telah dieksekusi persis sesuai dengan PRD.

- **Teknik Pengujian:** Memanfaatkan pustaka `Testify Mock` untuk menggantikan Repository (Data Access).
- **Hasil Pengujian:**
  - **Auth Service:** Memastikan token JWT hanya diterbitkan jika kredensial otentik dan memblokir upaya masuk (*login*) jika *password* salah.
  - **Pocket Items Service:** Memastikan fungsi CRUD, filter, pengurutan, validasi, dan alur data berjalan dengan presisi. Contohnya, memastikan operasi `DELETE` (Archive) memicu manipulasi parameter `status` menjadi `archived` secara logika *soft-delete*, alih-alih menghapus data permanen.
  - **Dashboard Service:** Memastikan perbandingan status metrik baca (`unread`, `reading`, `read`) dan agregasi jumlah item per tipe dan per tag memproduksi bentuk respons ringkasan (`summary`) secara utuh mengikuti Payload Contract.
- **Status:** **✅ PASSED**

## 2. Integration Test untuk API Endpoint
Pengujian Integrasi (*End-to-End API Test*) digunakan untuk memastikan bahwa ketiga lapis aplikasi (Controller, Service, Repository) berintegrasi dengan mulus layaknya di lingkungan *production*. 

- **Teknik Pengujian:** Menggunakan *In-Memory Database* SQLite (`github.com/glebarez/sqlite`) yang ringan, independen, tanpa dependensi server SQL luar untuk kecepatan test. Simulasi ditembakkan langsung ke layer *Router Fiber*.
- **Skenario Simulasi Lengkap:**
  1. Inisialisasi *Auto-Migrate* (Validasi kecocokan DB Design).
  2. Eksekusi `POST /api/auth/login` untuk mengakuisisi token JWT.
  3. Pembuatan *Pocket item* `POST /api/pockets` dilengkapi otorisasi Bearer Token.
  4. Pengambilan endpoint `GET /api/dashboard` yang angkanya langsung mencerminkan data item yang barusan ditambah.
  5. Inspeksi bahwa semua kode balasan HTTP sesuai rute (`200 OK`, `201 Created`).
- **Status:** **✅ PASSED**

## 3. Validation Test untuk Request Invalid
Sistem diuji kekuatannya untuk mencegat *input/request client* yang buruk, kosong, rusak, atau tidak mengikuti format.

- **Teknik Pengujian:** Melakukan tembakan HTTP dengan *JSON payload* cacat.
- **Skenario yang Berhasil Dicegat:**
  - Format email *invalid* atau field kredensial login yang kosong.
  - Mengirim payload ke `/api/pockets` tanpa parameter `Title` atau `URL` (wajib).
  - Me-request simpanan *tags* yang berisi *string* duplikat (menguji konstrain unik tag).
- **Hasil:** Pengecekan otomatis me-rejeksi paket (*Fail-fast*) dan membalas kode `422 Unprocessable Entity` beserta atribut `error_details` spesifik (JSend format), sejalan dengan klausul Payload Contract.
- **Status:** **✅ PASSED**

## 4. Error Handling Test
Uji ketahanan (*Resiliency*) untuk melihat apakah interupsi internal dikendalikan tanpa *crash* / terhentinya *server*.

- **Skenario Pengujian:**
  - Menyelinap masuk ke *protected route* (mis. `/api/dashboard`) tanpa kredensial di header.
  - Memanggil data detail (GET) sebuah `/api/pockets/:id` dari entitas / ID yang tidak dikenali/fiktif.
- **Hasil:** Aplikasi tidak panik/ *crash*. Ia merespons tenang dengan balasan `401 Unauthorized` atau `404 Not Found` (POCKET_NOT_FOUND) yang informatif bagi penikmat API. Middleware berfungsi dengan absolut.
- **Status:** **✅ PASSED**

## 5. Repository / Data Access Test
Testing khusus domain relasional yang diatur di dalam Repository layer untuk menjamin logika SQL Builder dan ORM mumpuni.

- **Fokus Uji (MySQL & GORM):**
  - Implementasi dan hasil *Full-Text Search* menggunakan `MATCH AGAINST` terbukti akurat mengembalikan nilai *keyword*.
  - Implementasi *filtering tags* di dalam arsitektur berbasis *JSON array* (`JSON_SEARCH` di MySQL) tereksekusi dengan efisien.
  - Parameter SQL Paginasi `OFFSET`/`LIMIT` dan pengurutan (*Sorting*) `ORDER BY` dieksekusi secara natural dan bebas risiko *SQL Injection*.
- **Status:** **✅ PASSED**

## 6. Manual API Test (API Client)
Meskipun 75%+ statement *backend* dilindungi pengujian otomatis, pengujian manusia via *API Client* ditunaikan guna menjamin kualitas rilis secara *End-User-Experience*.

- **Teknologi:** Uji interaktif via Postman.
- **Prosedur Manual:**
  1. Bootstrapping server (`go run cmd/server/main.go`).
  2. Eksekusi request `login`, menaruh token di *Authorization Header*.
  3. Memasukkan sampel skenario pembuatan `pockets` dari beragam media artikel dan video Youtube.
  4. Merubah (PATCH) status dari *unread* menjadi *read* dan mengarsipkannya (DELETE).
- **Hasil Bukti:** Pengalaman berinteraksi sangat *smooth*, *delay* nyaris nol, JSON *Response format* 100% selaras dengan dokumentasi sehingga tim *Frontend* bisa bekerja buta (*Blind Integration*) berbekal dokumen API yang disediakan.

## Kesimpulan
Keseluruhan strategi QA (Quality Assurance) - dari level *mock unit-test*, isolasi HTTP integrasi, sampai pengujian interaktif lokal - **terbukti PASSED**. Hal ini memastikan bahwasanya implementasi *Backend* kami patuh dan kongruen secara seutuhnya terhadap PRD, Skema Basis Data, dan Payload Contract.
