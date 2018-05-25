package builder

import (
	"fmt"
	"go/build"
	"go/types"
	"testing"

	"github.com/dave/services/copier"

	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"
)

func TestAll(t *testing.T) {

	masterList := map[string]string{}

	gopath, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(gopath)
	goroot, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(goroot)
	if err := copier.Copy(filepath.Join(build.Default.GOROOT, "src"), filepath.Join(goroot, "src"), osfs.New("/"), osfs.New("/")); err != nil {
		t.Fatal(err)
	}
	if err := copier.Copy(filepath.Join(build.Default.GOPATH, "src/github.com/gopherjs/gopherjs/js"), filepath.Join(goroot, "src/github.com/gopherjs/gopherjs/js"), osfs.New("/"), osfs.New("/")); err != nil {
		t.Fatal(err)
	}
	if err := copier.Copy(filepath.Join(build.Default.GOPATH, "src/github.com/gopherjs/gopherjs/nosync"), filepath.Join(goroot, "src/github.com/gopherjs/gopherjs/nosync"), osfs.New("/"), osfs.New("/")); err != nil {
		t.Fatal(err)
	}

	goroot1 := memfs.New()
	if err := copier.Copy("/src", "/goroot/src", osfs.New(build.Default.GOROOT), goroot1); err != nil {
		t.Fatal(err)
	}
	if err := copier.Copy("/src/github.com/gopherjs/gopherjs/js", "/goroot/src/github.com/gopherjs/gopherjs/js", osfs.New(build.Default.GOPATH), goroot1); err != nil {
		t.Fatal(err)
	}
	if err := copier.Copy("/src/github.com/gopherjs/gopherjs/nosync", "/goroot/src/github.com/gopherjs/gopherjs/nosync", osfs.New(build.Default.GOPATH), goroot1); err != nil {
		t.Fatal(err)
	}

	listCmd := exec.Command("go", "list", "./...")
	listCmd.Env = []string{
		fmt.Sprintf("GOPATH=%s", gopath),
		fmt.Sprintf("GOROOT=%s", goroot),
	}
	listCmd.Dir = filepath.Join(goroot, "src")
	stdLibPackagesBytes, err := listCmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	stdLibPackages := strings.Split(strings.TrimSpace(string(stdLibPackagesBytes)), "\n")
	excluded := map[string]bool{
		"builtin":        true,
		"internal/cpu":   true,
		"net/http/pprof": true,
		"plugin":         true,
		"runtime/cgo":    true,
	}
	skip := false
	skipUntil := ""
	for _, p := range stdLibPackages {
		if p == skipUntil {
			skip = false
		}
		if skip {
			continue
		}
		if excluded[p] {
			continue
		}
		fmt.Println("Testing", p)
		if err := testPackage(p, goroot, gopath, goroot1, masterList); err != nil {
			t.Fatal(err)
		}
	}

}

func testPackage(path, goroot, gopath string, goroot1 billy.Filesystem, masterList map[string]string) error {
	os.RemoveAll(filepath.Join(goroot, "pkg"))
	os.RemoveAll(filepath.Join(gopath, "pkg"))
	outpath := filepath.Join(gopath, "pkg", "out.js")
	packagesFromCommand := map[string]PackageOutput{}
	buildCmd := exec.Command("gopherjs", "build", path, "-o", outpath)
	buildCmd.Env = []string{
		fmt.Sprintf("GOPATH=%s", gopath),
		fmt.Sprintf("GOROOT=%s", goroot),
	}
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return err
	}
	walkFunc := func(root string) func(path string, info os.FileInfo, err error) error {
		return func(fpath string, info os.FileInfo, err error) error {

			if info.IsDir() {
				return nil
			}

			rel, err := filepath.Rel(root, fpath)
			if err != nil {
				return err
			}

			rel = filepath.ToSlash(rel)

			// ignore everything in the src directory
			if strings.HasPrefix(rel, "src/") {
				return nil
			}

			// ignore everything that's not an archive
			if !strings.HasSuffix(rel, ".a") {
				return nil
			}

			// find the package path
			path := strings.TrimSuffix(rel, ".a")
			path = strings.TrimPrefix(path, "pkg/")
			path = strings.TrimPrefix(path, fmt.Sprintf("%s_%s_js/", build.Default.GOOS, build.Default.GOARCH))
			path = strings.TrimPrefix(path, fmt.Sprintf("%s_js/", build.Default.GOOS))

			a, err := readArchive(osfs.New("/"), fpath, path, map[string]*types.Package{})

			contents, hash, err := GetPackageCode(a, false, false)
			if err != nil {
				return err
			}
			packagesFromCommand[path] = PackageOutput{
				Path:     path,
				Hash:     hash,
				Contents: contents,
			}
			return nil
		}
	}
	filepath.Walk(gopath, walkFunc(gopath))
	filepath.Walk(goroot, walkFunc(goroot))

	gopath1 := memfs.New()
	temp := memfs.New()

	s := New(&Options{
		Root:      goroot1,
		Path:      gopath1,
		Temporary: temp,
	})
	_, a, err := s.BuildImportPath(path)
	if err != nil {
		return err
	}
	hasMain := false
	expectedPackages := len(s.Archives)
	if a.Name == "main" {
		hasMain = true
		expectedPackages = len(s.Archives) - 1
	}
	if expectedPackages != len(packagesFromCommand) {
		fmt.Println("From code:")
		for _, a := range s.Archives {
			fmt.Println(a.ImportPath)
		}
		fmt.Println("From command:")
		for _, p := range packagesFromCommand {
			fmt.Println(p.Path)
		}
		return fmt.Errorf("%d packages from command, %d from code", len(packagesFromCommand), len(s.Archives))
	}
	for _, a := range s.Archives {
		if hasMain && a.ImportPath == path {
			continue
		}
		contents, hash, err := GetPackageCode(a, false, false)
		if err != nil {
			return err
		}
		p1 := packagesFromCommand[a.ImportPath]
		if fmt.Sprintf("%x", hash) != fmt.Sprintf("%x", p1.Hash) {
			return fmt.Errorf("package %s different\nFROM COMMAND:\n%s\nFROM CODE:\n%s", a.ImportPath, string(contents), string(p1.Contents))
		}
		if h, ok := masterList[a.ImportPath]; ok {
			if fmt.Sprintf("%x", hash) != h {
				return fmt.Errorf("package %s different master list", a.ImportPath)
			}
		} else {
			masterList[a.ImportPath] = fmt.Sprintf("%x", hash)
		}
	}
	return nil
}

func TestUnvendorPath(t *testing.T) {
	cases := map[string]string{
		"bytes":                                      "bytes",
		"crypto/rc4":                                 "crypto/rc4",
		"crypto/rc4/vendor":                          "crypto/rc4/vendor",
		"cmd/vendor/github.com/google/pprof/profile": "github.com/google/pprof/profile",
		"testing/internal-vendor/testdeps":           "testing/internal-vendor/testdeps",
		"/testing/internal-vendor/testdeps":          "/testing/internal-vendor/testdeps",
		"/testing/vendor/testdeps":                   "testdeps",
	}
	for input, expected := range cases {
		output := UnvendorPath(input)
		if output != expected {
			t.Fatalf("output %s not %s for input %s", output, expected, input)
		}
	}
}
