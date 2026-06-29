# 03 - Database Design

## Overview

Dokumen ini mendefinisikan rancangan database MySQL untuk Pocket App. Mencakup entity, skema tabel, relasi, index, soft delete strategy, dan migration/seed plan.

---

## 1. Entity Utama

```
┌──────────┐         ┌─────────────────┐
│  users   │ 1 ───── N │  pocket_items  │
└──────────┘         └─────────────────┘
```

Hanya 2 tabel utama. Tags disimpan sebagai **JSON column** di dalam `pocket_items` (keputusan berdasarkan analisis PRD — tags bersifat per-item, tidak shared, cukup untuk MVP dengan LIKE search).

---

## 2. Tabel: `users`

### 2.1 Definisi

```sql
CREATE TABLE users (
    id          VARCHAR(36)  NOT NULL,
    name        VARCHAR(100) NOT NULL,
    email       VARCHAR(255) NOT NULL,
    password    VARCHAR(255) NOT NULL,
    avatar_url  VARCHAR(500) NULL,
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    PRIMARY KEY (id),
    UNIQUE KEY uq_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 2.2 Field Definition

| Field | Type | Nullable | Default | Keterangan |
|---|---|---|---|---|
| `id` | VARCHAR(36) | NO | - | UUID v4 |
| `name` | VARCHAR(100) | NO | - | Display name |
| `email` | VARCHAR(255) | NO | - | Unique, dipakai login |
| `password` | VARCHAR(255) | NO | - | bcrypt hash |
| `avatar_url` | VARCHAR(500) | YES | NULL | URL avatar opsional |
| `created_at` | DATETIME(3) | NO | NOW | Auto set |
| `updated_at` | DATETIME(3) | NO | NOW | Auto update |

### 2.3 Index

| Index | Column | Type | Tujuan |
|---|---|---|---|
| PRIMARY | `id` | PRIMARY KEY | Lookup by ID |
| `uq_users_email` | `email` | UNIQUE | Prevent duplicate email, fast login lookup |

---

## 3. Tabel: `pocket_items`

### 3.1 Definisi

```sql
CREATE TABLE pocket_items (
    id           VARCHAR(36)   NOT NULL,
    user_id      VARCHAR(36)   NOT NULL,
    title        VARCHAR(120)  NOT NULL,
    url          VARCHAR(2048) NULL,
    description  TEXT          NULL,
    content_type ENUM('article','video','document','note') NOT NULL DEFAULT 'article',
    status       ENUM('unread','reading','read','archived') NOT NULL DEFAULT 'unread',
    is_favorite  TINYINT(1)    NOT NULL DEFAULT 0,
    tags         JSON          NULL,
    created_at   DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at   DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    PRIMARY KEY (id),
    CONSTRAINT fk_pocket_items_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,

    INDEX idx_pocket_items_user_id      (user_id),
    INDEX idx_pocket_items_status       (user_id, status),
    INDEX idx_pocket_items_content_type (user_id, content_type),
    INDEX idx_pocket_items_is_favorite  (user_id, is_favorite),
    INDEX idx_pocket_items_created_at   (user_id, created_at DESC),
    FULLTEXT INDEX ft_pocket_items_search (title, url, description)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 3.2 Field Definition

| Field | Type | Nullable | Default | Keterangan |
|---|---|---|---|---|
| `id` | VARCHAR(36) | NO | - | UUID v4 |
| `user_id` | VARCHAR(36) | NO | - | FK → `users.id` |
| `title` | VARCHAR(120) | NO | - | Min 3, max 120 char |
| `url` | VARCHAR(2048) | YES | NULL | Wajib kecuali type=note |
| `description` | TEXT | YES | NULL | Max 500 char (validated at app level) |
| `content_type` | ENUM | NO | `article` | article/video/document/note |
| `status` | ENUM | NO | `unread` | unread/reading/read/archived |
| `is_favorite` | TINYINT(1) | NO | `0` | Boolean: 0=false, 1=true |
| `tags` | JSON | YES | NULL | Array of strings, e.g. `["react","frontend"]` |
| `created_at` | DATETIME(3) | NO | NOW | Auto set |
| `updated_at` | DATETIME(3) | NO | NOW | Auto update |

### 3.3 Index Strategy

| Index | Column(s) | Type | Tujuan |
|---|---|---|---|
| PRIMARY | `id` | PRIMARY KEY | Lookup by ID |
| `fk_pocket_items_user` | `user_id` | FK | Join ke users |
| `idx_pocket_items_user_id` | `user_id` | INDEX | Filter by owner |
| `idx_pocket_items_status` | `user_id, status` | COMPOSITE | Filter status per user |
| `idx_pocket_items_content_type` | `user_id, content_type` | COMPOSITE | Filter type per user |
| `idx_pocket_items_is_favorite` | `user_id, is_favorite` | COMPOSITE | Filter favorite per user |
| `idx_pocket_items_created_at` | `user_id, created_at` | COMPOSITE | Sort default per user |
| `ft_pocket_items_search` | `title, url, description` | FULLTEXT | Full-text search |

> **Catatan FULLTEXT:** FULLTEXT index tidak meng-cover kolom `tags` (JSON). Untuk search terhadap tags, digunakan `JSON_SEARCH(tags, 'one', ?)` atau `JSON_CONTAINS(tags, JSON_QUOTE(?))`. Untuk MVP, approach ini cukup. Jika performa menjadi concern, tags dapat dipindah ke tabel terpisah.

### 3.4 Soft Delete / Archive Strategy

Pocket App **tidak menggunakan hard delete**. Archive dilakukan dengan mengubah nilai `status` menjadi `'archived'`.

```
DELETE /api/pockets/:id
    │
    └── UPDATE pocket_items SET status = 'archived', updated_at = NOW()
        WHERE id = ? AND user_id = ?
```

**Query list utama** selalu menyertakan filter `status != 'archived'` secara default:
```sql
-- List aktif (default)
SELECT * FROM pocket_items WHERE user_id = ? AND status != 'archived'

-- List archive
SELECT * FROM pocket_items WHERE user_id = ? AND status = 'archived'
```

Ini lebih sederhana dari `deleted_at` nullable (soft delete pattern Gorm) karena `archived` sudah memiliki makna bisnis yang jelas di PRD.

---

## 4. GORM Model

### 4.1 User Model

```go
// internal/model/user.go
package model

import "time"

type User struct {
    ID        string    `gorm:"type:varchar(36);primaryKey"`
    Name      string    `gorm:"type:varchar(100);not null"`
    Email     string    `gorm:"type:varchar(255);not null;uniqueIndex"`
    Password  string    `gorm:"type:varchar(255);not null"`
    AvatarURL *string   `gorm:"type:varchar(500)"`
    CreatedAt time.Time `gorm:"precision:3;autoCreateTime"`
    UpdatedAt time.Time `gorm:"precision:3;autoUpdateTime"`
}

func (User) TableName() string {
    return "users"
}
```

### 4.2 PocketItem Model

```go
// internal/model/pocket_item.go
package model

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
    "time"
)

// StringArray adalah custom type untuk JSON column tags
type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
    if s == nil {
        return "[]", nil
    }
    b, err := json.Marshal(s)
    return string(b), err
}

func (s *StringArray) Scan(value interface{}) error {
    if value == nil {
        *s = StringArray{}
        return nil
    }
    bytes, ok := value.([]byte)
    if !ok {
        return fmt.Errorf("cannot scan type %T into StringArray", value)
    }
    return json.Unmarshal(bytes, s)
}

type PocketItem struct {
    ID          string      `gorm:"type:varchar(36);primaryKey"`
    UserID      string      `gorm:"type:varchar(36);not null;index"`
    Title       string      `gorm:"type:varchar(120);not null"`
    URL         *string     `gorm:"type:varchar(2048)"`
    Description *string     `gorm:"type:text"`
    ContentType string      `gorm:"type:enum('article','video','document','note');not null;default:'article'"`
    Status      string      `gorm:"type:enum('unread','reading','read','archived');not null;default:'unread'"`
    IsFavorite  bool        `gorm:"type:tinyint(1);not null;default:0"`
    Tags        StringArray `gorm:"type:json"`
    CreatedAt   time.Time   `gorm:"precision:3;autoCreateTime"`
    UpdatedAt   time.Time   `gorm:"precision:3;autoUpdateTime"`

    User User `gorm:"foreignKey:UserID;references:ID"`
}

func (PocketItem) TableName() string {
    return "pocket_items"
}
```

---

## 5. Migration Plan

### 5.1 AutoMigrate (GORM)

GORM AutoMigrate digunakan untuk development dan initial setup. AutoMigrate akan:
- Membuat tabel jika belum ada
- Menambah kolom baru jika belum ada
- **Tidak akan** menghapus kolom atau mengubah tipe kolom yang sudah ada

```go
// internal/database/migration.go
func Migrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &model.User{},
        &model.PocketItem{},
    )
}
```

### 5.2 Index yang Perlu Dibuat Manual

GORM AutoMigrate tidak selalu membuat FULLTEXT index dan composite index dengan tepat. Index berikut perlu dibuat manual setelah AutoMigrate:

```sql
-- Jalankan setelah AutoMigrate berhasil

-- Composite indexes
CREATE INDEX IF NOT EXISTS idx_pocket_items_status
    ON pocket_items (user_id, status);

CREATE INDEX IF NOT EXISTS idx_pocket_items_content_type
    ON pocket_items (user_id, content_type);

CREATE INDEX IF NOT EXISTS idx_pocket_items_is_favorite
    ON pocket_items (user_id, is_favorite);

CREATE INDEX IF NOT EXISTS idx_pocket_items_created_at
    ON pocket_items (user_id, created_at DESC);

-- FULLTEXT index untuk search
ALTER TABLE pocket_items
    ADD FULLTEXT INDEX ft_pocket_items_search (title, url, description);
```

### 5.3 Seed Data

```go
// internal/database/seed.go
func Seed(db *gorm.DB, cfg *config.Config) error {
    // Seed mock user
    var count int64
    db.Model(&model.User{}).Where("email = ?", cfg.SeedUserEmail).Count(&count)
    if count > 0 {
        return nil // already seeded
    }

    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(cfg.SeedUserPass), bcrypt.DefaultCost)
    user := model.User{
        ID:       uuid.New().String(),
        Name:     cfg.SeedUserName,
        Email:    cfg.SeedUserEmail,
        Password: string(hashedPassword),
    }
    if err := db.Create(&user).Error; err != nil {
        return err
    }

    // Seed sample pocket items
    url1 := "https://example.com/react-performance"
    desc1 := "A guide about React rendering optimization"
    url2 := "https://example.com/typescript-generics"
    desc2 := "Deep dive into TypeScript generics"
    desc3 := "Personal notes about scalable frontend architecture"

    items := []model.PocketItem{
        {
            ID: uuid.New().String(), UserID: user.ID,
            Title: "React Performance Guide", URL: &url1, Description: &desc1,
            ContentType: "article", Status: "unread", IsFavorite: true,
            Tags: model.StringArray{"frontend", "react"},
        },
        {
            ID: uuid.New().String(), UserID: user.ID,
            Title: "Understanding TypeScript Generics", URL: &url2, Description: &desc2,
            ContentType: "article", Status: "reading", IsFavorite: false,
            Tags: model.StringArray{"typescript", "frontend"},
        },
        {
            ID: uuid.New().String(), UserID: user.ID,
            Title: "Frontend System Design Notes", Description: &desc3,
            ContentType: "note", Status: "read", IsFavorite: true,
            Tags: model.StringArray{"architecture", "frontend"},
        },
    }

    return db.Create(&items).Error
}
```

---

## 6. Query Patterns

### 6.1 List dengan Dynamic Filter

```sql
SELECT *
FROM pocket_items
WHERE user_id = ?
  AND status != 'archived'                          -- selalu, kecuali archive page
  AND (? = '' OR status = ?)                        -- filter status (opsional)
  AND (? = '' OR content_type = ?)                  -- filter type (opsional)
  AND (? = FALSE OR is_favorite = TRUE)             -- filter favorite (opsional)
  AND (
    ? = ''                                          -- jika search kosong, skip
    OR MATCH(title, url, description) AGAINST(? IN BOOLEAN MODE)
    OR JSON_SEARCH(tags, 'one', ?) IS NOT NULL      -- search dalam tags
  )
ORDER BY created_at DESC                            -- default sort
LIMIT ? OFFSET ?
```

### 6.2 Dashboard Summary

```sql
SELECT
    COUNT(*) AS total_items,
    SUM(status = 'unread') AS unread_items,
    SUM(status = 'reading') AS reading_items,
    SUM(status = 'read') AS read_items,
    SUM(status = 'archived') AS archived_items,
    SUM(is_favorite = 1) AS favorite_items
FROM pocket_items
WHERE user_id = ?;
```

### 6.3 Get Detail (dengan ownership check)

```sql
SELECT * FROM pocket_items
WHERE id = ? AND user_id = ?
LIMIT 1;
```

Jika tidak ditemukan (baik karena tidak ada maupun bukan miliknya) → return `POCKET_NOT_FOUND`.

---

## 7. Pertimbangan & Tradeoff

| Keputusan | Alternatif | Alasan Dipilih |
|---|---|---|
| Tags sebagai JSON column | Tabel `pocket_item_tags` terpisah | Lebih simpel untuk MVP, tidak perlu JOIN, cukup untuk data kecil |
| Soft delete via `status='archived'` | `deleted_at DATETIME NULL` (GORM soft delete) | Lebih ekspresif secara bisnis, status archived sudah ada di domain |
| UUID VARCHAR(36) | BIGINT AUTO_INCREMENT | Lebih aman untuk distributed setup masa depan, tidak expose row count |
| FULLTEXT index | LIKE '%keyword%' | Performa lebih baik untuk search, terutama bila data bertambah |
| DATETIME(3) | TIMESTAMP | DATETIME tidak terpengaruh timezone MySQL, lebih predictable |

---

*Dokumen ini menjadi acuan untuk implementasi repository layer dan migration script.*
