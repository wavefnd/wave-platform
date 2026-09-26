# Webhooks

Wave sends webhook events as HTTP `POST` requests to configured endpoints.

## Generic Webhook Delivery

For generic webhook endpoints, the request body is the JSON representation of the event.

### Request Headers

Each webhook request includes:

* `Content-Type: application/json`
* `User-Agent: Wave-Platform-Webhook/1.0`
* `X-Wave-Event`: the event type.
* `X-Wave-Delivery`: the unique delivery ID.
* `X-Wave-Signature-256`: an HMAC-SHA256 signature of the exact request body bytes.

The signature has the following format:

```text
sha256=<lowercase hexadecimal HMAC-SHA256 digest>
```

The signature must be calculated using the webhook signing secret and the **exact bytes received in the HTTP request body**.

Do not parse and re-serialize the JSON before verifying the signature, because even semantically equivalent JSON can have different byte representations.

## Event Payload

A generic webhook request contains the following JSON fields:

```json
{
  "id": "evt_123",
  "type": "blog.published",
  "title": "Example event",
  "summary": "Example summary",
  "author_name": "Example Author",
  "image_url": "https://example.com/image.webp",
  "resource_id": "resource_123",
  "url": "https://example.com/resource",
  "occurred_at": "2026-09-26T10:00:00Z"
}
```

The fields are:

| Field         | Description                                   |
| ------------- | --------------------------------------------- |
| `id`          | Unique event ID.                              |
| `type`        | Event type.                                   |
| `title`       | Event title.                                  |
| `summary`     | Event summary. Omitted when empty.            |
| `author_name` | Author name. Omitted when empty.              |
| `image_url`   | Image URL. Omitted when empty.                |
| `resource_id` | ID of the resource associated with the event. |
| `url`         | URL of the associated resource.               |
| `occurred_at` | Time when the event occurred.                 |

`summary`, `author_name`, and `image_url` use `omitempty`, so they are not included in the JSON when they are empty.

## Signature Verification

Receivers should verify `X-Wave-Signature-256` before processing the webhook.

The signature is an HMAC-SHA256 digest calculated over the exact request body bytes using the webhook signing secret.

A minimal Go receiver can verify it as follows:

```go
func verifySignature(body []byte, header string, secret string) bool {
	if !strings.HasPrefix(header, "sha256=") {
		return false
	}

	received, err := hex.DecodeString(strings.TrimPrefix(header, "sha256="))
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := mac.Sum(nil)

	return hmac.Equal(expected, received)
}
```

The request body should be read as raw bytes and passed directly to the verification function. The JSON should not be parsed and re-serialized before verification.

Use a constant-time comparison such as `hmac.Equal` when comparing the received signature with the calculated digest.

## Signature Test Vector

The following synthetic values can be used to test webhook signature verification. They do not use a real webhook secret or destination.

**Signing secret:**

```text
test-secret
```

**Request body:**

```json
{"id":"evt_test_001","type":"blog.published","title":"Test event","resource_id":"resource_001","url":"https://example.test/events/1","occurred_at":"2026-09-26T10:00:00Z"}
```

**Expected `X-Wave-Signature-256` header:**

```text
sha256=ca0c1649a062d5d3a5d558a53e3c05f727d9104907edc55f9c716f378c07c0b9
```

The signature is calculated over the exact UTF-8 bytes of the request body shown above using HMAC-SHA256 with the signing secret `test-secret`.

Changing whitespace, field ordering, or any other byte in the request body changes the resulting signature.

## Delivery and Retries

Wave considers a webhook delivery successful when the receiver returns any HTTP `2xx` status code.

If a delivery fails, Wave retries it. A delivery can be attempted up to five times. After the fifth failed attempt, the delivery is marked as failed.

Before the final attempt, failed deliveries are deferred and scheduled for another attempt using an increasing delay.

Receivers should therefore be prepared to receive the same event more than once. Webhook handling should be idempotent, using the event ID (`id`) to detect an event that has already been processed.

The `X-Wave-Delivery` header identifies the delivery record, while the event's `id` identifies the event itself. The same delivery ID can be present across retry attempts.

### Retry Schedule

For failed attempts before the final attempt, the next retry is scheduled with an increasing delay:

| Failed attempt |                Next retry |
| -------------- | ------------------------: |
| 1              |                 2 minutes |
| 2              |                 4 minutes |
| 3              |                 8 minutes |
| 4              |                16 minutes |
| 5              | Delivery marked as failed |

These delays are calculated from the time of the failed attempt.

## Event Scopes

Webhook endpoints can be either **account-scoped** or **platform-scoped**.

Account-scoped endpoints receive events available to users within an account. Platform-scoped endpoints can also receive platform-only events.

The following events are platform-only:

* `mailing-list.post`
* `patch.received`

Platform-only events are not delivered to account-scoped endpoints.

Account-scoped endpoints cannot be used to receive platform-only events. Platform-scoped endpoints are required for those events.

The event type is available in both the JSON payload's `type` field and the `X-Wave-Event` header.

## Replay Protection

Webhook signatures are calculated from the request body and signing secret. The webhook request does not include a timestamp header.

Wave therefore does not currently provide timestamp-based replay protection at the webhook verification layer.

Receivers that require replay protection should track previously processed delivery or event IDs and apply their own freshness or deduplication policy.

In particular, receivers should not assume that a valid signature means the request is a newly generated request. A previously captured request with a valid signature can still pass signature verification.

## Discord Webhooks

Discord endpoints receive a Discord webhook payload rather than the generic event JSON.

The payload contains:

* `username`: `Wave`
* `allowed_mentions`: configured with an empty `parse` list.
* `embeds`: an array containing the event embed.

The embed includes the event title, URL, timestamp, footer, and description. The author and image are included when available.

### Text Truncation

Discord webhook text is normalized before delivery:

* The title is trimmed and limited to 256 Unicode characters.
* The author name is trimmed and limited to 256 Unicode characters.
* The summary is converted to a preview, with Markdown image syntax removed and whitespace normalized, then limited to 120 Unicode characters.
* Empty descriptions are omitted from the Discord payload.

### Image URLs

For Discord deliveries, Wave only includes image URLs that match its supported public Wave image path format. When a valid image is available, Wave sends the resulting public URL as the Discord embed image URL.

Wave controls the URL included in the Discord request, but it does not guarantee how Discord fetches or caches the referenced image.

## Pending Lifecycle Changes

Webhook lifecycle behavior is still subject to the pending changes tracked in:

* [Issue #15](https://github.com/wavefnd/wave-platform/issues/15)
* [Issue #16](https://github.com/wavefnd/wave-platform/issues/16)

This document describes the webhook behavior present in the current implementation and does not attempt to define behavior that is still being changed by those issues.
