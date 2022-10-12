package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/google/go-github/github"
	"github.com/lwch/runtime"
	"github.com/tdewolff/minify/v2"
	htmlmini "github.com/tdewolff/minify/v2/html"
	"golang.org/x/oauth2"
)

func main() {
	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		fmt.Println("Missing GITHUB_TOKEN env")
		os.Exit(1)
	}
	repo, ok := os.LookupEnv("GITHUB_REPOSITORY")
	if !ok {
		fmt.Println("Missing GITHUB_REPOSITORY env")
		os.Exit(1)
	}
	version, ok := os.LookupEnv("BUILD_VERSION")
	if !ok {
		fmt.Println("Missing GITHUB_REF env")
		os.Exit(1)
	}

	tmp := strings.SplitN(repo, "/", 2)
	if len(tmp) != 2 {
		fmt.Println("Invalid repo")
		os.Exit(1)
	}

	log.Printf("repo=%s", repo)

	owner, repo := tmp[0], tmp[1]

	ac := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	})
	ocli := oauth2.NewClient(context.Background(), ac)
	gcli := github.NewClient(ocli)

	changelog := getChangeLog(version)

	log.Printf("create release version=%s", version)
	releaseID := createOrDrop(gcli, owner, repo, version, changelog)

	files, err := filepath.Glob(filepath.Join("release", "*"))
	runtime.Assert(err)
	for _, file := range files {
		fi, err := os.Stat(file)
		runtime.Assert(err)
		if fi.IsDir() {
			continue
		}
		upload(gcli, owner, repo, releaseID, file)
	}
}

func createOrDrop(cli *github.Client, owner, repo, version, body string) int64 {
	branch := version

	rel, rep, err := cli.Repositories.GetReleaseByTag(context.Background(), owner, repo, branch)
	if err == nil {
		defer rep.Body.Close()
		log.Println("old release found, delete...")
		r, _ := cli.Repositories.DeleteRelease(context.Background(), owner, repo, rel.GetID())
		if r != nil {
			defer r.Body.Close()
		}
	}

	log.Printf("create release %s...", version)
	var release github.RepositoryRelease
	release.TagName = &branch
	release.Name = &version
	release.Body = &body
	ret, rep, err := cli.Repositories.CreateRelease(
		context.Background(), owner, repo, &release)
	runtime.Assert(err)
	defer rep.Body.Close()
	return ret.GetID()
}

func upload(cli *github.Client, owner, repo string, id int64, dir string) {
	log.Printf("upload file %s...", dir)
	f, err := os.Open(dir)
	runtime.Assert(err)
	defer f.Close()
	var opt github.UploadOptions
	opt.Name = filepath.Base(dir)
	var rep *github.Response
	_, rep, err = cli.Repositories.UploadReleaseAsset(
		context.Background(), owner, repo, id, &opt, f)
	runtime.Assert(err)
	defer rep.Body.Close()
}

func getChangeLog(version string) string {
	data, err := os.ReadFile("CHANGELOG.md")
	runtime.Assert(err)

	root := markdown.Parse(data, parser.New())
	list := root.GetChildren()
	if len(list) > 0 {
		if !isChangeLog(list[0]) {
			panic("is not CHANGELOG.md")
		}
	}
	doc := new(ast.Document)
	render := func() string {
		data := markdown.Render(doc, html.NewRenderer(html.RendererOptions{}))
		m := minify.New()
		m.AddFunc("text/html", htmlmini.Minify)
		dt, err := m.Bytes("text/html", data)
		if err == nil {
			return string(dt)
		}
		return string(data)
	}
	var nodes []ast.Node
	var latest Version
	for _, node := range list[1:] {
		ver, ok := isVersion(node)
		if ok {
			if len(nodes) > 0 {
				doc.SetChildren(nodes)
				if latest.String() == version {
					return render()
				}
			}
			nodes = nodes[:0]
			latest = ver
			continue
		}
		nodes = append(nodes, node)
	}
	if len(nodes) > 0 {
		doc.SetChildren(nodes)
		return render()
	}

	return ""
}

func getContent(node ast.Node) string {
	contentToString := func(a, b []byte) string {
		if len(a) > 0 {
			return string(a)
		}
		if len(b) > 0 {
			return string(b)
		}
		return ""
	}
	if c := node.AsContainer(); c != nil {
		return contentToString(c.Literal, c.Content)
	}
	leaf := node.AsLeaf()
	return contentToString(leaf.Literal, leaf.Content)
}

func isChangeLog(node ast.Node) bool {
	if _, ok := node.(*ast.Heading); !ok {
		return false
	}
	list := node.GetChildren()
	if len(list) == 0 {
		return false
	}
	return getContent(list[0]) == "CHANGELOG"
}

func isVersion(node ast.Node) (Version, bool) {
	var ver Version
	if _, ok := node.(*ast.Heading); !ok {
		return ver, false
	}
	list := node.GetChildren()
	if len(list) == 0 {
		return ver, false
	}
	var err error
	ver, err = ParseVersion(getContent(list[0]))
	if err == nil {
		return ver, true
	}
	return ver, false
}

type Version struct {
	data [3]int
}

func ParseVersion(str string) (Version, error) {
	str = strings.TrimPrefix(str, "v")
	var ret Version
	tmp := strings.SplitN(str, ".", 3)
	if len(tmp) != 3 {
		return ret, errors.New("invalid version")
	}
	n, err := strconv.ParseInt(tmp[0], 10, 64)
	if err != nil {
		return ret, errors.New("invalid major version")
	}
	ret.data[0] = int(n)
	n, err = strconv.ParseInt(tmp[1], 10, 64)
	if err != nil {
		return ret, errors.New("invalid minor version")
	}
	ret.data[1] = int(n)
	n, err = strconv.ParseInt(tmp[2], 10, 64)
	if err != nil {
		return ret, errors.New("invalid patch version")
	}
	ret.data[2] = int(n)
	return ret, nil
}

func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.data[0], v.data[1], v.data[2])
}
