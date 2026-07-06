package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rob121/cannon/extension"
)

func TestCapabilitiesExposeHooks(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/capabilities", nil)
	rec := httptest.NewRecorder()

	newServer().Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var got struct {
		Capabilities map[string]string `json:"capabilities"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode capabilities: %v", err)
	}
	if got.Capabilities["hooks"] != "/hooks" {
		t.Fatalf("expected hooks capability /hooks, got %#v", got.Capabilities)
	}
}

func TestMetaIncludesDescription(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/meta", nil)
	rec := httptest.NewRecorder()

	newServer().Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var got struct {
		Name        string `json:"name"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Version     string `json:"version"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode meta: %v", err)
	}
	if got.Name != extensionName || got.Title != extensionTitle || got.Description != extensionDescription || got.Version != extensionVersion {
		t.Fatalf("unexpected meta: %+v", got)
	}
}

func TestHooksListSubscribesToAfterRender(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hooks", nil)
	rec := httptest.NewRecorder()

	newServer().Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var got extension.HookListResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode hooks: %v", err)
	}
	if len(got.Hooks) != 1 || got.Hooks[0] != onAfterRender {
		t.Fatalf("expected onAfterRender subscription, got %#v", got.Hooks)
	}
}

func TestAfterRenderGzipsBodyWhenAccepted(t *testing.T) {
	resp := gzipAfterRender(extension.HookWireRequest{
		WireRequest: extension.WireRequest{
			Header: http.Header{"Accept-Encoding": {"br, gzip"}},
		},
		Event: onAfterRender,
		Arguments: map[string]any{
			"body": "hello world",
			"headers": map[string]any{
				"Content-Type": []any{"text/html; charset=utf-8"},
			},
		},
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if resp.Arguments["body_encoding"] != "base64" {
		t.Fatalf("expected base64 body encoding, got %#v", resp.Arguments)
	}
	raw, err := base64.StdEncoding.DecodeString(resp.Arguments["body_base64"].(string))
	if err != nil {
		t.Fatalf("decode gzip body: %v", err)
	}
	reader, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("open gzip body: %v", err)
	}
	out, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read gzip body: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close gzip body: %v", err)
	}
	if string(out) != "hello world" {
		t.Fatalf("expected decompressed body %q, got %q", "hello world", string(out))
	}

	headers := resp.Arguments["headers"].(map[string][]string)
	if got := strings.Join(headers["Content-Encoding"], ","); got != "gzip" {
		t.Fatalf("expected gzip content encoding, got %q", got)
	}
	if got := strings.Join(headers["Vary"], ","); got != "Accept-Encoding" {
		t.Fatalf("expected vary Accept-Encoding, got %q", got)
	}
}

func TestAfterRenderSkipsWhenGzipNotAccepted(t *testing.T) {
	resp := gzipAfterRender(extension.HookWireRequest{
		WireRequest: extension.WireRequest{
			Header: http.Header{"Accept-Encoding": {"br"}},
		},
		Event:     onAfterRender,
		Arguments: map[string]any{"body": "hello world"},
	})

	if len(resp.Arguments) != 0 {
		t.Fatalf("expected no argument updates, got %#v", resp.Arguments)
	}
}
