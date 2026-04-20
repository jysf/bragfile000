# API Contract

> Replace with your actual API contract. If your project has no external
> API, delete this file.

## Overview

[REPLACE: Base URL, auth mechanism, versioning scheme.]

## Endpoints

### `GET /resource/:id`

[REPLACE: Description]

**Request:**
- Params: `id` (uuid)
- Headers: `Authorization: Bearer <token>`

**Response (200):**
```json
{ "id": "...", "field": "..." }
```

**Errors:**
- `404` — Resource not found
- `403` — Not authorized
- `401` — Not authenticated

## Error Response Shape

```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Human-readable message",
    "details": {}
  }
}
```

## Rate Limits

[REPLACE]

## Auth

[REPLACE: reference the DEC-* that formalized the auth decision.]
