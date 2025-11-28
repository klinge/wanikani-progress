# API Validation Rules

This document describes the validation rules for all API endpoints.

## GET /api/subjects

### Query Parameters

| Parameter | Type | Required | Validation | Example |
|-----------|------|----------|------------|---------|
| `type` | string | No | Must be one of: `radical`, `kanji`, `vocabulary` | `?type=kanji` |
| `level` | integer | No | Must be between 1 and 60 | `?level=5` |

### Example Requests

Valid:
```
GET /api/subjects
GET /api/subjects?type=kanji
GET /api/subjects?level=5
GET /api/subjects?type=vocabulary&level=10
```

Invalid:
```
GET /api/subjects?type=invalid     # 400: type must be radical, kanji, or vocabulary
GET /api/subjects?level=0          # 400: level must be between 1 and 60
GET /api/subjects?level=61         # 400: level must be between 1 and 60
GET /api/subjects?level=abc        # 400: level must be a valid integer
```

## GET /api/assignments

### Query Parameters

| Parameter | Type | Required | Validation | Example |
|-----------|------|----------|------------|---------|
| `srs_stage` | integer | No | Must be between 0 and 9 | `?srs_stage=4` |

### SRS Stage Values

- 0: Initiate
- 1-4: Apprentice (I-IV)
- 5-6: Guru (I-II)
- 7: Master
- 8: Enlightened
- 9: Burned

### Example Requests

Valid:
```
GET /api/assignments
GET /api/assignments?srs_stage=0
GET /api/assignments?srs_stage=9
```

Invalid:
```
GET /api/assignments?srs_stage=-1   # 400: srs_stage must be between 0 and 9
GET /api/assignments?srs_stage=10   # 400: srs_stage must be between 0 and 9
GET /api/assignments?srs_stage=abc  # 400: srs_stage must be a valid integer
```

## GET /api/reviews

### Query Parameters

| Parameter | Type | Required | Validation | Example |
|-----------|------|----------|------------|---------|
| `from` | date | No | Must be in YYYY-MM-DD format | `?from=2024-01-01` |
| `to` | date | No | Must be in YYYY-MM-DD format | `?to=2024-01-31` |

### Additional Validation

- If both `from` and `to` are provided, `from` must be before or equal to `to`

### Example Requests

Valid:
```
GET /api/reviews
GET /api/reviews?from=2024-01-01
GET /api/reviews?to=2024-01-31
GET /api/reviews?from=2024-01-01&to=2024-01-31
GET /api/reviews?from=2024-01-01&to=2024-01-01  # Same date is valid
```

Invalid:
```
GET /api/reviews?from=2024/01/01              # 400: from must be in YYYY-MM-DD format
GET /api/reviews?to=2024-01-32                # 400: to must be in YYYY-MM-DD format
GET /api/reviews?from=2024-01-31&to=2024-01-01  # 400: from must be before or equal to to
```

## GET /api/statistics

### Query Parameters

| Parameter | Type | Required | Validation | Example |
|-----------|------|----------|------------|---------|
| `from` | date | No | Must be in YYYY-MM-DD format | `?from=2024-01-01` |
| `to` | date | No | Must be in YYYY-MM-DD format | `?to=2024-01-31` |

### Additional Validation

- If both `from` and `to` are provided, `from` must be before or equal to `to`

### Example Requests

Valid:
```
GET /api/statistics
GET /api/statistics?from=2024-01-01
GET /api/statistics?to=2024-01-31
GET /api/statistics?from=2024-01-01&to=2024-01-31
```

Invalid:
```
GET /api/statistics?from=2024/01/01              # 400: from must be in YYYY-MM-DD format
GET /api/statistics?from=2024-01-31&to=2024-01-01  # 400: from must be before or equal to to
```

## GET /api/statistics/latest

No query parameters. Returns the most recent statistics snapshot.

### Example Requests

Valid:
```
GET /api/statistics/latest
```

## POST /api/sync

No query parameters. Triggers a manual sync operation.

### Example Requests

Valid:
```
POST /api/sync
```

## GET /api/sync/status

No query parameters. Returns the current sync status.

### Example Requests

Valid:
```
GET /api/sync/status
```

## Error Response Format

All validation errors return a 400 Bad Request status with the following JSON structure:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid query parameters",
    "details": {
      "field_name": "Specific error message for this field"
    }
  }
}
```

### Example Error Responses

Invalid level:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid query parameters",
    "details": {
      "level": "Must be between 1 and 60"
    }
  }
}
```

Invalid date format:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid query parameters",
    "details": {
      "from": "Must be in YYYY-MM-DD format"
    }
  }
}
```

Invalid date range:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid query parameters",
    "details": {
      "from": "Must be before or equal to 'to' date"
    }
  }
}
```

## Other Error Types

### Authentication Error (401)
```json
{
  "error": {
    "code": "AUTH_ERROR",
    "message": "Authentication failed",
    "details": {
      "detail": "Invalid or missing API token"
    }
  }
}
```

### Network Error (503)
```json
{
  "error": {
    "code": "NETWORK_ERROR",
    "message": "Unable to connect to WaniKani API",
    "details": {
      "detail": "Please check your network connection and try again"
    }
  }
}
```

### Rate Limit Error (429)
```json
{
  "error": {
    "code": "RATE_LIMIT_ERROR",
    "message": "Rate limit exceeded",
    "details": {
      "detail": "Too many requests to WaniKani API. Please try again later"
    }
  }
}
```

### Not Found (404)
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "No statistics found"
  }
}
```

### Sync In Progress (409)
```json
{
  "error": {
    "code": "SYNC_IN_PROGRESS",
    "message": "A sync operation is already in progress"
  }
}
```

### Internal Server Error (500)
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An internal error occurred"
  }
}
```
