package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

const version = "0.4.0"
const buildDir = "tmp"
const releaseDir = "release"

type target struct {
	os      string
	arch    string
	ext     string
	packExt string
}

// go tool dist list
var targets = []target{
	// darwin
	{"darwin", "amd64", "", ".tar.gz"},
	{"darwin", "arm64", "", ".tar.gz"},
	// linux
	{"linux", "386", "", ".tar.gz"},
	{"linux", "amd64", "", ".tar.gz"},
	{"linux", "arm", "", ".tar.gz"},
	{"linux", "arm64", "", ".tar.gz"},
	// windows
	{"windows", "386", ".exe", ".zip"},
	{"windows", "amd64", ".exe", ".zip"},
	{"windows", "arm", ".exe", ".zip"},
	{"windows", "arm64", ".exe", ".zip"},
}

func main() {
	os.RemoveAll(releaseDir)
	runtime.Assert(os.MkdirAll(releaseDir, 0755))
	bindata()
	for _, target := range targets {
		logging.Info("build target %s/%s...", target.os, target.arch)
		build(target)
	}
}

func bindata() {
	cmd := exec.Command("go", "run", "contrib/bindata/main.go",
		"-pkg", "shell",
		"-o", "code/client/tunnel/shell/assets.go",
		"-prefix", "html/shell",
		"html/shell/...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	runtime.Assert(cmd.Run())

	cmd = exec.Command("go", "run", "contrib/bindata/main.go",
		"-pkg", "dashboard",
		"-o", "code/client/dashboard/assets.go",
		"-prefix", "html/dashboard",
		"html/dashboard/...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	runtime.Assert(cmd.Run())
}

func build(t target) {
	os.RemoveAll(buildDir)
	runtime.Assert(os.MkdirAll(buildDir, 0755))

	err := filepath.Walk("conf", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		path = strings.TrimPrefix(path, "conf")
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(buildDir, path), 0755)
		}
		return copyFile("conf"+path, filepath.Join(buildDir, path))
	})
	runtime.Assert(err)
	err = copyFile("CHANGELOG.md", path.Join(buildDir, "CHANGELOG.md"))
	runtime.Assert(err)

	ldflags := "-X 'main._GIT_HASH=" + gitHash() + "' " +
		"-X 'main._GIT_REVERSION=" + gitReversion() + "' " +
		"-X 'main._BUILD_TIME=" + buildTime() + "' " +
		"-X 'main._VERSION=" + version + "'"

	logging.Info("build server...")
	cmd := exec.Command("go", "build", "-o", path.Join(buildDir, "np-svr"+t.ext),
		"-ldflags", ldflags,
		path.Join("code", "server", "main.go"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		fmt.Sprintf("GOOS=%s", t.os),
		fmt.Sprintf("GOARCH=%s", t.arch))
	runtime.Assert(cmd.Run())

	logging.Info("build client...")
	cmd = exec.Command("go", "build", "-o", path.Join(buildDir, "np-cli"+t.ext),
		"-ldflags", ldflags,
		path.Join("code", "client", "main.go"),
		path.Join("code", "client", "connect.go"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		fmt.Sprintf("GOOS=%s", t.os),
		fmt.Sprintf("GOARCH=%s", t.arch))
	runtime.Assert(cmd.Run())

	logging.Info("packing...")
	pack(buildDir, path.Join(releaseDir, "natpass_"+version+"_"+t.os+"_"+t.arch+t.packExt))
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	fi, err := srcFile.Stat()
	if err != nil {
		return err
	}
	err = dstFile.Chmod(fi.Mode())
	if err != nil {
		return err
	}
	_, err = io.Copy(dstFile, srcFile)
	return err
}

func gitHash() string {
	var buf bytes.Buffer
	cmd := exec.Command("git", "log", "-n1", "--pretty=format:%h")
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	runtime.Assert(cmd.Run())
	return buf.String()
}

func gitReversion() string {
	var buf bytes.Buffer
	cmd := exec.Command("git", "log", "--oneline")
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	runtime.Assert(cmd.Run())
	cnt := len(strings.Split(buf.String(), "\n"))
	return fmt.Sprintf("%d", cnt)
}

func buildTime() string {
	return time.Now().Format(time.RFC3339)
}

func pack(src, dst string) {
	switch {
	case strings.HasSuffix(dst, ".tar.gz"):
		packTarGZ(src, dst)
	case strings.HasSuffix(dst, ".zip"):
		packZip(src, dst)
	}
}

func packTarGZ(src, dst string) {
	f, err := os.Create(dst)
	runtime.Assert(err)
	defer f.Close()
	gw := gzip.NewWriter(f)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	src, err = filepath.Abs(src)
	runtime.Assert(err)
	err = filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == src {
			return nil
		}
		dst := strings.TrimPrefix(path, src)
		dst = filepath.Join("natpass_"+version, dst)
		hdr, err := tar.FileInfoHeader(info, dst)
		if err != nil {
			return err
		}
		hdr.Name = dst
		err = tw.WriteHeader(hdr)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		f, err = os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})
	runtime.Assert(err)
}

func packZip(src, dst string) {
	f, err := os.Create(dst)
	runtime.Assert(err)
	defer f.Close()
	zw := zip.NewWriter(f)
	defer zw.Close()
	src, err = filepath.Abs(src)
	runtime.Assert(err)
	err = filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == src {
			return nil
		}
		dst := strings.TrimPrefix(path, src)
		dst = filepath.Join("natpass_"+version, dst)
		hdr, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		hdr.Name = dst
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		f, err = os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	})
	runtime.Assert(err)
}
