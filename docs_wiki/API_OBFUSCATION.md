# API Obfuscation Mechanism

## Overview

The API Obfuscation feature is designed to obscure JSON data in transit between the backend and the frontend. This adds a layer of stealth to the monitoring system, making traffic analysis more difficult for casual observers. The system uses Base64 encoding for the response body of specific API endpoints.

## Configuration

Obfuscation is controlled via the `config.yaml` file.

```yaml
monitoring:
  obfuscate_api: true  # Set to true to enable, false to disable
```

If enabled, the backend will automatically encode eligible API responses. If disabled, the backend serves standard JSON, and the frontend transparently handles the standard response.

## Backend Implementation

The core logic resides in `internal/monitoring/middleware/obfuscator.go`.

### Middleware Logic

1.  **Scope**: The middleware intercepts requests starting with `/api/`.
2.  **Exclusions**: The following paths are explicitly excluded from obfuscation to support streaming or static content:
    *   `/api/logs` (SSE stream)
    *   `/api/cpu` (SSE stream)
    *   `/api/user/photos` (Binary/Static)
3.  **Content Negotiation**: The middleware only processes responses where the `Content-Type` includes `application/json`. `text/event-stream` is skipped.
4.  **Encoding**: The response body is read into a buffer, encoded using Standard Base64 (padding with `=`), and written back to the response.
5.  **Headers**:
    *   `X-Obfuscated: true` is set to indicate the response is encoded.
    *   `Content-Length` is updated to reflect the size of the encoded body to prevent truncation or keep-alive issues.

### Code Reference

See `Obfuscator` function in `internal/monitoring/middleware/obfuscator.go`.

## Frontend Implementation

The frontend handles de-obfuscation transparently using a global `window.fetch` interceptor in `web/monitoring/assets/js/app.js`.

### Interceptor Strategy

The interceptor wraps the native `fetch` API and applies a "Parse First, Decode Second" strategy to ensure robustness.

1.  **Check Content Type**: Ignores non-JSON and `text/event-stream` responses.
2.  **Strategy 1: Try Parse**:
    *   Attempts to `JSON.parse()` the response body directly.
    *   If successful, the response is standard JSON (not obfuscated). It returns the original response.
3.  **Strategy 2: Try Decode (Fallback)**:
    *   If parsing fails (typical for Base64 strings), it attempts to decode the body.
    *   **Normalization**: Replaces URL-safe characters (`-` to `+`, `_` to `/`) and removes whitespace.
    *   **Padding**: Ensures the string length is a multiple of 4 by adding `=` padding.
    *   **Decoding**: Uses `atob()` and `TextDecoder` (UTF-8) to convert the Base64 string back to text.
    *   **Verification**: Attempts to `JSON.parse()` the decoded string.
    *   If valid JSON is found, a new `Response` object is created with the decoded content and returned.
4.  **Final Fallback**:
    *   If both strategies fail, the original body is returned.

This approach ensures the frontend continues to work seamlessly whether obfuscation is enabled or disabled, or if headers are stripped by proxies.

## Troubleshooting

### "SyntaxError: Unexpected token 'e', ..."
This error occurs when the frontend tries to parse the raw Base64 string as JSON. This usually indicates the interceptor failed to detect or decode the obfuscated response. The "Parse First, Decode Second" strategy resolves this by specifically catching parse errors and triggering the decode logic.

### Truncated Data / Network Errors
If the `Content-Length` header does not match the actual body size (e.g., if the body size changed due to encoding but the header remained the original size), browsers may truncate the response. The middleware explicitly sets the correct `Content-Length` of the encoded body.

### CORS and Headers
The `X-Obfuscated` header is exposed in the CORS configuration (`internal/monitoring/server.go`) to allow the frontend to detect encryption status explicitly, though the heuristic fallback logic ensures functionality even if this header is stripped.
