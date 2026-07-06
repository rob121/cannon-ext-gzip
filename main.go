package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/rob121/cannon/extension"
)

const (
	extensionName        = "cannon-ext-gzip"
	extensionTitle       = "Gzip Compression"
	extensionDescription = "Compresses template-rendered HTML responses with gzip through Cannon's onAfterRender hook when browsers advertise gzip support."
	extensionVersion     = "0.1.1"
	updateURLBase        = "https://github.com/rob121/cannon-ext-gzip/releases/download"
	onAfterRender        = "onAfterRender"
)

func main() {
	if err := newServer().Run(); err != nil {
		log.Fatal(err)
	}
}

func newServer() *extension.Server {
	s := extension.New(extension.Info{
		Name:          extensionName,
		Title:         extensionTitle,
		Description:   extensionDescription,
		Version:       extensionVersion,
		UpdateURLBase: updateURLBase,
	})
	s.OnHook(onAfterRender, gzipAfterRender)
	return s
}

func gzipAfterRender(req extension.HookWireRequest) extension.HookWireResponse {
	if !acceptsGzip(req.Header) {
		return extension.HookOK(nil)
	}

	args := extension.HookArguments(req)
	body, _ := args["body"].(string)
	if body == "" || isEncoded(args) || !compressible(args) {
		return extension.HookOK(nil)
	}

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write([]byte(body)); err != nil {
		return extension.HookOK(nil)
	}
	if err := zw.Close(); err != nil {
		return extension.HookOK(nil)
	}

	return extension.HookOK(map[string]any{
		"body_base64":   base64.StdEncoding.EncodeToString(buf.Bytes()),
		"body_encoding": "base64",
		"headers": map[string][]string{
			"Content-Encoding": {"gzip"},
			"Content-Length":   {strconv.Itoa(buf.Len())},
			"Vary":             {"Accept-Encoding"},
		},
	})
}

func acceptsGzip(header http.Header) bool {
	for key, vals := range header {
		if !strings.EqualFold(key, "Accept-Encoding") {
			continue
		}
		for _, val := range vals {
			for _, part := range strings.Split(val, ",") {
				if strings.EqualFold(strings.TrimSpace(strings.Split(part, ";")[0]), "gzip") {
					return true
				}
			}
		}
	}
	return false
}

func isEncoded(args map[string]any) bool {
	headers := hookHeaders(args)
	for key, vals := range headers {
		if strings.EqualFold(key, "Content-Encoding") && len(vals) > 0 && vals[0] != "" {
			return true
		}
	}
	return false
}

func compressible(args map[string]any) bool {
	headers := hookHeaders(args)
	contentType := ""
	for key, vals := range headers {
		if strings.EqualFold(key, "Content-Type") && len(vals) > 0 {
			contentType = vals[0]
			break
		}
	}
	if contentType == "" {
		return true
	}
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "text/") ||
		strings.Contains(contentType, "json") ||
		strings.Contains(contentType, "javascript") ||
		strings.Contains(contentType, "xml") ||
		strings.Contains(contentType, "svg")
}

func hookHeaders(args map[string]any) http.Header {
	headers := http.Header{}
	raw, _ := args["headers"].(map[string]any)
	if raw == nil {
		raw, _ = args["header"].(map[string]any)
	}
	for key, val := range raw {
		switch v := val.(type) {
		case []string:
			headers[key] = append(headers[key], v...)
		case []any:
			for _, item := range v {
				if s, ok := item.(string); ok {
					headers.Add(key, s)
				}
			}
		case string:
			headers.Set(key, v)
		}
	}
	return headers
}
