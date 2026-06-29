# 04 - Payload Contract

## Overview

Dokumen ini mendefinisikan contract lengkap untuk semua API endpoint Pocket App. Contract ini menjadi acuan tunggal bagi implementasi backend dan pengujian API. Setiap endpoint didefinisikan dengan request, response sukses, response error, validation rule, dan status code.

---

## Konvensi

- Base URL: `http://localhost:3000`
- Semua request/response menggunakan `Content-Type: application/json`
- Semua protected endpoint membutuhkan header: `Authorization: Bearer <token>`
- Timestamp menggunakan format ISO 8601: `2026-06-26T10:00:00Z`
- Semua ID menggunakan UUID v4

---

## Response Wrapper

### Success (Single Object)
```json
{
  "data": { },
  "message": "string"
}
```

### Success (Paginated List)
```json
{
  "data": [],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 25,
    "totalPage": 3
  }
}
```

### Error
```json
{
  "code": "ERROR_CODE",
  "message": "Human readable message",
  "details": [
    { "field": "field_name", "message": "Error detail" }
  ]
}
```
> `details` hanya muncul pada `VALIDATION_ERROR`.

---

## Error Code Reference

| Code | HTTP Status | Deskripsi |
|---|---|---|
| `VALIDATION_ERROR` | 422 | Input request tidak valid |
| `INVALID_CREDENTIAL` | 401 | Email atau password salah |
| `UNAUTHORIZED` | 401 | Token tidak ada atau tidak valid |
| `POCKET_NOT_FOUND` | 404 | Pocket item tidak ditemukan |
| `INTERNAL_ERROR` | 500 | Server error tidak terduga |

---

## Enum Reference

### Content Type
| Value | Label |
|---|---|
| `article` | Article |
| `video` | Video |
| `document` | Document |
| `note` | Note |

### Status
| Value | Label |
|---|---|
| `unread` | Unread |
| `reading` | Reading |
| `read` | Read |
| `archived` | Archived |

### Sort Option
| Value | Behavior |
|---|---|
| `createdAt:desc` | Newest (default) |
| `createdAt:asc` | Oldest |
| `title:asc` | Title A-Z |
| `title:desc` | Title Z-A |
| `updatedAt:desc` | Recently Updated |

---

---

# AUTH ENDPOINTS

---

## POST /api/auth/login

Login dengan mock credential dan mendapatkan JWT token.

**Auth Required:** Tidak

### Request Body

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

| Field | Type | Required | Rule |
|---|---|---|---|
| `email` | string | Ya | Format email valid |
| `password` | string | Ya | Minimum 6 karakter |

### Response 200 — Login Berhasil

```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "John Doe",
      "email": "user@example.com",
      "avatarUrl": null
    }
  },
  "message": "Login success"
}
```

### Response 422 — Validation Error

```json
{
  "code": "VALIDATION_ERROR",
  "message": "Validation error",
  "details": [
    { "field": "email", "message": "Email is required" }
  ]
}
```

### Response 401 — Credential Salah

```json
{
  "code": "INVALID_CREDENTIAL",
  "message": "Email or password is incorrect"
}
```

### Validation Rules

| Field | Rule | Error Message |
|---|---|---|
| `email` | required | `Email is required` |
| `email` | valid email format | `Email format is invalid` |
| `password` | required | `Password is required` |
| `password` | min 6 characters | `Password must be at least 6 characters` |

---

## POST /api/auth/logout

Logout user. Token invalidation dilakukan di sisi client (stateless JWT).

**Auth Required:** Ya

### Request Body
Tidak ada.

### Response 200 — Logout Berhasil

```json
{
  "data": null,
  "message": "Logout success"
}
```

### Response 401 — Tidak Terautentikasi

```json
{
  "code": "UNAUTHORIZED",
  "message": "Unauthorized"
}
```

---

---

# DASHBOARD ENDPOINTS

---

## GET /api/dashboard

Mendapatkan ringkasan statistik pocket item milik user yang sedang login.

**Auth Required:** Ya

### Request
Tidak ada query parameter.

### Response 200 — Berhasil

```json
{
  "data": {
    "totalItems": 15,
    "unreadItems": 7,
    "readingItems": 3,
    "readItems": 4,
    "archivedItems": 1,
    "favoriteItems": 5,
    "recentlyAdded": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440001",
        "title": "React Performance Guide",
        "url": "https://example.com/react-performance",
        "contentType": "article",
        "status": "unread",
        "isFavorite": true,
        "tags": ["frontend", "react"],
        "createdAt": "2026-06-26T10:00:00Z"
      }
    ]
  },
  "message": "Dashboard loaded"
}
```

> `recentlyAdded` berisi maksimal 5 item terbaru dengan status bukan `archived`.

### Response 401 — Tidak Terautentikasi

```json
{
  "code": "UNAUTHORIZED",
  "message": "Unauthorized"
}
```

---

---

# POCKET ITEM ENDPOINTS

---

## GET /api/pockets

Mendapatkan daftar pocket item aktif (status != archived) milik user dengan support search, filter, sort, dan pagination.

**Auth Required:** Ya

### Query Parameters

| Param | Type | Required | Default | Deskripsi |
|---|---|---|---|---|
| `search` | string | Tidak | `""` | Keyword untuk search di title, url, description, tags |
| `status` | string | Tidak | `""` | Filter: `unread`, `reading`, `read` |
| `type` | string | Tidak | `""` | Filter: `article`, `video`, `document`, `note` |
| `favorite` | boolean | Tidak | `false` | Filter: `true` hanya tampilkan item favorite |
| `page` | integer | Tidak | `1` | Nomor halaman |
| `limit` | integer | Tidak | `10` | Item per halaman (max 50) |
| `sort` | string | Tidak | `createdAt:desc` | Sort option |

**Contoh request:**
```
GET /api/pockets?search=react&status=unread&type=article&favorite=true&page=1&limit=10&sort=createdAt:desc
```

> **Catatan:** Endpoint ini juga digunakan untuk archive list dengan menambahkan `status=archived`. Ketika `status=archived` dikirim, filter `status != archived` diabaikan dan diganti dengan `status = archived`.

### Response 200 — Berhasil

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "title": "React Performance Guide",
      "url": "https://example.com/react-performance",
      "description": "A guide about React rendering optimization",
      "contentType": "article",
      "status": "unread",
      "isFavorite": true,
      "tags": ["frontend", "react"],
      "createdAt": "2026-06-26T10:00:00Z",
      "updatedAt": "2026-06-26T10:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "totalPage": 1
  }
}
```

### Response 200 — List Kosong (Empty State)

```json
{
  "data": [],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 0,
    "totalPage": 0
  }
}
```

### Response 401 — Tidak Terautentikasi

```json
{
  "code": "UNAUTHORIZED",
  "message": "Unauthorized"
}
```

---

## POST /api/pockets

Membuat pocket item baru.

**Auth Required:** Ya

### Request Body

```json
{
  "title": "React Performance Guide",
  "url": "https://example.com/react-performance",
  "description": "A guide about React rendering optimization",
  "contentType": "article",
  "tags": ["frontend", "react"]
}
```

| Field | Type | Required | Rule |
|---|---|---|---|
| `title` | string | Ya | Min 3, max 120 karakter |
| `url` | string | Kondisional | Wajib jika `contentType` = `article`, `video`, `document`. Format URL valid. |
| `description` | string | Tidak | Max 500 karakter |
| `contentType` | string | Ya | Nilai valid: `article`, `video`, `document`, `note` |
| `tags` | array of string | Tidak | Max 10 tags. Setiap tag max 24 karakter. Tidak boleh duplikat. |

### Response 201 — Berhasil Dibuat

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "title": "React Performance Guide",
    "url": "https://example.com/react-performance",
    "description": "A guide about React rendering optimization",
    "contentType": "article",
    "status": "unread",
    "isFavorite": false,
    "tags": ["frontend", "react"],
    "createdAt": "2026-06-26T10:00:00Z",
    "updatedAt": "2026-06-26T10:00:00Z"
  },
  "message": "Pocket item created successfully"
}
```

### Response 422 — Validation Error

```json
{
  "code": "VALIDATION_ERROR",
  "message": "Validation error",
  "details": [
    { "field": "title", "message": "Title is required" },
    { "field": "url",   "message": "URL is required" }
  ]
}
```

### Response 401 — Tidak Terautentikasi

```json
{
  "code": "UNAUTHORIZED",
  "message": "Unauthorized"
}
```

### Validation Rules

| Field | Rule | Error Message |
|---|---|---|
| `title` | required | `Title is required` |
| `title` | min 3 chars | `Title must be at least 3 characters` |
| `title` | max 120 chars | `Title must not exceed 120 characters` |
| `url` | required if type != note | `URL is required` |
| `url` | valid URL format | `URL is invalid` |
| `description` | max 500 chars | `Description must not exceed 500 characters` |
| `contentType` | required | `Content type is required` |
| `contentType` | valid enum | `Content type is invalid` |
| `tags` | max 10 items | `Maximum 10 tags allowed` |
| `tags[i]` | max 24 chars | `Tag must not exceed 24 characters` |
| `tags[i]` | no duplicate | `Duplicate tag is not allowed` |

---

## GET /api/pockets/:id

Mendapatkan detail satu pocket item.

**Auth Required:** Ya

### Path Parameter

| Param | Type | Deskripsi |
|---|---|---|
| `id` | string (UUID) | ID pocket item |

### Response 200 — Berhasil

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "title": "React Performance Guide",
    "url": "https://example.com/react-performance",
    "description": "A guide about React rendering optimization",
    "contentType": "article",
    "status": "unread",
    "isFavorite": true,
    "tags": ["frontend", "react"],
    "createdAt": "2026-06-26T10:00:00Z",
    "updatedAt": "2026-06-26T10:00:00Z"
  }
}
```

### Response 404 — Tidak Ditemukan

```json
{
  "code": "POCKET_NOT_FOUND",
  "message": "Pocket item not found"
}
```

> Response 404 dikembalikan baik ketika item tidak ada maupun ketika item ada tapi bukan milik user yang sedang login (security: jangan expose existence).

### Response 401 — Tidak Terautentikasi

```json
{
  "code": "UNAUTHORIZED",
  "message": "Unauthorized"
}
```

---

## PUT /api/pockets/:id

Mengupdate data pocket item.

**Auth Required:** Ya

### Path Parameter

| Param | Type | Deskripsi |
|---|---|---|
| `id` | string (UUID) | ID pocket item |

### Request Body

```json
{
  "title": "Updated React Performance Guide",
  "url": "https://example.com/react-performance-v2",
  "description": "Updated description",
  "contentType": "article",
  "tags": ["frontend", "react", "performance"]
}
```

> Semua field dikirim (full update / PUT semantics). Validation rules sama dengan POST /api/pockets.

### Response 200 — Berhasil Diupdate

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "title": "Updated React Performance Guide",
    "url": "https://example.com/react-performance-v2",
    "description": "Updated description",
    "contentType": "article",
    "status": "unread",
    "isFavorite": false,
    "tags": ["frontend", "react", "performance"],
    "createdAt": "2026-06-26T10:00:00Z",
    "updatedAt": "2026-06-26T11:00:00Z"
  },
  "message": "Pocket item updated successfully"
}
```

### Response 422 — Validation Error

```json
{
  "code": "VALIDATION_ERROR",
  "message": "Validation error",
  "details": [
    { "field": "url", "message": "URL is invalid" }
  ]
}
```

### Response 404 — Tidak Ditemukan

```json
{
  "code": "POCKET_NOT_FOUND",
  "message": "Pocket item not found"
}
```

### Response 401 — Tidak Terautentikasi

```json
{
  "code": "UNAUTHORIZED",
  "message": "Unauthorized"
}
```

---

## DELETE /api/pockets/:id

Mengarsipkan pocket item (soft delete — mengubah status menjadi `archived`).

**Auth Required:** Ya

### Path Parameter

| Param | Type | Deskripsi |
|---|---|---|
| `id` | string (UUID) | ID pocket item |

### Response 200 — Berhasil Diarsipkan

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001"
  },
  "message": "Pocket item archived successfully"
}
```

### Response 404 — Tidak Ditemukan

```json
{
  "code": "POCKET_NOT_FOUND",
  "message": "Pocket item not found"
}
```

### Response 401 — Tidak Terautentikasi

```json
{
  "code": "UNAUTHORIZED",
  "message": "Unauthorized"
}
```

---

## PATCH /api/pockets/:id/status

Mengupdate reading status pocket item.

**Auth Required:** Ya

### Path Parameter

| Param | Type | Deskripsi |
|---|---|---|
| `id` | string (UUID) | ID pocket item |

### Request Body

```json
{
  "status": "reading"
}
```

| Field | Type | Required | Rule |
|---|---|---|---|
| `status` | string | Ya | Nilai valid: `unread`, `reading`, `read` |

> Status `archived` tidak dapat diset melalui endpoint ini. Archive hanya melalui `DELETE /api/pockets/:id`.

### Response 200 — Berhasil Diupdate

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "status": "reading",
    "updatedAt": "2026-06-26T11:00:00Z"
  },
  "message": "Status updated successfully"
}
```

### Response 422 — Validation Error

```json
{
  "code": "VALIDATION_ERROR",
  "message": "Validation error",
  "details": [
    { "field": "status", "message": "Status is invalid" }
  ]
}
```

### Response 404 — Tidak Ditemukan

```json
{
  "code": "POCKET_NOT_FOUND",
  "message": "Pocket item not found"
}
```

### Validation Rules

| Field | Rule | Error Message |
|---|---|---|
| `status` | required | `Status is required` |
| `status` | valid enum (unread/reading/read) | `Status is invalid` |

---

## PATCH /api/pockets/:id/favorite

Toggle atau set status favorite pocket item.

**Auth Required:** Ya

### Path Parameter

| Param | Type | Deskripsi |
|---|---|---|
| `id` | string (UUID) | ID pocket item |

### Request Body

```json
{
  "isFavorite": true
}
```

| Field | Type | Required | Rule |
|---|---|---|---|
| `isFavorite` | boolean | Ya | `true` atau `false` |

### Response 200 — Berhasil Diupdate

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "isFavorite": true,
    "updatedAt": "2026-06-26T11:00:00Z"
  },
  "message": "Favorite updated successfully"
}
```

### Response 422 — Validation Error

```json
{
  "code": "VALIDATION_ERROR",
  "message": "Validation error",
  "details": [
    { "field": "isFavorite", "message": "isFavorite is required" }
  ]
}
```

### Response 404 — Tidak Ditemukan

```json
{
  "code": "POCKET_NOT_FOUND",
  "message": "Pocket item not found"
}
```

### Validation Rules

| Field | Rule | Error Message |
|---|---|---|
| `isFavorite` | required (boolean) | `isFavorite is required` |

---

## Route Summary

| Method | Endpoint | Auth | Deskripsi |
|---|---|---|---|
| POST | `/api/auth/login` | Tidak | Login user |
| POST | `/api/auth/logout` | Ya | Logout user |
| GET | `/api/dashboard` | Ya | Dashboard summary |
| GET | `/api/pockets` | Ya | List pocket items |
| POST | `/api/pockets` | Ya | Create pocket item |
| GET | `/api/pockets/:id` | Ya | Get pocket detail |
| PUT | `/api/pockets/:id` | Ya | Update pocket item |
| DELETE | `/api/pockets/:id` | Ya | Archive pocket item |
| PATCH | `/api/pockets/:id/status` | Ya | Update reading status |
| PATCH | `/api/pockets/:id/favorite` | Ya | Toggle favorite |

---

*Dokumen ini menjadi acuan untuk implementasi handler, service, dan API collection testing.*
