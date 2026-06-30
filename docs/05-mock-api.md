# Rancangan Mock API (Mock Behavior) - Pocket App Backend

Dokumen ini mendeskripsikan rancangan *Mock API* berdasarkan *Payload Contract* yang disepakati. Mock API ini digunakan oleh tim Frontend (*Mobile/Web Client*) untuk melakukan integrasi dan memvalidasi kontrak secara paralel (*blind integration*) sebelum atau selama implementasi *Backend* benar-benar selesai dibangun.

## 1. Tujuan dan Penggunaan Mock API
Mock API berfungsi sebagai *Single Source of Truth* operasional di masa pengembangan awal. 
- **Validasi Contract:** Memastikan Frontend mengirim request body dan tipe data yang benar, serta mampu mengelola struktur kembalian (JSend format) dengan semestinya.
- **Development Paralel:** Frontend tidak terhambat (*blocking*) menunggu Backend selesai menulis logika database.
- **Tools Pendukung:** Mock ini bisa di-host menggunakan tools populer seperti **Postman Mock Server**, **Stoplight Prism**, **WireMock**, atau sekadar **JSON Server**.

---

## 2. Daftar Endpoint Mock Terlengkap

Berikut adalah semua endpoint yang di-mock berdasarkan kontrak:

| Endpoint | Method | Skenario Validasi | Keterangan |
| :--- | :--- | :--- | :--- |
| `/api/auth/login` | POST | Success, Invalid Credential | Autentikasi untuk mendapat dummy JWT Token. |
| `/api/auth/logout` | POST | Success Logout | Simulasi mencabut akses JWT token. |
| `/api/dashboard` | GET | Success, Empty State, Unauthorized | Menampilkan agregasi mock metrik untuk UI dashboard. |
| `/api/pockets` | GET | Success List, Empty State | Menampilkan paginasi dummy data *pocket items*. |
| `/api/pockets` | POST | Success Created, Validation Error | Mock payload penyimpanan link/artikel baru. |
| `/api/pockets/:id` | GET | Success Detail, Not Found | Simulasi mendapat 1 data sukses dan gagal (NotFound). |
| `/api/pockets/:id` | PUT | Success Update, Validation Error | Simulasi mengganti seluruh data item (Update total). |
| `/api/pockets/:id/status` | PATCH | Success Update Status | Simulasi mengubah status (`unread`, `reading`, `read`). |
| `/api/pockets/:id/favorite`| PATCH | Success Toggle Favorite | Simulasi mengubah boolean favorit. |
| `/api/pockets/:id` | DELETE | Success Archive | Simulasi menghapus (archive) item. |

---

## 3. Skenario & Contoh Data Dummy (API Lengkap)

### A. Auth Endpoints

#### 1. Login - Success (200 OK)
**Endpoint:** `POST /api/auth/login`
**Payload Request:** `{"email": "user@example.com", "password": "password123"}`
**Response:**
```json
{
  "status": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.dummy_mock_token.abcdef12345"
  }
}
```

#### 2. Login - Invalid Credential (401 Unauthorized)
**Endpoint:** `POST /api/auth/login`
**Response:**
```json
{
  "status": "error",
  "message": "Email atau password salah",
  "error_code": "INVALID_CREDENTIAL"
}
```

#### 3. Logout - Success (200 OK)
**Endpoint:** `POST /api/auth/logout`
**Response:**
```json
{
  "status": "success",
  "data": null
}
```

---

### B. Dashboard Endpoints

#### 1. Dashboard - Success (200 OK)
**Endpoint:** `GET /api/dashboard`
**Response:**
```json
{
  "status": "success",
  "data": {
    "total_items": 150,
    "total_favorites": 45,
    "status_summary": {
      "unread": 100,
      "reading": 10,
      "read": 40
    },
    "type_summary": {
      "article": 120,
      "video": 30
    },
    "recent_items": [
      {
        "id": 1,
        "title": "Tutorial: Get started with Go",
        "url": "https://go.dev/doc/tutorial/getting-started",
        "type": "article",
        "created_at": "2026-06-30T10:00:00Z"
      }
    ]
  }
}
```

---

### C. Pocket Items Endpoints

#### 1. Create Pocket Item - Success (201 Created)
**Endpoint:** `POST /api/pockets`
**Payload Request:**
```json
{
  "url": "https://example.com/mock-article",
  "title": "Mock Article",
  "description": "This is a mock description",
  "type": "article",
  "tags": ["mock", "test"]
}
```
**Response:**
```json
{
  "status": "success",
  "data": {
    "id": 101,
    "url": "https://example.com/mock-article",
    "title": "Mock Article",
    "description": "This is a mock description",
    "type": "article",
    "tags": ["mock", "test"],
    "status": "unread",
    "is_favorite": false,
    "created_at": "2026-06-30T10:05:00Z",
    "updated_at": "2026-06-30T10:05:00Z"
  }
}
```

#### 2. Create Pocket Item - Validation Error (422 Unprocessable Entity)
**Endpoint:** `POST /api/pockets`
**Payload Request:** `{"title": "Missing URL"}`
**Response:**
```json
{
  "status": "fail",
  "message": "Validasi payload request gagal",
  "error_code": "VALIDATION_ERROR",
  "error_details": [
    {
      "field": "url",
      "message": "Field 'url' is required and cannot be empty"
    }
  ]
}
```

#### 3. List Pocket Items - Success (200 OK)
**Endpoint:** `GET /api/pockets?page=1&limit=10`
**Response:**
```json
{
  "status": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "url": "https://go.dev",
        "title": "Golang Website",
        "description": "Go is an open source programming language",
        "type": "article",
        "tags": ["golang"],
        "status": "unread",
        "is_favorite": true,
        "created_at": "2026-06-30T10:00:00Z",
        "updated_at": "2026-06-30T10:00:00Z"
      }
    ],
    "meta": {
      "page": 1,
      "limit": 10,
      "total_items": 1,
      "total_pages": 1
    }
  }
}
```

#### 4. List Pocket Items - Empty State (200 OK)
**Endpoint:** `GET /api/pockets?search=not_found_keyword`
**Response:**
```json
{
  "status": "success",
  "data": {
    "items": [],
    "meta": {
      "page": 1,
      "limit": 10,
      "total_items": 0,
      "total_pages": 0
    }
  }
}
```

#### 5. Get Pocket Item Detail - Success (200 OK)
**Endpoint:** `GET /api/pockets/1`
**Response:**
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "url": "https://go.dev",
    "title": "Golang Website",
    "description": "Go is an open source programming language",
    "type": "article",
    "tags": ["golang"],
    "status": "unread",
    "is_favorite": true,
    "created_at": "2026-06-30T10:00:00Z",
    "updated_at": "2026-06-30T10:00:00Z"
  }
}
```

#### 6. Get Pocket Item Detail - Not Found (404 Not Found)
**Endpoint:** `GET /api/pockets/999`
**Response:**
```json
{
  "status": "fail",
  "message": "Pocket item tidak ditemukan",
  "error_code": "POCKET_NOT_FOUND"
}
```

#### 7. Update Pocket Item - Success (200 OK)
**Endpoint:** `PUT /api/pockets/1`
**Payload Request:**
```json
{
  "url": "https://go.dev",
  "title": "Golang Website (Updated)",
  "description": "Updated description",
  "type": "article",
  "tags": ["golang", "updated"]
}
```
**Response:**
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "url": "https://go.dev",
    "title": "Golang Website (Updated)",
    "description": "Updated description",
    "type": "article",
    "tags": ["golang", "updated"],
    "status": "unread",
    "is_favorite": true,
    "created_at": "2026-06-30T10:00:00Z",
    "updated_at": "2026-06-30T12:00:00Z"
  }
}
```

#### 8. Update Status - Success (200 OK)
**Endpoint:** `PATCH /api/pockets/1/status`
**Payload Request:** `{"status": "read"}`
**Response:**
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "status": "read",
    "updated_at": "2026-06-30T12:05:00Z"
  }
}
```

#### 9. Toggle Favorite - Success (200 OK)
**Endpoint:** `PATCH /api/pockets/1/favorite`
**Payload Request:** `{"is_favorite": false}`
**Response:**
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "is_favorite": false,
    "updated_at": "2026-06-30T12:10:00Z"
  }
}
```

#### 10. Delete (Archive) Pocket Item - Success (200 OK)
**Endpoint:** `DELETE /api/pockets/1`
**Response:**
```json
{
  "status": "success",
  "message": "Pocket item berhasil diarsipkan"
}
```

### D. Global Errors

#### 1. Unauthorized (401 Unauthorized)
**Endpoint:** *All Protected Endpoints* (e.g. `GET /api/dashboard` without token)
**Response:**
```json
{
  "status": "error",
  "message": "Token tidak valid atau sudah expired",
  "error_code": "UNAUTHORIZED_ACCESS"
}
```

---

## 4. Cara Mock Digunakan untuk Validasi Contract

Berikut adalah tata cara praktis penggunaan API Mock ini untuk proses pengembangan Frontend:

1. **Konfigurasi Postman Mock Server:**
   Developer membuat *Collection* di Postman berdasarkan JSON di atas, kemudian menekan tombol **"Mock Collection"**. Postman akan meng-generate URL *cloud* (seperti `https://<id>.mock.pstmn.io`).
2. **Setup Environment Frontend:**
   Frontend cukup mengganti variabel environment mereka (mis. `.env.development` pada aplikasi React/Flutter) dari `API_BASE_URL=http://localhost:3000` menjadi `API_BASE_URL=https://<id>.mock.pstmn.io`.
3. **Trigger Edge-Case Menggunakan Headers:**
   Untuk menguji UI state khusus seperti validasi form merah (*validation error*) tanpa harus repot menyalahkan format input manual, Frontend dapat menyisipkan HTTP header bawaan mock server (misal `x-mock-response-code: 422`). Mock server akan otomatis membalas respons error sehingga tim UI bisa mengetes tampilan error message secara instan.
4. **Validasi Payload Skema Otomatis (Prism Tooling):**
   Jika menggunakan mock tools otomatis berbayar seperti **Stoplight Prism**, alat tersebut dapat membaca spesifikasi Open API YAML. Ketika tim FE mem-*post* tipe data yang salah (misalnya `is_favorite` diisi dengan `"false"` string bukan boolean `false`), Prism akan memblokirnya di layer *mock*, memastikan bahwa kode di aplikasi FE sudah "type-safe" sebelum benar-benar ditembak ke *Backend* produksi.
