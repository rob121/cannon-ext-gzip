# Cannon Gzip Extension

This extension subscribes to Cannon's `onAfterRender` hook and returns a gzip-compressed response body when the request advertises `Accept-Encoding: gzip`.

## Cannon Hook Contract

Cannon's `onAfterRender` hook must behave like a final response mutation hook:

1. Render the page/layout into a buffer first.
2. Fire `onAfterRender` before writing to the client.
3. Include hook arguments:
   - `headers`
   - `body`
   - `layout`
   - `page`
4. Apply returned hook arguments:
   - `body` replaces the response body as UTF-8 text.
   - `body_base64` replaces the response body as binary bytes.
   - `body_encoding: "base64"` identifies base64 binary payloads.
   - `headers` are merged into the outgoing response headers.

The base64 path is necessary because extension hooks are JSON over a Unix socket, and gzip output is binary.

## Behavior

- Skips requests that do not include `Accept-Encoding: gzip`.
- Skips responses that already have a `Content-Encoding`.
- Compresses text-like content types such as HTML, JSON, XML, JavaScript, SVG, and other `text/*` responses.
- Returns `Content-Encoding: gzip`, `Vary: Accept-Encoding`, `Content-Length`, and a base64-encoded compressed body.
