//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const oboeVersion = "1.6.1"

func main() {
	if err := clean(); err != nil {
		panic(err)
	}

	tmp, err := os.MkdirTemp("", "oboe-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmp)

	if err := download(tmp); err != nil {
		panic(err)
	}

	if err := prepareOboe(tmp); err != nil {
		panic(err)
	}

}

func clean() error {
	fmt.Println("Cleaning *.cpp and *.h files")
	if err := filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != "." {
			return filepath.SkipDir
		}
		base := filepath.Base(path)
		if base == "README-oboe.md" {
			return os.Remove(base)
		}
		if base == "LICENSE-oboe" {
			return os.Remove(base)
		}
		if !strings.HasPrefix(base, "oboe_") {
			return nil
		}
		if !strings.HasSuffix(base, ".cpp") && !strings.HasSuffix(base, ".h") {
			return nil
		}
		return os.Remove(base)
	}); err != nil {
		return err
	}
	return nil
}

func prepareOboe(tmp string) error {

	reInclude := regexp.MustCompile(`(?m)^#include\s+([<"])(.+)[>"]$`)

	fmt.Println("Copying *.cpp and *.h files")
	for _, dir := range []string{"src", "include"} {
		dir := dir
		indir := filepath.Join(tmp, "oboe-"+oboeVersion, dir)
		if err := filepath.Walk(indir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".cpp") && !strings.HasSuffix(path, ".h") {
				return nil
			}

			f, err := filepath.Rel(indir, path)
			if err != nil {
				return err
			}
			ext := filepath.Ext(f)
			curTs := strings.Split(f[:len(f)-len(ext)], string(filepath.Separator))
			outfn := "oboe_" + strings.Join(curTs, "_") + ext

			if _, err := os.Stat(outfn); err == nil {
				return fmt.Errorf("%s must not exist", outfn)
			}

			in, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Replace #include paths.
			in = reInclude.ReplaceAllFunc(in, func(inc []byte) []byte {
				m := reInclude.FindSubmatch(inc)
				f := string(m[2])

				searchDirs := []string{filepath.Dir(path)}
				if dir == "src" {
					searchDirs = append(searchDirs, filepath.Join(tmp, "oboe-"+oboeVersion, "src"))
				}
				searchDirs = append(searchDirs, filepath.Join(tmp, "oboe-"+oboeVersion, "include"))
				for _, searchDir := range searchDirs {
					path := filepath.Join(searchDir, f)
					e, err := exists(path)
					if err != nil {
						panic(err)
					}
					if !e {
						continue
					}

					f, err := filepath.Rel(filepath.Join(tmp, "oboe-"+oboeVersion), path)
					if err != nil {
						panic(err)
					}
					ext := filepath.Ext(f)
					ts := strings.Split(f[:len(f)-len(ext)], string(filepath.Separator))

					ts = ts[1:]
					newpath := "oboe_" + strings.Join(ts, "_") + ext
					return []byte(`#include "` + newpath + `"`)
				}
				return inc
			})

			out, err := os.Create(outfn)
			if err != nil {
				return err
			}
			defer out.Close()

			if _, err := io.Copy(out, bytes.NewReader(in)); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}
	}

	fmt.Println("Copying README.md and LICENSE")
	for _, f := range []string{"README.md", "LICENSE"} {
		infn := filepath.Join(tmp, "oboe-"+oboeVersion, f)

		ext := filepath.Ext(f)
		outfn := f[:len(f)-len(ext)] + "-oboe" + ext

		in, err := os.Open(infn)
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := os.Create(outfn)
		if err != nil {
			return err
		}
		defer out.Close()

		if _, err := io.Copy(out, in); err != nil {
			return err
		}
	}

	return nil
}

func download(tmp string) error {
	fn := "oboe-" + oboeVersion + ".tar.gz"
	if e, err := exists(fn); err != nil {
		return err
	} else if !e {
		url := "https://github.com/google/oboe/releases" + fn
		fmt.Fprintf(os.Stderr, "%s not found: please download it from %s\n", fn, url)
		return errors.New("oboe source do not exists")
	}

	fmt.Printf("Copying %s to %s\n", fn, filepath.Join(tmp, fn))
	in, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(filepath.Join(tmp, fn))
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	fmt.Printf("Extracting %s\n", fn)
	cmd := exec.Command("tar", "-xzf", fn)
	cmd.Stderr = os.Stderr
	cmd.Dir = tmp
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func exists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
