# Pocket App Backend - Final Report (Laporan Akhir Pengerjaan)
A robust backend REST API for a Pocket-like application, built with Go and Fiber. This service allows users to manage saved articles, videos, documents, and notes, complete with tagging, favorite toggles, dan searching capabilities.
## 1. Summary Hasil Pengerjaan
Proyek ini berhasil mengimplementasikan backend service untuk aplikasi mirip Pocket. Backend dibangun menggunakan arsitektur *Clean Architecture* (Layered: Handler/Controller, Usecase/Service, Repository). Secara keseluruhan, aplikasi ini sudah mendukung Autentikasi (JWT), Manajemen Pocket Items (CRUD lengkap dengan pagination, filter komprehensif, pencarian full-text, pengurutan, dll), dan Dashboard statistik. 
Selain fungsionalitas, proyek ini dilengkapi dengan mekanisme otomatisasi database (auto-migrate dan auto-seed) yang mempermudah proses setup (pengguna hanya perlu menjalankan server dan backend akan siap digunakan), serta dokumentasi dan struktur tes yang terorganisir dengan baik.
## 2. Fitur yang Berhasil dan Belum Berhasil Diimplementasikan
### Fitur yang Berhasil Diimplementasikan:
- **Authentication**: Login dengan JWT dan middleware untuk proteksi endpoint.
- **Pocket Items CRUD Lengkap**:
  - `POST /api/pockets`: Pembuatan item baru dengan validasi field wajib.
  - `GET /api/pockets`: Menampilkan list item dengan dukungan fitur advanced:
    - Pagination (page, limit).
    - Sorting berdasarkan `created_at`, `updated_at`, dan `title`.
    - Filtering berdasarkan `status`, `favorite`, `type`, dan `tags` (bisa multi-tag dengan logika pencarian array JSON).
    - Pencarian teks (Full-Text Search) yang cepat pada field `title`, `url`, dan `description`.
  - `GET /api/pockets/:id`: Detail dari sebuah pocket item.
  - `PUT /api/pockets/:id`: Update keseluruhan data item (termasuk tag).
  - `PATCH /api/pockets/:id/status` dan `PATCH /api/pockets/:id/favorite`: Update field spesifik dengan payload minimal untuk operasi cepat (misal saat membaca artikel di frontend).
  - `DELETE /api/pockets/:id`: Mengarsipkan item (Soft-Delete secara logika dengan mengubah status ke `archived`, tidak dihapus permanen).
- **Dashboard API**: Menampilkan metrik agregat yang lengkap (total item, item difavoritkan, read vs unread status, distribusi per tipe konten), list item terbaru, serta agregasi/distribusi item berdasarkan tag.
- **Validasi Input**: Validasi payload request dengan menggunakan Go Playground Validator (memastikan URL valid, tags harus unik).
### Fitur yang Belum Berhasil / Belum Diimplementasikan (Out of Scope):
- **User Registration (Register API)**: Saat ini pembuatan user bergantung pada bawaan Seeder untuk mempercepat demonstrasi (karena flow aslinya cukup rumit jika ada verifikasi email).
- **Social Login (Google/Github OAuth)**: Belum diimplementasikan dan butuh integrasi ke OAuth provider.
- **Auto-Extract Content (Web Scraping)**: Mengambil metadata (title, deskripsi, konten artikel teks penuh) dari sebuah link URL secara otomatis di-background belum dibuat. Saat ini pengguna masih harus mengisi `title` dan `description` manual via POST.
## 3. Keputusan Teknis Penting
- **Penyimpanan Tags sebagai JSON Array**: Memutuskan untuk menyimpan field `tags` dalam format `JSON` string array secara langsung pada kolom tabel `pocket_items` (karena MySQL mendukung JSON dan operasi search JSON). Hal ini sangat menyederhanakan arsitektur database daripada membuat relasi `Many-to-Many` dengan tabel tags terpisah. Performanya cukup untuk ratusan ribu record di awal.
- **Pencarian Teks menggunakan MySQL FULLTEXT INDEX**: Pencarian dengan operator standard `LIKE '%keyword%'` sangat lambat untuk data teks besar. Sebagai solusinya, ditambahkan indeks Full-Text pada field `title`, `url`, dan `description`. Pencarian dieksekusi dengan GORM `MATCH() AGAINST()`.
- **Repository Pattern dan Interface Abstraction**: Semua modul dipisahkan berdasarkan layer (Handler -> Usecase -> Repository). Interface digunakan untuk dependency injection. Keputusan ini terbukti ampuh karena ketika pembuatan *Unit Test*, kita dapat melakukan *mocking* Repository dengan sangat mudah menggunakan `Testify Mock` (tidak perlu hit DB saat test Usecase).
## 4. Tradeoff dan Known Issues
### Tradeoff:
- **Ketergantungan Spesifik ke Fitur MySQL**: Untuk filter `tags` dan fitur search, *raw query* yang dipakai sangat spesifik menggunakan syntax MySQL (`JSON_SEARCH` dan `MATCH AGAINST`). Konsekuensinya, proyek ini tidak dapat dipindahkan (drop-in) langsung ke engine database lain seperti PostgreSQL atau SQLite murni tanpa refactoring code repository layer (testing in-memory SQLite untuk beberapa fungsi query spesifik tidak akan berjalan 100% mulus). Oleh karena itu, kita sarankan MySQL sebagai target pengujian.
### Known Issues:
1. **Warning Duplicate Key Error (1061) pada Auto-Migrate**: Saat GORM menjalankan fungsi Auto-Migration dan membuat custom Full-text index, jika index tersebut sudah ada, mungkin akan menampilkan `Error 1061: Duplicate key name`. Hal ini telah dimitigasi dan di-*catch* pada code agar tidak *fatal* sehingga server dapat terus menyala.
## 5. Hasil Testing
Aplikasi telah diuji menggunakan berbagai jenis pengetesan dan menghasilkan *Code Coverage* yang stabil dan komprehensif. Laporan detil dari Test dapat dilihat pada file dokumentasi [TESTING.md](TESTING.md).
- **Unit Testing**: Telah ditulis pengujian unit dengan mocking untuk seluruh lapisan Use Case (Business Logic), dan *table-driven testing* untuk handler layer logic. Testing dapat dieksekusi cepat tanpa database fisik.
- **Integration Setup**: GORM telah di-setup untuk memudahkan operasional tes jika menggunakan database test spesifik.
- Status Test terakhir adalah **PASSED** via `go test ./... -v`.
## 6. Cara AI Digunakan (Proses Pair-Programming AI)
Selama proyek, AI (Antigravity IDE / Gemini Pro) digunakan sebagai *Co-Pilot / Pair Programmer* utama yang memiliki wewenang untuk mengeksekusi shell commands, menulis, merombak, dan me-review code:
1. **Analisis Requirement (Translating PRD to Tech Docs)**: AI memformulasikan dokumentasi Analisa PRD, Design DB, dan API Contract agar developer mengerti ekspektasi dari permintaan sistem secara detail.
2. **Boilerplate & Layered Architecture**: AI melakukan inisialisasi skeleton aplikasi (Fiber, GORM) dari nol dan menstrukturisasinya menjadi `cmd`, `internal/`, dan `pkg/`.
3. **Problem Solving Kompleks**: Saat implementasi filter multi-tag dengan tipe JSON Array dan implementasi Fulltext search di GORM, AI mendesain SQL Query Builder dengan aman (mencegah SQL Injection).
4. **Automated Test Generation**: AI dimanfaatkan untuk mengeksplor kemungkinan edge-case dan men-generate unit test block dan mock repository yang panjang (yang sangat melelahkan jika ditulis manual).
5. **Debugging & Refactoring Code**: AI mengeksekusi proses refactor untuk pemisahan *Business Logic* dari controller ke usecase, menyelesaikan problem duplicate payload error, serta migrasi.
## 7. Link Recording Proses Pengerjaan
*(Silakan isi atau ganti link berikut dengan tautan video presentasi/recording hasil pengerjaan Anda)*
[**Video Recording Proses Pengerjaan**](https://drive.google.com/file/d/1Iqbf6kNlyaohzEnMoN5hKZ2H_z97YpdR/view?usp=drive_link)
## 8. Improvement Plan (Rencana Pengembangan)
- **Implementasi Caching**: Mengintegrasikan **Redis** pada endpoint `/api/dashboard` dan `/api/pockets`. Dashboard memerlukan *heavy computation* ketika jumlah row database menjadi sangat besar, sehingga agregat metric (seperti per-tag info) sebaiknya dicache.
- **Message Broker & Background Worker (RabbitMQ / Kafka)**: Jika kedepannya kita ingin mengambil Thumbnail atau men-*scrape* konten teks asli dari URL (sehingga pengguna dapat membacanya secara offline tanpa iklan), *task* ini dapat di-*offload* ke worker asynchronous.
- **Dockerization dan CI/CD**: Menambahkan file `Dockerfile`, `docker-compose.yml`, serta Github Actions script untuk standarisasi environment deployment otomatis dan unit test check setiap ada *Pull Request*.
- **Meilisearch / Elasticsearch integration**: Jika data text dari jutaan artikel disimpan di database, MySQL Fulltext bisa kehabisan nafas. Menggunakan specialized Search Engine (seperti Elastic) adalah solusi natural berikutnya