package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

func TestInstallWithProgressReportsBytesAndPhases(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("tar.gz test")
	}
	// Build a minimal tar.gz containing a fake macron binary
	var tarBuf bytes.Buffer
	gz := gzip.NewWriter(&tarBuf)
	tw := tar.NewWriter(gz)
	fakeBin := bytes.Repeat([]byte("macron-bin-data-"), 4096) // ~64KB
	if err := tw.WriteHeader(&tar.Header{Name: "macron", Mode: 0o755, Size: int64(len(fakeBin))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(fakeBin); err != nil {
		t.Fatal(err)
	}
	_ = tw.Close()
	_ = gz.Close()
	archive := tarBuf.Bytes()
	sum := sha256.Sum256(archive)
	asset := AssetName("9.9.9")
	sumsBody := hex.EncodeToString(sum[:]) + "  " + asset + "\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/archive":
			w.Header().Set("Content-Length", strconv.Itoa(len(archive)))
			_, _ = w.Write(archive)
		case "/sums":
			_, _ = w.Write([]byte(sumsBody))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	fakeSelf := filepath.Join(dir, "macron")
	if err := os.WriteFile(fakeSelf, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	origSelf, origPreflight := SelfPath, Preflight
	SelfPath = func() (string, error) { return fakeSelf, nil }
	Preflight = func(path, version string) error { return nil }
	defer func() {
		SelfPath = origSelf
		Preflight = origPreflight
	}()

	rel := &Release{
		TagName: "v9.9.9",
		Assets: []struct {
			Name   string `json:"name"`
			URL    string `json:"browser_download_url"`
			APIURL string `json:"url"`
			Size   int64  `json:"size"`
		}{
			{Name: asset, URL: srv.URL + "/archive", Size: int64(len(archive))},
			{Name: "SHA256SUMS", URL: srv.URL + "/sums", Size: int64(len(sumsBody))},
		},
	}

	var events []Progress
	if err := InstallWithProgress(context.Background(), rel, func(p Progress) {
		events = append(events, p)
	}); err != nil {
		t.Fatalf("InstallWithProgress failed: %v", err)
	}
	if len(events) < 4 {
		t.Fatalf("expected at least 4 progress events, got %d: %+v", len(events), events)
	}
	last := events[len(events)-1]
	if last.Phase != "done" || last.Percent != 100 || last.Downloaded != int64(len(archive)) {
		t.Fatalf("unexpected final progress: %+v", last)
	}
}
