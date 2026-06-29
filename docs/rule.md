# CLAUDE.md — Pocket App Backend

Rules ini mendefinisikan standar pengerjaan, konvensi kode, dan keputusan teknis yang harus diikuti secara konsisten selama pengembangan backend Pocket App. Dokumen ini adalah acuan utama dan tidak boleh diabaikan.

---

## Project Context

- **Project:** Pocket App — Personal pocket item management backend
- **Stack:** Go 1.22+, Fiber v2, GORM v2, MySQL 8.0, JWT (golang-jwt v5)
- **Architecture:** Layered Architecture (Handler → Service → Repository)
- **Docs:** Seluruh keputusan teknis ada di `docs/` — baca sebelum membuat keputusan baru
- **PRD:** `docs/01-prd-analysis.md` adalah source of truth requirement
- **API Contract:** `docs/04-payload-contract.md` adalah source of truth response format

---

## 1. Prinsip Utama

```
1. Readability over cleverness       — kode harus mudah dibaca reviewer
2. Explicit over implicit            — jangan sembunyikan logic di magic/reflection
3. Errors are values                 — selalu handle error, jangan ignore
4. Layered responsibility            — setiap layer punya satu tanggung jawab
5. Contract-first                    — implementasi mengikuti payload contract, bukan sebaliknya
```

---

## 2. Struktur Folder

Ikuti struktur ini **secara ketat**. Jangan buat folder baru tanpa alasan yang jelas.

```
pocket-app/
├── cmd/
│   └── server/
│       └── main.go                  # Entry point, dependency wiring
│
├── internal/
│   ├── config/
│   │   └── config.go                # Load env, Config struct
│   │
│   ├── database/
│   │   ├── mysql.go                 # Init GORM connection
│   │   ├── migration.go             # AutoMigrate
│   │   └── seed.go                  # Seed mock user + items
│   │
│   ├── middleware/
│   │   ├── auth.go                  # JWT extraction & validation
│   │   ├── error_handler.go         # Global Fiber error handler
│   │   └── logger.go                # Request logger
│   │
│   ├── domain/
│   │   ├── auth/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── dto.go
│   │   ├── pocket/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── dto.go
│   │   └── dashboard/
│   │       ├── handler.go
│   │       ├── service.go
│   │       ├── repository.go
│   │       └── dto.go
│   │
│   ├── model/
│   │   ├── user.go                  # GORM model
│   │   └── pocket_item.go           # GORM model + StringArray type
│   │
│   ├── router/
│   │   └── router.go                # Route registration
│   │
│   └── pkg/
│       ├── apperror/
│       │   └── apperror.go          # AppError, FieldError, error constructors
│       ├── response/
│       │   └── response.go          # Success(), Error(), Paginated() helpers
│       ├── jwt/
│       │   └── jwt.go               # GenerateToken(), ParseToken()
│       ├── validator/
│       │   └── validator.go         # Custom validators + FormatErrors()
│       └── hash/
│           └── hash.go              # HashPassword(), ComparePassword()
│
├── docs/
├── .env.example
├── .env                             # Jangan di-commit
├── go.mod
├── go.sum
└── README.md
```

**Rules:**
- Semua package di bawah `internal/` — tidak ada yang boleh diimport dari luar module ini
- `cmd/server/main.go` hanya berisi wiring, tidak ada logic
- Setiap domain folder berisi tepat 4 file: `handler.go`, `service.go`, `repository.go`, `dto.go`

---

## 3. Layer Rules

### 3.1 Handler Layer (`handler.go`)

**Boleh:**
- Import `github.com/gofiber/fiber/v2`
- Parse `c.Body()`, `c.Params()`, `c.Query()`
- Bind request ke DTO struct
- Call service
- Call `response.Success()` / `response.Error()`

**Dilarang:**
- Mengandung business logic apapun
- Import `gorm.io/gorm` atau package database
- Menulis query langsung
- Mengakses `c.Locals("user_id")` lebih dari sekali — ambil sekali, simpan ke variabel

```go
// BENAR
func (h *PocketHandler) Create(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(string)

    var req dto.CreatePocketRequest
    if err := c.BodyParser(&req); err != nil {
        return apperror.BadRequest("Invalid request body")
    }

    if errs := validator.Validate(req); errs != nil {
        return apperror.ValidationError(errs)
    }

    result, err := h.service.Create(c.Context(), userID, req)
    if err != nil {
        return err
    }

    return response.Created(c, result, "Pocket item created successfully")
}

// SALAH — business logic di handler
func (h *PocketHandler) Create(c *fiber.Ctx) error {
    // ❌ jangan lakukan ini
    if req.ContentType != "note" && req.URL == "" {
        return c.Status(422).JSON(...)
    }
}
```

### 3.2 Service Layer (`service.go`)

**Boleh:**
- Mengandung seluruh business logic
- Memanggil repository
- Melakukan validasi kondisional (conditional URL, dedup tag, dsb.)
- Return `AppError` untuk kasus error bisnis

**Dilarang:**
- Import `github.com/gofiber/fiber/v2`
- Mengakses HTTP context langsung
- Menulis query GORM langsung

```go
// BENAR — conditional validation ada di service
func (s *PocketService) Create(ctx context.Context, userID string, req dto.CreatePocketRequest) (*dto.PocketResponse, error) {
    // Business rule BR-004: URL wajib kecuali note
    if req.ContentType != "note" && (req.URL == nil || *req.URL == "") {
        return nil, apperror.ValidationError([]apperror.FieldError{
            {Field: "url", Message: "URL is required"},
        })
    }

    // Business rule BR-009: dedup tags
    req.Tags = deduplicateTags(req.Tags)

    item := &model.PocketItem{
        ID:          uuid.New().String(),
        UserID:      userID,
        Title:       req.Title,
        ContentType: req.ContentType,
        Status:      "unread",    // BR-010: default unread
        IsFavorite:  false,       // BR-011: default false
        Tags:        req.Tags,
    }

    if err := s.repo.Create(ctx, item); err != nil {
        return nil, err
    }

    return toResponse(item), nil
}
```

### 3.3 Repository Layer (`repository.go`)

**Boleh:**
- Import `gorm.io/gorm`
- Menulis query GORM
- Return `*model.X` atau `error`

**Dilarang:**
- Mengandung business logic
- Import `fiber` atau package domain lain
- Return DTO langsung — selalu return model

```go
// BENAR
func (r *PocketRepository) FindByIDAndUserID(ctx context.Context, id, userID string) (*model.PocketItem, error) {
    var item model.PocketItem
    err := r.db.WithContext(ctx).
        Where("id = ? AND user_id = ?", id, userID).
        First(&item).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, apperror.NotFound("Pocket item not found")
    }
    return &item, err
}

// SALAH — business logic di repository
func (r *PocketRepository) FindActive(...) {
    // ❌ jangan hardcode business rule di sini tanpa parameter
    r.db.Where("status != 'archived'")
}
```

---

## 4. Error Handling

### 4.1 AppError

Selalu gunakan `AppError` untuk error yang perlu disampaikan ke client. Jangan return `errors.New("...")` langsung dari service.

```go
// internal/pkg/apperror/apperror.go

type AppError struct {
    Code       string
    Message    string
    HTTPStatus int
    Details    []FieldError
}

type FieldError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

func (e *AppError) Error() string { return e.Message }

// Constructors
func NotFound(msg string) *AppError
func Unauthorized(msg string) *AppError
func ValidationError(details []FieldError) *AppError
func Internal(msg string) *AppError
func InvalidCredential() *AppError
```

### 4.2 Error Propagation

```go
// Propagate error dari repository ke service ke handler
result, err := s.repo.FindByIDAndUserID(ctx, id, userID)
if err != nil {
    return nil, err   // ✅ propagate as-is jika sudah AppError
}

// Jangan wrap AppError dengan error baru
return nil, fmt.Errorf("failed to find: %w", err)  // ❌
```

### 4.3 Global Error Handler

Semua error di-handle di `middleware/error_handler.go`. Handler ini menangkap `*AppError` dan format lain, lalu mengubahnya menjadi response JSON konsisten.

```go
// Jangan lakukan manual error response di handler
return c.Status(404).JSON(fiber.Map{"error": "not found"})  // ❌

// Cukup return error, biarkan global handler yang format
return apperror.NotFound("Pocket item not found")  // ✅
```

---

## 5. Response Format

**Wajib** menggunakan helper dari `internal/pkg/response/`. Jangan pernah menulis `c.Status().JSON()` manual di handler.

```go
// internal/pkg/response/response.go

// Single object — 200
response.Success(c, data, "message")

// Created — 201
response.Created(c, data, "message")

// Paginated list — 200
response.Paginated(c, data, meta)

// Error — status dari AppError
// (di-handle global, tidak dipanggil manual di handler)
```

**Format response mengikuti `docs/04-payload-contract.md` secara ketat.** Field name menggunakan `camelCase` di JSON.

---

## 6. DTO Convention

Semua DTO ada di `dto.go` masing-masing domain.

```go
// Naming convention
type CreatePocketRequest struct { ... }   // Request body
type UpdatePocketRequest struct { ... }   // PUT request
type PocketResponse struct { ... }        // Response single item
type PocketListQuery struct { ... }       // Query params untuk list

// JSON tags selalu camelCase
type PocketResponse struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    ContentType string    `json:"contentType"`
    IsFavorite  bool      `json:"isFavorite"`
    CreatedAt   time.Time `json:"createdAt"`
}

// Validator tags pada request DTO
type CreatePocketRequest struct {
    Title       string   `json:"title" validate:"required,min=3,max=120"`
    URL         *string  `json:"url"`
    Description *string  `json:"description" validate:"omitempty,max=500"`
    ContentType string   `json:"contentType" validate:"required,oneof=article video document note"`
    Tags        []string `json:"tags" validate:"omitempty,max=10,dive,max=24"`
}
```

**Rules DTO:**
- Request DTO hanya boleh punya field yang relevan dengan input user
- Response DTO tidak boleh expose field sensitif (`password`, internal ID dari DB, dsb.)
- Field nullable menggunakan pointer (`*string`, `*bool`)
- Jangan return `model.X` langsung dari service — selalu convert ke DTO

---

## 7. Model Convention

```go
// GORM model ada di internal/model/
// Selalu definisikan TableName()
func (PocketItem) TableName() string { return "pocket_items" }

// UUID sebagai primary key — set di BeforeCreate hook atau di service
// Jangan pakai gorm.Model (menggunakan uint ID) — kita pakai UUID string

// Pointer untuk nullable field
URL         *string `gorm:"type:varchar(2048)"`
Description *string `gorm:"type:text"`
```

---

## 8. Database & Query Rules

### 8.1 User Isolation

**Setiap query ke `pocket_items` wajib menyertakan `user_id`.**

```go
// ✅ BENAR — selalu filter by user_id
r.db.Where("id = ? AND user_id = ?", id, userID).First(&item)

// ❌ SALAH — tidak aman, bisa akses data user lain
r.db.Where("id = ?", id).First(&item)
```

### 8.2 Soft Delete

Archive = set `status = 'archived'`. Jangan gunakan GORM soft delete (`DeletedAt`).

```go
// Archive
r.db.Model(&model.PocketItem{}).
    Where("id = ? AND user_id = ?", id, userID).
    Updates(map[string]interface{}{
        "status":     "archived",
        "updated_at": time.Now(),
    })
```

### 8.3 List Query

List aktif selalu filter `status != 'archived'` secara default, **kecuali** parameter `status=archived` dikirim.

```go
query := r.db.Where("user_id = ?", userID)

if filter.Status == "archived" {
    query = query.Where("status = ?", "archived")
} else {
    query = query.Where("status != ?", "archived")
    if filter.Status != "" {
        query = query.Where("status = ?", filter.Status)
    }
}
```

### 8.4 Context

Selalu pass `context.Context` ke semua query GORM menggunakan `.WithContext(ctx)`.

```go
r.db.WithContext(ctx).Where(...).Find(&items)  // ✅
r.db.Where(...).Find(&items)                   // ❌
```

---

## 9. Validation Rules

Dua layer validasi:

**Layer 1 — Struct tag (go-playground/validator):**
```go
validate:"required,min=3,max=120"
```

**Layer 2 — Business rule (service layer):**
- Conditional URL validation (content_type != note → URL required)
- Duplicate tag check
- Tags count max 10
- Status tidak boleh `archived` di PATCH /status

Jangan pindahkan Layer 2 ke handler atau repository.

**Format error validation mengikuti contract:**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Validation error",
  "details": [{ "field": "title", "message": "Title is required" }]
}
```

Field name pada `details` menggunakan nama JSON field (camelCase), bukan nama Go struct field.

---

## 10. Authentication & JWT

```go
// JWT claims wajib berisi minimal:
type JWTClaims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    Name   string `json:"name"`
    jwt.RegisteredClaims
}

// Ambil user_id di handler via Locals — set oleh middleware auth
userID := c.Locals("user_id").(string)

// Middleware auth wajib ada di semua route kecuali POST /api/auth/login
```

**Rules:**
- Token disimpan di `Authorization: Bearer <token>` header
- Middleware auth return `UNAUTHORIZED` (401) jika token tidak ada atau tidak valid
- Expiry JWT dikonfigurasi via env `JWT_EXPIRY_HOURS`

---

## 11. Configuration

Semua konfigurasi dibaca dari environment variable melalui `internal/config/config.go`. Jangan hardcode value apapun di luar config.

```go
// ❌ SALAH
db, _ := gorm.Open("root:secret@tcp(localhost:3306)/pocket_app")

// ✅ BENAR
dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?...", cfg.DBUser, cfg.DBPassword, ...)
```

File `.env` tidak boleh di-commit ke repository. `.env.example` wajib selalu up-to-date.

---

## 12. Naming Conventions

| Context | Convention | Contoh |
|---|---|---|
| Package | lowercase, singkat | `pocket`, `auth`, `apperror` |
| Interface | nama + `er` atau deskriptif | `PocketRepository`, `AuthService` |
| Struct | PascalCase | `PocketItem`, `CreatePocketRequest` |
| Function/Method | PascalCase (exported) | `FindByID`, `Create`, `Archive` |
| Variable | camelCase | `userID`, `pocketItem` |
| Constant | PascalCase atau SCREAMING_SNAKE | `StatusUnread`, `ErrNotFound` |
| JSON field | camelCase | `isFavorite`, `contentType`, `createdAt` |
| DB column | snake_case | `user_id`, `is_favorite`, `created_at` |
| File | snake_case | `pocket_item.go`, `error_handler.go` |

---

## 13. Testing Rules

```
Unit test    → internal/domain/<domain>/service_test.go
Integration  → internal/domain/<domain>/handler_test.go
```

**Wajib test:**
- Semua business logic di service layer
- Happy path + validation error + not found untuk setiap endpoint
- User isolation: pastikan user A tidak bisa akses item user B

**Minimal test yang harus ada sebelum PR:**
- `TestLogin_Success`
- `TestLogin_InvalidCredential`
- `TestCreatePocket_ArticleWithURL_Success`
- `TestCreatePocket_ArticleWithoutURL_ValidationError`
- `TestCreatePocket_NoteWithoutURL_Success`
- `TestGetPocketDetail_NotOwner_NotFound`
- `TestArchivePocket_HidesFromDefaultList`

**Mock repository** menggunakan interface — jangan test langsung ke database di unit test.

---

## 14. Git & Commit Convention

```
feat: add create pocket item endpoint
fix: handle nil pointer in pocket response mapper
refactor: extract tag deduplication to helper function
test: add unit tests for pocket service
docs: update payload contract for status endpoint
chore: add .env.example
```

Format: `<type>: <description singkat dalam bahasa inggris>`

---

## 15. Hal yang Dilarang (Hard Rules)

```
❌ Jangan ignore error dengan _
❌ Jangan gunakan panic() kecuali di main.go saat startup
❌ Jangan return model.X dari service — selalu return DTO
❌ Jangan query database tanpa filter user_id di pocket endpoints
❌ Jangan hardcode credential, secret, atau DSN di luar config
❌ Jangan commit file .env
❌ Jangan buat endpoint baru tanpa mendefinisikannya di payload contract dulu
❌ Jangan gunakan status 'archived' sebagai nilai valid di PATCH /status
❌ Jangan gunakan fmt.Println untuk logging di production code
❌ Jangan import package domain lain secara silang (misal pocket import auth domain)
```

---

## 16. Checklist Sebelum Implementasi Endpoint Baru

```
[ ] Endpoint sudah ada di docs/04-payload-contract.md
[ ] Request DTO sudah dibuat dengan validation tag
[ ] Response DTO sudah dibuat dengan JSON tag camelCase
[ ] Service method sudah dibuat dengan business logic yang tepat
[ ] Repository method sudah dibuat dengan user_id filter
[ ] Route sudah didaftarkan di router.go dengan middleware auth
[ ] Error response menggunakan AppError constructor
[ ] Response menggunakan helper dari pkg/response
[ ] Unit test service sudah ditulis
[ ] Manual test via Bruno/Postman sudah dijalankan
[ ] README endpoint list sudah diupdate
```

---

## 17. Referensi Dokumen

| Dokumen | Isi |
|---|---|
| `docs/01-prd-analysis.md` | Domain, fitur, business rules, asumsi |
| `docs/02-technical-plan.md` | Stack, arsitektur, folder structure, testing strategy |
| `docs/03-database-design.md` | Skema MySQL, GORM model, migration, seed |
| `docs/04-payload-contract.md` | Semua endpoint, request/response, validation rules |
| `docs/05-mock-api.md` | 31 skenario mock, contoh data |

**Jika ada konflik antara CLAUDE.md dan dokumen di `docs/`, selalu refer ke `docs/04-payload-contract.md` untuk hal yang berkaitan dengan API contract.**
