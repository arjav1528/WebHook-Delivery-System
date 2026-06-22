# Webhook Delivery System - HTTP Request Schema

## Base URL
```
http://localhost:3000
```

---

## 1. Health Check

```
GET /
```

### Response
```json
{
  "message": "Hello World"
}
```

---

## 2. Register Webhook

```
POST /webhooks
Content-Type: application/json
```

### Request Body
```json
{
  "url": "https://example.com/webhook",
  "event": "user.created"
}
```

### Response (200 OK)
```json
{
  "message": "Webhook registered successfully",
  "id": "507f1f77bcf86cd799439011",
  "secret": "K7f9mZ2nX5pL8qR4tW6vYaB3cD1eF9gH2jJ5kL7mN9oP2qR4s"
}
```

### Supported Events
- `user.created`
- `user.deleted`

---

## 3. Trigger Event

```
POST /events
Content-Type: application/json
```

### Request Body
```json
{
  "type": "user.created",
  "data": {
    "user_id": "123456",
    "name": "Alice Johnson",
    "email": "alice@example.com"
  }
}
```

### Response (202 Accepted)
```json
{
  "message": "event accepted",
  "event_id": "507f1f77bcf86cd799439012",
  "webhooks": 3
}
```

---

## 4. Get Delivery Status

```
GET /deliveries/:id
```

### Response (200 OK)
```json
{
  "delivery_id": "507f1f77bcf86cd799439013",
  "status": "success",
  "retry_count": 0,
  "last_error": "",
  "created_at": "2024-01-15T10:30:45Z",
  "updated_at": "2024-01-15T10:30:47Z"
}
```

### Status Values
- `pending` - Not yet processed or scheduled for retry
- `success` - Delivered successfully (2xx response)
- `failed` - Failed after all retries exhausted

---

## Webhook Payload Format

Every webhook POST request receives:

```
POST {webhook_url}
Content-Type: application/json
X-Webhook-Signature: sha256=<HMAC_SIGNATURE>
```

### Body
```json
{
  "event_id": "507f1f77bcf86cd799439012",
  "type": "user.created",
  "data": {
    "user_id": "123456",
    "name": "Alice Johnson",
    "email": "alice@example.com"
  }
}
```

---

## HMAC Signature

**Header:** `X-Webhook-Signature`  
**Format:** `sha256=<hex_encoded_hash>`

### Verification (Node.js)
```javascript
const crypto = require('crypto');

const signature = req.headers['x-webhook-signature'];
const secret = 'YOUR_WEBHOOK_SECRET';
const payload = JSON.stringify(req.body);

const hash = crypto.createHmac('sha256', secret)
  .update(payload)
  .digest('hex');

const expectedSignature = 'sha256=' + hash;

if (signature !== expectedSignature) {
  throw new Error('Invalid signature');
}
```

### Verification (Python)
```python
import hmac
import hashlib

signature = request.headers.get('X-Webhook-Signature')
secret = 'YOUR_WEBHOOK_SECRET'
payload = request.get_data()

expected_sig = 'sha256=' + hmac.new(
  secret.encode(),
  payload,
  hashlib.sha256
).hexdigest()

if signature != expected_sig:
  raise ValueError('Invalid signature')
```

### Verification (Go)
```go
import (
  "crypto/hmac"
  "crypto/sha256"
  "encoding/hex"
)

signature := r.Header.Get("X-Webhook-Signature")
secret := "YOUR_WEBHOOK_SECRET"
payload := body

h := hmac.New(sha256.New, []byte(secret))
h.Write(payload)
expected := "sha256=" + hex.EncodeToString(h.Sum(nil))

if signature != expected {
  return fmt.Errorf("invalid signature")
}
```

---

## HTTP Status Codes

| Code | Meaning |
|------|---------|
| `200` | OK - Success |
| `202` | Accepted - Event queued |
| `400` | Bad Request - Invalid input |
| `404` | Not Found - Resource doesn't exist |
| `500` | Server Error |

---

## Error Responses

```json
{
  "error": "error message here"
}
```

### Common Errors

**Invalid URL:**
```json
{
  "error": "URL must start with http:// or https://"
}
```

**Invalid Event Type:**
```json
{
  "error": "invalid event type: invalid.event. allowed: user.created, user.deleted"
}
```

**Delivery Not Found:**
```json
{
  "error": "delivery not found"
}
```

**Invalid Delivery ID:**
```json
{
  "error": "invalid delivery ID"
}
```

---

## curl Examples

### Register Webhook
```bash
curl -X POST http://localhost:3000/webhooks \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://webhook.site/unique-uuid",
    "event": "user.created"
  }'
```

### Trigger Event
```bash
curl -X POST http://localhost:3000/events \
  -H "Content-Type: application/json" \
  -d '{
    "type": "user.created",
    "data": {
      "user_id": "123",
      "name": "Alice"
    }
  }'
```

### Get Delivery Status
```bash
curl http://localhost:3000/deliveries/507f1f77bcf86cd799439013
```