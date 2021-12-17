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

var version = "0.0.0"

const buildDir = "tmp"
const releaseDir = "release"

type target struct {
	os      string
	arch    string
	cc      string
	cxx     string
	ext     string
	packExt string
}

// go tool dist list
var targets = []target{
	// darwin
	{
		os:      "darwin",
		arch:    "amd64",
		packExt: ".tar.gz",
	},
	{
		os:      "darwin",
		arch:    "arm64",
		packExt: ".tar.gz",
	},
	// linux
	{
		os:      "linux",
		arch:    "386",
		packExt: ".tar.gz",
	},
	{
		os:      "linux",
		arch:    "amd64",
		packExt: ".tar.gz",
	},
	{
		os:      "linux",
		arch:    "arm",
		packExt: ".tar.gz",
	},
	{
		os:      "linux",
		arch:    "arm64",
		packExt: ".tar.gz",
	},
	// windows
	{
		os:   "windows",
		arch: "386",
		cc:   "i686-w64-mingw32-gcc", cxx: "i686-w64-mingw32-g++",
		ext: ".exe", packExt: ".zip",
	},
	{
		os:   "windows",
		arch: "amd64",
		cc:   "x86_64-w64-mingw32-gcc", cxx: "x86_64-w64-mingw32-g++",
		ext: ".exe", packExt: ".zip",
	},
	{
		os:   "windows",
		arch: "arm",
		cc:   "x86_64-w64-mingw32-gcc", cxx: "x86_64-w64-mingw32-g++",
		ext: ".exe", packExt: ".zip",
	},
	{
		os:   "windows",
		arch: "arm64",
		cc:   "x86_64-w64-mingw32-gcc", cxx: "x86_64-w64-mingw32-g++",
		ext: ".exe", packExt: ".zip",
	},
}

func main() {
	if v, ok := os.LookupEnv("BUILD_VERSION"); ok {
		version = v
	}

	logging.Info("go env...")
	cmd := exec.Command("go", "env")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	runtime.Assert(cmd.Run())

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
		"-o", "code/client/rule/shell/assets.go",
		"-prefix", "html/shell",
		"html/shell/...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	runtime.Assert(cmd.Run())

	cmd = exec.Command("go", "run", "contrib/bindata/main.go",
		"-pkg", "vnc",
		"-o", "code/client/rule/vnc/assets.go",
		"-prefix", "html/vnc",
		"html/vnc/...")
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
	err = copyFile("docs/startup.md", path.Join(buildDir, "startup.md"))
	runtime.Assert(err)

	ldflags := "-X 'main.gitHash=" + gitHash() + "' " +
		"-X 'main.gitReversion=" + gitReversion() + "' " +
		"-X 'main.buildTime=" + buildTime() + "' " +
		"-X 'main.version=" + version + "' "
	// "--extldflags '-static -fpic -lssp'"

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
	env := append(os.Environ(),
		fmt.Sprintf("GOOS=%s", t.os),
		fmt.Sprintf("GOARCH=%s", t.arch),
		fmt.Sprintf("CC=%s", t.cc),
		fmt.Sprintf("CXX=%s", t.cxx))
	args := []string{"build", "-o", path.Join(buildDir, "np-cli"+t.ext), "-ldflags", ldflags}
	if t.os == "windows" && !strings.Contains(t.arch, "arm") {
		args = append(args, "-tags", "vnc")
		env = append(env, "CGO_ENABLED=1")
	} else if t.os == "linux" && !strings.Contains(t.arch, "arm") {
		env = append(env, "CGO_ENABLED=1")
	}
	args = append(args,
		path.Join("code", "client", "main.go"))
	cmd = exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env
	runtime.Assert(cmd.Run())

	// if t.os == "linux" && !strings.Contains(t.arch, "arm") {
	// 	ldflags := "-X 'main.gitHash=" + gitHash() + "' " +
	// 		"-X 'main.gitReversion=" + gitReversion() + "' " +
	// 		"-X 'main.buildTime=" + buildTime() + "' " +
	// 		"-X 'main.version=" + version + "'"
	// 	logging.Info("build client.vnc...")
	// 	env := append(os.Environ(),
	// 		fmt.Sprintf("GOOS=%s", t.os),
	// 		fmt.Sprintf("GOARCH=%s", t.arch))
	// 	args := []string{"build", "-o", path.Join(buildDir, "np-cli.vnc"), "-ldflags", ldflags}
	// 	args = append(args, "-tags", "vnc")
	// 	env = append(env, "CGO_ENABLED=1")
	// 	args = append(args,
	// 		path.Join("code", "client", "main.go"))
	// 	cmd = exec.Command("go", args...)
	// 	cmd.Stdout = os.Stdout
	// 	cmd.Stderr = os.Stderr
	// 	cmd.Env = env
	// 	runtime.Assert(cmd.Run())
	// }

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
		if info.IsDir() {
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
