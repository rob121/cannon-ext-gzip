# Cannon Gzip Extension

Compresses rendered Cannon responses with gzip by subscribing to Cannon's `onAfterRender` hook.

## Features

- Compresses text-like responses when the browser advertises `Accept-Encoding: gzip`.
- Skips responses that already have a `Content-Encoding` header.
- Returns gzip bytes through Cannon's hook response contract using base64 encoding.
- Sets `Content-Encoding: gzip`, `Vary: Accept-Encoding`, and `Content-Length`.

## Cannon Capabilities

This extension exposes:

- `/meta`: extension name, description, version, and update URL base.
- `/capabilities`: hook capability.
- `/hooks`: subscribes to `onAfterRender`.

## Hook Behavior

Cannon should fire `onAfterRender` after rendering the response body and before writing it to the client. The extension expects hook arguments such as:

- `headers`
- `body`
- `layout`
- `page`

When compression is applied, the extension returns:

- `body_base64`: gzipped response body.
- `body_encoding: "base64"`: identifies binary payload encoding.
- `headers`: response headers to merge into the outgoing response.

## Configuration

No configuration is required.

## Releases

The extension reports:

- Name: `cannon-ext-gzip`
- Version: `0.1.1`
- Update URL base: `https://github.com/rob121/cannon-ext-gzip/releases/download`

GitHub releases should include a `cannon-extension.json` manifest plus platform binaries.
