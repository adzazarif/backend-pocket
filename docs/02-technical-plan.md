# 02 - Technical Plan

## Overview

Dokumen ini mendefinisikan rencana teknis backend Pocket App secara lengkap. Seluruh keputusan arsitektur, stack, struktur folder, dan strategi implementasi didokumentasikan di sini sebagai acuan tunggal sebelum coding dimulai.

---

## 1. Tech Stack

### 1.1 Core Stack

| Layer | Technology | Versi | Alasan Pemilihan |
|---|---|---|---|
| Language | Go | 1.22+ | Statically typed, performant, cocok untuk REST API, ekosistem mature |
| Web Framework | Fiber v2 | v2.52+ | API mirip Express, performa tinggi (fasthttp), middleware ecosystem lengkap |
| ORM / Query Builder | GORM | v2 | Mature, mendukung MySQL, migration, hooks, dan eager loading |
| Database | MySQL | 8.0+ | Relational, ACID compliant, FULLTEXT index untuk search |
| Authentication | JWT (golang-jwt) | v5 | Stateless auth, mudah di-extend ke real auth |
| Validation | go-playground/validator | v10 | Tag-based validation, custom validator support |
| Password Hashing | bcrypt (golang.org/x/crypto) | latest | Industry standard untuk password hashing |
| Configuration | godotenv + os.Getenv | - | Simple, tidak over-engineer untuk MVP |
| Testing | testify + httptest | - | Standard Go testing, assertion helpers |
| UUID | google/uuid | v1 | UUID v4 untuk primary key |

### 1.2 Alasan Memilih Fiber vs Gin vs Echo

```
Fiber  → Performa tertinggi (fasthttp), syntax ekspresif, cocok untuk proyek baru
Gin    → Paling populer, ekosistem luas, net/http compatible
Echo   → Middleware bersih, built-in validator support

Pilihan: Fiber — performa dan DX terbaik untuk scope MVP ini
```

### 1.3 Alasan Memilih GORM vs sqlx vs raw SQL

```
GORM    → Auto migration, hooks, association — cocok untuk domain model yang well-defined
sqlx    → Lebih control, cocok jika query kompleks dominan
raw SQL → Maximum control, verbose

Pilihan: GORM — domain model Pocket App cukup sederhana, migration automation sangat membantu
```

---

## 2. Backend Architecture

Pocket App menggunakan arsitektur **Layered Architecture** (bukan DDD penuh, tapi dengan separation of concern yang jelas). Pattern ini dipilih karena:
- Mudah dipahami reviewer
- Testable per layer
- Cukup untuk scope MVP

### 2.1 Layer Diagram

```
┌─────────────────────────────────────────────────────────┐
│                     HTTP Request                        │
└─────────────────────────┬───────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────┐
│                  Middleware Layer                        │
│  (JWT Auth, Request Logger, Error Recovery, CORS)       │
└─────────────────────────┬───────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────┐
│                   Handler Layer                         │
│  (Parse request, call service, return response)         │
│  Tidak mengandung business logic                        │
└─────────────────────────┬───────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────┐
│                   Service Layer                         │
│  (Business logic, validation lanjutan, orchestration)   │
│  Tidak tahu tentang HTTP                                │
└─────────────────────────┬───────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────┐
│                 Repository Layer                        │
│  (Database access, query building, GORM calls)          │
│  Tidak tahu tentang business logic                      │
└─────────────────────────┬───────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────┐
│                    MySQL Database                       │
└─────────────────────────────────────────────────────────┘
```

### 2.2 Prinsip Per Layer

**Handler Layer**
- Menerima `*fiber.Ctx`
- Parse dan bind request body/params/query
- Memanggil service
- Mengubah hasil service menjadi HTTP response
- **Tidak boleh** mengandung query database atau business logic

**Service Layer**
- Menerima plain Go struct (bukan `*fiber.Ctx`)
- Menjalankan business logic (conditional URL validation, dedup tag, dsb.)
- Memanggil repository
- **Tidak boleh** import `fiber` atau package HTTP apapun

**Repository Layer**
- Menerima parameter plain (string, struct, dsb.)
- Menjalankan query GORM
- Mengembalikan domain model atau error
- **Tidak boleh** mengandung business logic

---

## 3. Struktur Folder

```
pocket-app/
├── cmd/
│   └── server/
│       └── main.go                  # Entry point
│
├── internal/
│   ├── config/
│   │   └── config.go                # Load env, struct Config
│   │
│   ├── database/
│   │   ├── mysql.go                 # Init GORM connection
│   │   └── migration.go             # AutoMigrate + seed
│   │
│   ├── middleware/
│   │   ├── auth.go                  # JWT validation middleware
│   │   ├── error_handler.go         # Global error recovery
│   │   └── logger.go                # Request logger
│   │
│   ├── domain/
│   │   ├── auth/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── dto.go               # Request/Response structs
│   │   │
│   │   ├── pocket/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── dto.go
│   │   │
│   │   └── dashboard/
│   │       ├── handler.go
│   │       ├── service.go
│   │       ├── repository.go
│   │       └── dto.go
│   │
│   ├── model/
│   │   ├── user.go                  # GORM model User
│   │   └── pocket_item.go           # GORM model PocketItem
│   │
│   ├── router/
│   │   └── router.go                # Route registration
│   │
│   └── pkg/
│       ├── response/
│       │   └── response.go          # Helpers: Success(), Error(), Paginated()
│       ├── jwt/
│       │   └── jwt.go               # Generate & parse JWT
│       ├── validator/
│       │   └── validator.go         # Custom validators (URL conditional, dsb.)
│       └── hash/
│           └── hash.go              # bcrypt helpers
│
├── docs/                            # Dokumentasi (dokumen ini)
├── .env.example
├── .env
├── go.mod
├── go.sum
└── README.md
```

---

## 4. Module & Domain Boundary

### 4.1 Auth Module
**Tanggung jawab:**
- Login dengan mock credential
- Generate JWT token
- Logout (client-side token discard; server-side blacklist opsional)

**Depends on:** `model.User`, `pkg/jwt`, `pkg/hash`

### 4.2 Pocket Module
**Tanggung jawab:**
- CRUD pocket item
- Search, filter, sort, pagination
- Toggle favorite
- Update reading status
- Archive (soft delete)

**Depends on:** `model.PocketItem`, middleware auth (untuk `user_id`)

### 4.3 Dashboard Module
**Tanggung jawab:**
- Aggregate count: total, unread, reading, read, archived, favorite
- Recently added items (5 terbaru)

**Depends on:** `model.PocketItem`, middleware auth

---

## 5. Validation Strategy

### 5.1 Layer Validasi

```
Request masuk
    │
    ├── Layer 1: Struct binding validation (go-playground/validator)
    │   └── required, min, max, url format, enum check
    │
    └── Layer 2: Business rule validation (service layer)
        └── conditional URL (jika content_type != note maka URL required)
        └── duplicate tag check
        └── max tag count (10)
        └── tag max length (24 char)
```

### 5.2 Error Response Format untuk Validation

Jika validasi gagal, response mengikuti format PRD:

```json
{
  "code": "VALIDATION_ERROR",
  "message": "Validation error",
  "details": [
    { "field": "title", "message": "Title is required" },
    { "field": "url",   "message": "URL is required for article type" }
  ]
}
```

### 5.3 Custom Validator

Validator kustom yang perlu dibuat:

| Validator | Fungsi |
|---|---|
| `valid_url` | Cek format URL valid (http/https) |
| `content_type_enum` | Cek value dalam `article`, `video`, `document`, `note` |
| `status_enum` | Cek value dalam `unread`, `reading`, `read` |

---

## 6. Error Handling Strategy

### 6.1 Error Types

```go
// internal/pkg/apperror/apperror.go
type AppError struct {
    Code       string
    Message    string
    HTTPStatus int
    Details    []FieldError
}

type FieldError struct {
    Field   string
    Message string
}
```

### 6.2 Error Code Map

| Code | HTTP Status | Kapan Digunakan |
|---|---|---|
| `VALIDATION_ERROR` | 422 | Input tidak valid |
| `INVALID_CREDENTIAL` | 401 | Login gagal |
| `UNAUTHORIZED` | 401 | Token tidak ada / tidak valid |
| `POCKET_NOT_FOUND` | 404 | Item tidak ditemukan atau bukan milik user |
| `INTERNAL_ERROR` | 500 | Unexpected error |

> **Catatan:** Untuk kasus item tidak ditemukan, response yang dikembalikan adalah `POCKET_NOT_FOUND` (404), **bukan** `FORBIDDEN` (403). Ini untuk menghindari exposing bahwa item tersebut exist tapi bukan milik user (security best practice).

### 6.3 Global Error Handler

Fiber error handler terpusat di `middleware/error_handler.go`. Semua `panic` dan `AppError` ditangkap di sini, lalu diformat menjadi response konsisten.

---

## 7. Authentication Strategy

### 7.1 JWT Claims

```go
type JWTClaims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    Name   string `json:"name"`
    jwt.RegisteredClaims
}
```

### 7.2 Token Flow

```
POST /api/auth/login
    │
    ├── Validate request body
    ├── Cari user by email
    ├── Compare password (bcrypt)
    ├── Generate JWT (expiry: 24h dari env)
    └── Return token + user data

Protected endpoint request
    │
    ├── Middleware: extract "Authorization: Bearer <token>"
    ├── Parse & validate JWT
    ├── Set user_id ke fiber.Locals("user_id")
    └── Next handler
```

### 7.3 Mock Credential (Seed Data)

```
Email    : user@example.com
Password : password123
```

---

## 8. Configuration Strategy

### 8.1 Environment Variables

```env
# Server
APP_PORT=3000
APP_ENV=development

# Database
DB_HOST=localhost
DB_PORT=3306
DB_NAME=pocket_app
DB_USER=root
DB_PASSWORD=secret

# JWT
JWT_SECRET=your-super-secret-key-change-in-production
JWT_EXPIRY_HOURS=24

# Seed
SEED_USER_EMAIL=user@example.com
SEED_USER_PASSWORD=password123
SEED_USER_NAME=John Doe
```

### 8.2 Config Struct

```go
type Config struct {
    AppPort        string
    AppEnv         string
    DBHost         string
    DBPort         string
    DBName         string
    DBUser         string
    DBPassword     string
    JWTSecret      string
    JWTExpiryHours int
    SeedUserEmail  string
    SeedUserPass   string
    SeedUserName   string
}
```

---

## 9. Testing Strategy

### 9.1 Scope Testing

| Layer | Type | Tool |
|---|---|---|
| Service | Unit test | `testing` + `testify/assert` |
| Handler | Integration test | `httptest` + `testify` |
| Repository | Integration test (DB) | Test DB / SQLite mock |
| API | Manual / E2E | Bruno / Postman collection |

### 9.2 Test Coverage Target

| Area | Target Coverage |
|---|---|
| Service layer (business logic) | > 80% |
| Handler layer (HTTP binding) | > 70% |
| Repository layer | Happy path + not found |

### 9.3 Test File Convention

```
internal/domain/pocket/
├── handler.go
├── handler_test.go       # Integration test handler
├── service.go
├── service_test.go       # Unit test service
├── repository.go
└── repository_test.go    # DB test (optional)
```

### 9.4 Skenario Test Prioritas

**Auth:**
- Login valid → 200 + token
- Login email kosong → 422
- Login password salah → 401

**Pocket CRUD:**
- Create article valid → 201
- Create article tanpa URL → 422
- Create note tanpa URL → 201 (valid)
- Get list → 200 + pagination
- Get detail milik sendiri → 200
- Get detail milik orang lain / tidak ada → 404
- Update valid → 200
- Archive → 200, item tidak muncul di list
- Search → filter hasil sesuai keyword
- Toggle favorite → state berubah
- Update status → status berubah

---

## 10. Dependency Injection Pattern

Pocket App menggunakan **manual dependency injection** (tanpa framework seperti Wire) untuk kesederhanaan:

```go
// cmd/server/main.go
db := database.Connect(cfg)

// Repository
userRepo    := auth.NewUserRepository(db)
pocketRepo  := pocket.NewPocketRepository(db)

// Service
authSvc     := auth.NewAuthService(userRepo, cfg)
pocketSvc   := pocket.NewPocketService(pocketRepo)
dashSvc     := dashboard.NewDashboardService(pocketRepo)

// Handler
authHandler    := auth.NewAuthHandler(authSvc)
pocketHandler  := pocket.NewPocketHandler(pocketSvc)
dashHandler    := dashboard.NewDashboardHandler(dashSvc)

// Router
router.Setup(app, authHandler, pocketHandler, dashHandler)
```

---

## 11. Response Helper

Semua handler menggunakan helper terpusat untuk konsistensi:

```go
// internal/pkg/response/response.go

func Success(c *fiber.Ctx, status int, data interface{}, message string) error
func Paginated(c *fiber.Ctx, data interface{}, meta PaginationMeta) error
func Error(c *fiber.Ctx, status int, code string, message string) error
func ValidationError(c *fiber.Ctx, details []FieldError) error
```

---

## 12. Development Workflow

```
1. Buat/update model (internal/model/)
2. Jalankan migration (AutoMigrate)
3. Buat repository interface + implementation
4. Buat service dengan business logic
5. Buat handler (bind → call service → response)
6. Register route di router
7. Tulis unit test service
8. Tulis integration test handler
9. Test manual via Bruno/Postman
```

---

*Dokumen ini menjadi acuan untuk seluruh implementasi. Perubahan arsitektur yang signifikan harus diupdate di sini sebelum diimplementasikan.*
