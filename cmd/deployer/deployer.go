package deployer

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dave/jsgo/cmd/config"
)

func Start(config *config.Config) error {

	// create a temp dir
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	fpath := filepath.Join(dir, "out.wasm")

	args := []string{"build", "-o", fpath}

	// TODO: more args?

	path := "."
	if config.Path != "" {
		path = config.Path
	}
	args = append(args, path)

	cmd := exec.Command(config.Command, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOARCH=wasm")
	cmd.Env = append(cmd.Env, "GOOS=js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	if len(output) > 0 {
		return fmt.Errorf("%s", string(output))
	}

	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return err
	}
	sha := sha1.New()
	buf := bytes.NewBuffer(b)
	if _, err := io.Copy(sha, buf); err != nil {
		return err
	}
	fmt.Printf("%d: %x", len(b), sha.Sum(nil))

	return nil
}
