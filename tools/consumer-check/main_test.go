package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestCheckOutput(t *testing.T) {
	cases := []struct {
		name string
		out  string
		ok   bool
	}{
		{"exact", wantOutput, true},
		{"trailing newline", wantOutput + "\n", true},
		{"wrong AD date", strings.Replace(wantOutput, "2024-07-30", "2024-07-31", 1), false},
		{"wrong event", strings.Replace(wantOutput, "Kheer Khane Din", "Nothing", 1), false},
		{"missing event", "date: 2081/04/15 -> AD 2024-07-30", false},
		{"extra line", wantOutput + "\nextra", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := checkOutput(c.out)
			if (err == nil) != c.ok {
				t.Fatalf("checkOutput(%q) = %v, want ok=%v", c.out, err, c.ok)
			}
		})
	}
}

func TestOutsideRepo(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "sub")
	if err := os.MkdirAll(inside, 0o750); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := outsideRepo(root, root); err == nil {
		t.Error("the repository root passed as outside")
	}
	if err := outsideRepo(root, inside); err == nil {
		t.Error("a directory inside the repository passed as outside")
	}
	if err := outsideRepo(root, outside); err != nil {
		t.Errorf("a sibling directory failed: %v", err)
	}
}

func TestOutsideRepoDifferentVolume(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("volume names exist only on Windows")
	}
	if err := outsideRepo(`D:\checkout`, `C:\Users\runner\Temp\consumer`); err != nil {
		t.Errorf("a directory on another volume failed: %v", err)
	}
	if err := outsideRepo(`\\host\share\checkout`, `\\host\builds\consumer`); err != nil {
		t.Errorf("a directory on another UNC share failed: %v", err)
	}
	// The same volume spelled in another case is still the same volume.
	if err := outsideRepo(`E:\checkout`, `e:\checkout\consumer`); err == nil {
		t.Error("a directory inside the repository passed as outside when the drive letter case differs")
	}
	if err := outsideRepo(`\\host\share\checkout`, `\\HOST\SHARE\checkout\consumer`); err == nil {
		t.Error("a directory inside a UNC checkout passed as outside when the share case differs")
	}
}

func TestWriteModule(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "checkout")
	if err := writeModule(dir, root, modulePath); err != nil {
		t.Fatal(err)
	}
	mod, err := os.ReadFile(filepath.Join(dir, "go.mod")) //nolint:gosec // a file this test just wrote
	if err != nil {
		t.Fatal(err)
	}
	want := "replace " + modulePath + " => " + filepath.ToSlash(root)
	if !strings.Contains(string(mod), want) {
		t.Errorf("go.mod lacks %q:\n%s", want, mod)
	}
	if !strings.Contains(string(mod), "require "+modulePath+" v0.0.0") {
		t.Errorf("go.mod lacks the require line:\n%s", mod)
	}
	src, err := os.ReadFile(filepath.Join(dir, "main.go")) //nolint:gosec // a file this test just wrote
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(src), "IMPORT_PATH") {
		t.Error("main.go still has the import placeholder")
	}
	if !strings.Contains(string(src), `"`+modulePath+`"`) {
		t.Errorf("main.go does not import %s:\n%s", modulePath, src)
	}
}

func TestWritePublishedModule(t *testing.T) {
	dir := t.TempDir()
	if err := writePublishedModule(dir, "v0.1.0-beta.1"); err != nil {
		t.Fatal(err)
	}
	mod, err := os.ReadFile(filepath.Join(dir, "go.mod")) //nolint:gosec // a file this test just wrote
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mod), "require "+modulePath+" v0.1.0-beta.1\n") {
		t.Errorf("go.mod lacks the released version:\n%s", mod)
	}
	// A replace directive would point the check back at a checkout and prove nothing.
	if strings.Contains(string(mod), "replace") {
		t.Errorf("go.mod of a published check has a replace directive:\n%s", mod)
	}
}

func TestPublishedEnvUsesOnlyThePublicProxy(t *testing.T) {
	env := strings.Join(publishedEnv, "\n")
	for _, want := range []string{"GOPROXY=https://proxy.golang.org", "GOSUMDB=sum.golang.org", "GOPRIVATE=", "GONOSUMDB="} {
		if !strings.Contains(env, want+"\n") && !strings.HasSuffix(env, want) {
			t.Errorf("publishedEnv lacks %q", want)
		}
	}
	if strings.Contains(env, "direct") {
		t.Error("publishedEnv allows a direct fetch, which would hide a version the proxy lacks")
	}
}

func TestRunPublishedRejectsVersionWithoutV(t *testing.T) {
	if err := runPublished("0.1.0", 0); err == nil {
		t.Error("a version without the v prefix was accepted")
	}
}

func TestWriteModuleBreaksImport(t *testing.T) {
	dir := t.TempDir()
	if err := writeModule(dir, t.TempDir(), brokenPath); err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(filepath.Join(dir, "main.go")) //nolint:gosec // a file this test just wrote
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), `"`+brokenPath+`"`) {
		t.Errorf("main.go does not import the broken path:\n%s", src)
	}
}

func TestLatestVersion(t *testing.T) {
	cases := []struct {
		name string
		body string
		ok   bool
	}{
		{"pre-release while none is stable", `{"Version":"v0.1.0-beta.1","Time":"2026-10-21T15:49:47Z"}`, true},
		{"stable", `{"Version":"v0.2.0"}`, true},
		{"no version field", `{"Time":"2026-10-21T15:49:47Z"}`, false},
		{"not JSON", `<html>not found</html>`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/"+modulePath+"/@latest" {
					http.NotFound(w, r)
					return
				}
				fmt.Fprint(w, c.body) //nolint:errcheck // the test server writes a fixed body
			}))
			defer srv.Close()
			got, err := latestVersion(srv.URL)
			if (err == nil) != c.ok {
				t.Fatalf("latestVersion() = %q, %v, want ok=%v", got, err, c.ok)
			}
			if c.ok && !strings.HasPrefix(got, "v") {
				t.Errorf("latestVersion() = %q, want a version with the v prefix", got)
			}
		})
	}
}

func TestResolveVersion(t *testing.T) {
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		fmt.Fprint(w, `{"Version":"v0.1.0-beta.1"}`) //nolint:errcheck // the test server writes a fixed body
	}))
	defer srv.Close()

	got, err := resolveVersion(srv.URL, "latest")
	if err != nil || got != "v0.1.0-beta.1" {
		t.Fatalf("resolveVersion(latest) = %q, %v, want v0.1.0-beta.1", got, err)
	}
	if n := requests.Load(); n != 1 {
		t.Errorf("latest made %d requests, want 1", n)
	}
	got, err = resolveVersion(srv.URL, "v0.2.0")
	if err != nil || got != "v0.2.0" {
		t.Fatalf("resolveVersion(v0.2.0) = %q, %v, want v0.2.0", got, err)
	}
	if n := requests.Load(); n != 1 {
		t.Errorf("an explicit version made %d requests, want none", n-1)
	}
	if _, err := resolveVersion(srv.URL, "0.2.0"); err == nil {
		t.Error("a version without the v prefix was accepted")
	}
}

// proxyLog records the method and path of every request a fake proxy receives.
type proxyLog struct {
	mu      sync.Mutex
	entries []string
}

func (l *proxyLog) add(r *http.Request) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, r.Method+" "+r.URL.Path)
}

func (l *proxyLog) all() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.entries...)
}

func TestCheckMetadata(t *testing.T) {
	const version = "v0.1.0-beta.1"
	cases := []struct {
		name     string
		listCode int
		list     string
		infoCode int
		zipCode  int
		want     string
	}{
		{"listed and served", http.StatusOK, version + "\n", http.StatusOK, http.StatusOK, ""},
		{"not listed", http.StatusOK, "v0.1.0\n", http.StatusOK, http.StatusOK, "does not list " + version},
		{"list missing", http.StatusNotFound, "", http.StatusOK, http.StatusOK, "does not list " + version},
		{"info missing", http.StatusOK, version + "\n", http.StatusNotFound, http.StatusOK, "answered 404"},
		{"zip missing", http.StatusOK, version + "\n", http.StatusOK, http.StatusNotFound, "answered 404"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			log := &proxyLog{}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.add(r)
				switch r.URL.Path {
				case "/" + modulePath + "/@v/list":
					w.WriteHeader(c.listCode)
					fmt.Fprint(w, c.list) //nolint:errcheck // the test server writes a fixed body
				case "/" + modulePath + "/@v/" + version + ".info":
					w.WriteHeader(c.infoCode)
				case "/" + modulePath + "/@v/" + version + ".zip":
					w.WriteHeader(c.zipCode)
				default:
					http.NotFound(w, r)
				}
			}))
			defer srv.Close()

			err := checkMetadata(srv.URL, version)
			if c.want == "" {
				if err != nil {
					t.Fatalf("checkMetadata() = %v, want nil", err)
				}
				want := []string{
					"GET /" + modulePath + "/@v/list",
					"HEAD /" + modulePath + "/@v/" + version + ".info",
					"HEAD /" + modulePath + "/@v/" + version + ".zip",
				}
				if got := log.all(); !reflect.DeepEqual(got, want) {
					t.Errorf("requests = %v, want %v (the module files must be HEAD, never downloaded)", got, want)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("checkMetadata() = %v, want an error containing %q", err, c.want)
			}
		})
	}
}
