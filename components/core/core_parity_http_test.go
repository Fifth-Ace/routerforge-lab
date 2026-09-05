package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func parityFreeTCPAddress(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve TCP port: %v", err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatalf("release TCP port: %v", err)
	}
	return addr
}

func parityEnsureFrontend(t *testing.T) {
	t.Helper()

	root := filepath.Join("frontend", "build")
	index := filepath.Join(root, "index.html")

	if _, err := os.Stat(index); err == nil {
		return
	}

	if err := os.MkdirAll(filepath.Join(root, "_app", "immutable"), 0755); err != nil {
		t.Fatalf("mkdir frontend fixture: %v", err)
	}

	if err := os.WriteFile(
		index,
		[]byte("<!doctype html><html><body>routerforge-parity</body></html>"),
		0644,
	); err != nil {
		t.Fatalf("write frontend index fixture: %v", err)
	}

	asset := filepath.Join(root, "_app", "immutable", "parity.js")
	if err := os.WriteFile(asset, []byte("console.log('parity');"), 0644); err != nil {
		t.Fatalf("write immutable asset fixture: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Remove(asset)
		_ = os.Remove(index)
		_ = os.Remove(filepath.Join(root, "_app", "immutable"))
		_ = os.Remove(filepath.Join(root, "_app"))
		_ = os.Remove(root)
	})
}

func parityWaitHTTP(t *testing.T, base string) {
	t.Helper()

	client := &http.Client{Timeout: 500 * time.Millisecond}
	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		resp, err := client.Get(base + "/api/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("Core did not become ready at %s", base)
}

func parityRequest(t *testing.T, client *http.Client, method, url string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("%s %s: new request: %v", method, url, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	return resp
}

func parityReadBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return data
}

func TestCoreParityHTTPBaseline(t *testing.T) {
	parityEnsureFrontend(t)

	addr := parityFreeTCPAddress(t)
	base := "http://" + addr
	version := "parity-baseline"

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- startWeb(addr, version)
	}()

	parityWaitHTTP(t, base)

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	t.Run("health_get", func(t *testing.T) {
		resp := parityRequest(t, client, http.MethodGet, base+"/api/health")
		body := parityReadBody(t, resp)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status=%d body=%s", resp.StatusCode, body)
		}
		if got := resp.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
			t.Fatalf("Content-Type=%q", got)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-store" {
			t.Fatalf("Cache-Control=%q", got)
		}

		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("health JSON: %v: %s", err, body)
		}
		if payload["ok"] != true {
			t.Fatalf("health ok=%v", payload["ok"])
		}
		if payload["version"] != version {
			t.Fatalf("version=%v", payload["version"])
		}
		if payload["module_abi"] != "v1" {
			t.Fatalf("module_abi=%v", payload["module_abi"])
		}
	})

	t.Run("health_currently_accepts_post", func(t *testing.T) {
		resp := parityRequest(t, client, http.MethodPost, base+"/api/health")
		body := parityReadBody(t, resp)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("current baseline changed: POST /api/health status=%d body=%s",
				resp.StatusCode, body)
		}
	})

	t.Run("snapshot_method_policy", func(t *testing.T) {
		resp := parityRequest(t, client, http.MethodPost, base+"/api/snapshot")
		body := parityReadBody(t, resp)

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatalf("status=%d body=%s", resp.StatusCode, body)
		}
		if got := resp.Header.Get("Allow"); got != http.MethodGet {
			t.Fatalf("Allow=%q", got)
		}
	})

	t.Run("snapshot_without_dns_still_works", func(t *testing.T) {
		resp := parityRequest(t, client, http.MethodGet, base+"/api/snapshot")
		body := parityReadBody(t, resp)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status=%d body=%s", resp.StatusCode, body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-store" {
			t.Fatalf("Cache-Control=%q", got)
		}

		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("snapshot JSON: %v: %s", err, body)
		}
		if payload["version"] != version {
			t.Fatalf("version=%v", payload["version"])
		}
		if _, ok := payload["server_time"]; !ok {
			t.Fatal("server_time missing")
		}
		if _, ok := payload["dns_module_online"]; !ok {
			t.Fatal("dns_module_online missing")
		}
	})

	t.Run("catalog_method_policy", func(t *testing.T) {
		resp := parityRequest(t, client, http.MethodPost, base+"/api/catalog")
		body := parityReadBody(t, resp)

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatalf("status=%d body=%s", resp.StatusCode, body)
		}
		if got := resp.Header.Get("Allow"); got != http.MethodGet {
			t.Fatalf("Allow=%q", got)
		}
	})

	t.Run("catalog_baseline_shape", func(t *testing.T) {
		resp := parityRequest(t, client, http.MethodGet, base+"/api/catalog")
		body := parityReadBody(t, resp)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status=%d body=%s", resp.StatusCode, body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-store" {
			t.Fatalf("Cache-Control=%q", got)
		}

		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("catalog JSON: %v: %s", err, body)
		}

		for _, key := range []string{
			"generated_at",
			"read_only",
			"phase",
			"brand",
			"registry",
			"modules",
			"integrations",
			"package_management_enabled",
			"release",
		} {
			if _, ok := payload[key]; !ok {
				t.Fatalf("catalog field %q missing", key)
			}
		}
	})

	t.Run("unknown_module_is_404", func(t *testing.T) {
		resp := parityRequest(t, client, http.MethodGet, base+"/api/modules/no-such-module/health")
		body := parityReadBody(t, resp)

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status=%d body=%s", resp.StatusCode, body)
		}
	})

	t.Run("known_unavailable_module_is_503", func(t *testing.T) {
		resp := parityRequest(t, client, http.MethodGet, base+"/api/modules/system/health")
		body := parityReadBody(t, resp)

		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("status=%d body=%s", resp.StatusCode, body)
		}

		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("module error JSON: %v: %s", err, body)
		}

		if payload["module"] != "system" {
			t.Fatalf("module=%v", payload["module"])
		}
		if payload["running"] != false {
			t.Fatalf("running=%v", payload["running"])
		}
		if payload["mutation_api"] != false {
			t.Fatalf("mutation_api=%v", payload["mutation_api"])
		}
	})

	t.Run("admin_rejects_post", func(t *testing.T) {
		resp := parityRequest(t, client, http.MethodPost, base+"/api/admin/")
		body := parityReadBody(t, resp)

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatalf("status=%d body=%s", resp.StatusCode, body)
		}
		if got := resp.Header.Get("Allow"); got != http.MethodGet {
			t.Fatalf("Allow=%q", got)
		}

		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("admin error JSON: %v: %s", err, body)
		}
		if payload["mutation_api"] != false {
			t.Fatalf("mutation_api=%v", payload["mutation_api"])
		}
	})

	t.Run("spa_root", func(t *testing.T) {
		resp := parityRequest(t, client, http.MethodGet, base+"/")
		body := parityReadBody(t, resp)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status=%d body=%s", resp.StatusCode, body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-cache" {
			t.Fatalf("Cache-Control=%q", got)
		}
		if got := resp.Header.Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
			t.Fatalf("Content-Type=%q", got)
		}
	})

	t.Run("spa_fallback", func(t *testing.T) {
		resp := parityRequest(t, client, http.MethodGet, base+"/route/that/does/not/exist")
		body := parityReadBody(t, resp)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status=%d body=%s", resp.StatusCode, body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-cache" {
			t.Fatalf("Cache-Control=%q", got)
		}
	})

	t.Run("spa_non_get_does_not_fallback", func(t *testing.T) {
		resp := parityRequest(t, client, http.MethodPost, base+"/route/that/does/not/exist")
		body := parityReadBody(t, resp)

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status=%d body=%s", resp.StatusCode, body)
		}
	})

	t.Run("immutable_asset_cache", func(t *testing.T) {
		sub, err := frontendFS()
		if err != nil {
			t.Fatalf("frontendFS: %v", err)
		}

		var assetPath string
		err = fs.WalkDir(sub, "_app/immutable", func(p string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			assetPath = p
			return fs.SkipAll
		})
		if err != nil {
			t.Fatalf("walk immutable assets: %v", err)
		}
		if assetPath == "" {
			t.Fatal("no immutable frontend asset found")
		}

		resp := parityRequest(t, client, http.MethodGet, base+"/"+assetPath)
		body := parityReadBody(t, resp)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("asset=%s status=%d body=%s", assetPath, resp.StatusCode, body)
		}
		if got := resp.Header.Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
			t.Fatalf("asset=%s Cache-Control=%q", assetPath, got)
		}
	})

	t.Run("events_method_policy", func(t *testing.T) {
		resp := parityRequest(t, client, http.MethodPost, base+"/api/events")
		body := parityReadBody(t, resp)

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatalf("status=%d body=%s", resp.StatusCode, body)
		}
		if got := resp.Header.Get("Allow"); got != http.MethodGet {
			t.Fatalf("Allow=%q", got)
		}
	})

	t.Run("events_initial_frame", func(t *testing.T) {
		streamClient := &http.Client{Timeout: 6 * time.Second}
		resp := parityRequest(t, streamClient, http.MethodGet, base+"/api/events?interval_ms=1")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("status=%d body=%s", resp.StatusCode, body)
		}

		if got := resp.Header.Get("Content-Type"); got != "text/event-stream" {
			t.Fatalf("Content-Type=%q", got)
		}
		if got := resp.Header.Get("Cache-Control"); got != "no-cache, no-transform" {
			t.Fatalf("Cache-Control=%q", got)
		}
		if got := resp.Header.Get("X-Accel-Buffering"); got != "no" {
			t.Fatalf("X-Accel-Buffering=%q", got)
		}

		reader := bufio.NewReader(resp.Body)

		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read retry line: %v", err)
		}
		if line != "retry: 1500\n" {
			t.Fatalf("retry line=%q", line)
		}

		line, err = reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read retry separator: %v", err)
		}
		if line != "\n" {
			t.Fatalf("retry separator=%q", line)
		}

		line, err = reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read event line: %v", err)
		}
		if line != "event: snapshot\n" {
			t.Fatalf("event line=%q", line)
		}

		line, err = reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read data line: %v", err)
		}
		if !strings.HasPrefix(line, "data: {") {
			t.Fatalf("data line=%q", line)
		}
	})

	select {
	case err := <-serverDone:
		if err == nil {
			t.Fatal("startWeb unexpectedly returned nil")
		}
	default:
	}

	fmt.Println("Core HTTP parity baseline passed")
}
