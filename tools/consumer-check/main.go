// Command consumer-check builds a throwaway module outside this repository that imports
// github.com/sanamhub/nepdate-go through a replace directive to this checkout. The module converts
// one BS date, reads one patro event and prints both, so the check sees the module the way a user
// does: through the public import paths, outside the main module.
//
//	go run ./tools/consumer-check                 temporary module, removed on success
//	go run ./tools/consumer-check -dir E:\temp\go-consumer
//	go run ./tools/consumer-check -negative        break the import path and require failure
//	go run ./tools/consumer-check -published v0.1.0-beta.1 -wait 30m
//
// The child runs with GOPROXY=off, so only the local replacement can satisfy the import. The
// negative control proves the check runs the program instead of trusting a cached result.
//
// -published checks a released version instead: no replace directive, the module comes from
// proxy.golang.org and its hash from sum.golang.org, the same path a user's go get takes. -wait
// retries until the proxy serves the version, because a fresh tag can take minutes to appear.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	modulePath  = "github.com/sanamhub/nepdate-go"
	brokenPath  = modulePath + "-missing"
	consumerDir = "nepdate-consumer-"
)

// The date and its first event are fixed data of the coverage range, so the expected text is part
// of the check. A data or conversion change that moves either line fails the check on purpose.
const (
	wantOutput = "date: 2081/04/15 -> AD 2024-07-30\n" +
		"event: Kheer Khane Din"
)

// consumerSource is the program the throwaway module runs. The import path is substituted so the
// negative control can break it.
const consumerSource = `package main

import (
	"fmt"
	"os"

	"IMPORT_PATH"
	"github.com/sanamhub/nepdate-go/patro"
)

func main() {
	d, err := nepdate.New(2081, nepdate.Shrawan, 15)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	y, m, day := d.Gregorian()
	fmt.Printf("date: %v -> AD %04d-%02d-%02d\n", d, y, int(m), day)

	i := patro.Of(d)
	if i.EventCount() == 0 {
		fmt.Fprintln(os.Stderr, "no patro event for", d)
		os.Exit(1)
	}
	fmt.Printf("event: %s\n", i.Event(0).Name(nepdate.English))
}
`

func main() {
	dir := flag.String("dir", "", "write the consumer module here instead of a temporary directory")
	negative := flag.Bool("negative", false, "break the import path and require the build to fail")
	published := flag.String("published", "", "check this released version from proxy.golang.org instead of the checkout")
	wait := flag.Duration("wait", 0, "with -published, keep retrying for this long while the proxy lacks the version")
	flag.Parse()

	var err error
	if *published != "" {
		err = runPublished(*published, *wait)
	} else {
		err = run(*dir, *negative)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "consumer-check:", err)
		os.Exit(1)
	}
}

func run(dir string, negative bool) error {
	root, err := moduleRoot()
	if err != nil {
		return err
	}
	temporary := dir == ""
	if temporary {
		dir, err = os.MkdirTemp("", consumerDir)
		if err != nil {
			return err
		}
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		if temporary && ok {
			_ = os.RemoveAll(dir)
		}
	}()

	if err := outsideRepo(root, dir); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	importPath := modulePath
	if negative {
		importPath = brokenPath
	}
	if err := writeModule(dir, root, importPath); err != nil {
		return err
	}
	fmt.Println("consumer-check: module", dir)
	fmt.Println("consumer-check: replace", modulePath, "=>", root)

	resolved, err := goListModule(dir)
	if err != nil {
		return err
	}
	if err := sameDir(resolved, root); err != nil {
		return fmt.Errorf("%s resolves to %q, not the checkout %q: %w", modulePath, resolved, root, err)
	}

	if negative {
		return checkNegative(dir, importPath)
	}
	return checkPositive(dir)
}

func checkPositive(dir string) error {
	out, err := runConsumer(dir)
	fmt.Print(out)
	if err != nil {
		return fmt.Errorf("the consumer failed: %w", err)
	}
	if err := checkOutput(out); err != nil {
		return err
	}
	fmt.Println("consumer-check: ok")
	return nil
}

// checkNegative requires the broken import to fail for the right reason: a build that succeeds
// means the check is not exercising the import, and an unrelated failure would pass by accident.
func checkNegative(dir, importPath string) error {
	fmt.Println("consumer-check: negative control: import", importPath)
	out, err := runConsumer(dir)
	fmt.Print(out)
	if err == nil {
		return errors.New("the broken import still built, so the consumer check proves nothing")
	}
	if !strings.Contains(out, importPath) {
		return fmt.Errorf("the run failed without naming %q", importPath)
	}
	fmt.Println("consumer-check: negative control ok: the broken import failed as expected")
	return nil
}

// runConsumer runs the module and returns its combined output. A non-zero exit is returned as an
// error after the output is printed, so the caller can show the compiler's message.
func runConsumer(dir string) (string, error) {
	cmd := goCommand(dir, "run", ".")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

// goListModule asks the go command which directory provides the replaced module.
func goListModule(dir string) (string, error) {
	cmd := goCommand(dir, "list", "-m", "-f", "{{.Dir}}", modulePath)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go list -m: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// goCommand builds a go invocation in dir with the environment that keeps the check local and
// reproducible: no workspace, no proxy and no toolchain download.
func goCommand(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("go", args...) //nolint:gosec // the toolchain and constant arguments
	cmd.Dir = dir
	cmd.Env = withEnv(os.Environ(), "GOWORK=off", "GOPROXY=off", "GOFLAGS=-mod=mod", "GOTOOLCHAIN=local")
	return cmd
}

// publishedEnv is the environment of a user's go get: the public proxy and checksum database,
// nothing private, no workspace. Only proxy.golang.org is allowed, so a direct fetch from GitHub
// cannot hide a version the proxy does not serve.
var publishedEnv = []string{
	"GOWORK=off", "GOPROXY=https://proxy.golang.org", "GOSUMDB=sum.golang.org",
	"GONOSUMDB=", "GOPRIVATE=", "GOFLAGS=-mod=mod", "GOTOOLCHAIN=local",
}

// publishedCommand builds a go invocation in dir with publishedEnv.
func publishedCommand(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("go", args...) //nolint:gosec // the toolchain; the version is one argument
	cmd.Dir = dir
	cmd.Env = withEnv(os.Environ(), publishedEnv...)
	return cmd
}

// runPublished builds the consumer against a released version. It retries the download until wait
// has passed, then checks the resolved version, runs the program and checks its output.
func runPublished(version string, wait time.Duration) error {
	if !strings.HasPrefix(version, "v") {
		return fmt.Errorf("version %q must start with v", version)
	}
	dir, err := os.MkdirTemp("", consumerDir)
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	if err := writePublishedModule(dir, version); err != nil {
		return err
	}
	fmt.Println("consumer-check: module", dir)
	fmt.Println("consumer-check: require", modulePath, version, "from proxy.golang.org")

	deadline := time.Now().Add(wait)
	for {
		out, err := publishedCommand(dir, "mod", "download", modulePath+"@"+version).CombinedOutput()
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("go mod download %s@%s: %w\n%s", modulePath, version, err, out)
		}
		fmt.Printf("consumer-check: the proxy does not serve %s yet, retrying in 30s\n", version)
		time.Sleep(30 * time.Second)
	}

	got, err := publishedCommand(dir, "list", "-m", "-f", "{{.Version}}", modulePath).Output()
	if err != nil {
		return fmt.Errorf("go list -m: %w", err)
	}
	if v := strings.TrimSpace(string(got)); v != version {
		return fmt.Errorf("%s resolved to %s, want %s", modulePath, v, version)
	}

	cmd := publishedCommand(dir, "run", ".")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	fmt.Print(buf.String())
	if err != nil {
		return fmt.Errorf("the consumer failed: %w", err)
	}
	if err := checkOutput(buf.String()); err != nil {
		return err
	}
	fmt.Println("consumer-check: ok:", modulePath, version, "from proxy.golang.org")
	return nil
}

// writePublishedModule writes a go.mod that requires the released version with no replace
// directive, and the same main.go as the local check.
func writePublishedModule(dir, version string) error {
	mod := "module consumercheck\n\ngo 1.24\n\nrequire " + modulePath + " " + version + "\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o600); err != nil {
		return err
	}
	src := strings.ReplaceAll(consumerSource, "IMPORT_PATH", modulePath)
	return os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o600)
}

// withEnv replaces the values of the named keys in base and appends the rest.
func withEnv(base []string, kv ...string) []string {
	env := make([]string, 0, len(base)+len(kv))
	keys := make(map[string]bool, len(kv))
	for _, pair := range kv {
		key, _, _ := strings.Cut(pair, "=")
		keys[strings.ToUpper(key)] = true
	}
	for _, pair := range base {
		key, _, _ := strings.Cut(pair, "=")
		if !keys[strings.ToUpper(key)] {
			env = append(env, pair)
		}
	}
	return append(env, kv...)
}

// writeModule writes go.mod and main.go into dir. The replace path uses forward slashes, which
// go.mod accepts on every platform.
func writeModule(dir, root, importPath string) error {
	mod := "module consumercheck\n\ngo 1.24\n\n" +
		"require " + modulePath + " v0.0.0\n\n" +
		"replace " + modulePath + " => " + filepath.ToSlash(root) + "\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o600); err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(dir, "go.sum")); err != nil && !os.IsNotExist(err) {
		return err
	}
	src := strings.ReplaceAll(consumerSource, "IMPORT_PATH", importPath)
	return os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o600)
}

// moduleRoot returns the directory of the main module, so the tool works from any directory
// inside the repository.
func moduleRoot() (string, error) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}") //nolint:gosec // the go toolchain, no variable arguments
	cmd.Dir = "."
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go list -m: %w (run this from inside the repository)", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// outsideRepo reports an error when dir is inside root. A sibling directory is outside, and so is
// any directory on another Windows volume: filepath.Rel cannot compare those, but the OS temp
// directory is often on C: while the checkout is on another drive.
func outsideRepo(root, dir string) error {
	if differentVolume(root, dir) {
		return nil
	}
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return err
	}
	if rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
		return fmt.Errorf("%s is inside the repository %s; the consumer module must be outside", dir, root)
	}
	return nil
}

// differentVolume reports whether two absolute paths name different Windows volumes. On a platform
// without volume names both are empty, so it returns false.
func differentVolume(a, b string) bool {
	va, vb := filepath.VolumeName(a), filepath.VolumeName(b)
	if va == "" || vb == "" {
		return false
	}
	return !strings.EqualFold(va, vb)
}

// sameDir reports whether a and b are the same directory, following the platform's case rules.
func sameDir(a, b string) error {
	fa, err := os.Stat(a)
	if err != nil {
		return err
	}
	fb, err := os.Stat(b)
	if err != nil {
		return err
	}
	if !os.SameFile(fa, fb) {
		return errors.New("different directories")
	}
	return nil
}

// checkOutput compares the consumer's text with the expected date and event.
func checkOutput(out string) error {
	got := strings.TrimSpace(out)
	if got != wantOutput {
		return fmt.Errorf("consumer output:\n%s\nwant:\n%s", got, wantOutput)
	}
	return nil
}
