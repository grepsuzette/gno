package gno

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gnolang/gno/pkgs/std"
	"go.uber.org/multierr"
	"golang.org/x/tools/go/ast/astutil"
)

const (
	gnoRealmPkgsPrefixBefore = "gno.land/r/"
	gnoRealmPkgsPrefixAfter  = "github.com/gnolang/gno/examples/gno.land/r/"
	gnoPackagePrefixBefore   = "gno.land/p/"
	gnoPackagePrefixAfter    = "github.com/gnolang/gno/examples/gno.land/p/"
	gnoStdPkgBefore          = "std"
	gnoStdPkgAfter           = "github.com/gnolang/gno/stdlibs/stdshim"
)

var stdlibWhitelist = []string{
	// go
	"bufio",
	"bytes",
	"compress/gzip",
	"context",
	"crypto/md5",
	"crypto/sha1",
	"encoding/json",
	"encoding/base64",
	"encoding/binary",
	"encoding/xml",
	"flag",
	"fmt",
	"io",
	"io/util",
	"math",
	"math/big",
	"math/rand",
	"regexp",
	"sort",
	"strconv",
	"strings",
	"text/template",
	"unicode/utf8",

	// gno
	"std",
}

var importPrefixWhitelist = []string{
	"github.com/gnolang/gno/_test",
}

// TODO: func PrecompileFile: supports caching.
// TODO: func PrecompilePkg: supports directories.

func PrecompileAndCheckMempkg(mempkg *std.MemPackage) error {
	gofmt := "gofmt"

	tmpDir, err := ioutil.TempDir("", mempkg.Name)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir) // nolint: errcheck

	var errs error
	for _, mfile := range mempkg.Files {
		if !strings.HasSuffix(mfile.Name, ".gno") {
			continue // skip spurious file.
		}
		translated, err := Precompile(string(mfile.Body), "gno,tmp")
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		tmpFile := filepath.Join(tmpDir, mfile.Name)
		err = ioutil.WriteFile(tmpFile, []byte(translated), 0o644)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		err = PrecompileVerifyFile(tmpFile, gofmt)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
	}

	if errs != nil {
		return fmt.Errorf("precompile package: %w", errs)
	}
	return nil
}

func Precompile(source string, tags string) (string, error) {
	var out bytes.Buffer

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "tmp.gno", source, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("parse: %w", err)
	}

	transformed, err := precompileAST(fset, f)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	_, err = out.WriteString("// Code generated by github.com/gnolang/gno. DO NOT EDIT.\n\n//go:build " + tags + "\n// +build " + tags + "\n\n")
	if err != nil {
		return "", fmt.Errorf("write to buffer: %w", err)
	}
	err = format.Node(&out, fset, transformed)
	return out.String(), nil
}

// PrecompileVerifyFile tries to run `go fmt` against a precompiled .go file.
//
// This is fast and won't look the imports.
func PrecompileVerifyFile(path string, gofmtBinary string) error {
	// TODO: use cmd/parser instead of exec?

	args := []string{"-l", "-e", path}
	cmd := exec.Command(gofmtBinary, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, string(out))
		return fmt.Errorf("gofmt: %w", err)
	}
	return nil
}

// PrecompileBuildPackage tries to run `go build` against the precompiled .go files.
//
// This method is the most efficient to detect errors but requires that
// all the import are valid and available.
func PrecompileBuildPackage(fileOrPkg string, goBinary string) error {
	// TODO: use cmd/compile instead of exec?
	// TODO: find the nearest go.mod file, chdir in the same folder, rim prefix?
	// TODO: temporarily create an in-memory go.mod or disable go modules for gno?
	// TODO: ignore .go files that were not generated from gno?

	files := []string{}

	info, err := os.Stat(fileOrPkg)
	if err != nil {
		return fmt.Errorf("invalid file or package path: %w", err)
	}
	if !info.IsDir() {
		file := fileOrPkg
		files = append(files, file)
	} else {
		pkgDir := fileOrPkg
		goGlob := filepath.Join(pkgDir, "*.go")
		goMatches, err := filepath.Glob(goGlob)
		if err != nil {
			return fmt.Errorf("glob: %w", err)
		}
		for _, goMatch := range goMatches {
			switch {
			case strings.HasSuffix(goMatch, "."): // skip
			case strings.HasSuffix(goMatch, "_filetest.go"): // skip
			case strings.HasSuffix(goMatch, "_filetest.gno.gen.go"): // skip
			case strings.HasSuffix(goMatch, "_test.go"): // skip
			case strings.HasSuffix(goMatch, "_test.gno.gen.go"): // skip
			default:
				files = append(files, goMatch)
			}
		}
	}
	sort.Strings(files)
	args := append([]string{"build", "-v", "-tags=gno"}, files...)
	cmd := exec.Command(goBinary, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, string(out))
		return fmt.Errorf("std go compiler: %w", err)
	}

	return nil
}

func precompileAST(fset *token.FileSet, f *ast.File) (ast.Node, error) {
	var errs error

	imports := astutil.Imports(fset, f)

	// import whitelist
	for _, paragraph := range imports {
		for _, importSpec := range paragraph {
			importPath := strings.TrimPrefix(strings.TrimSuffix(importSpec.Path.Value, `"`), `"`)

			if strings.HasPrefix(importPath, gnoRealmPkgsPrefixBefore) {
				continue
			}

			if strings.HasPrefix(importPath, gnoPackagePrefixBefore) {
				continue
			}

			valid := false
			for _, whitelisted := range stdlibWhitelist {
				if importPath == whitelisted {
					valid = true
					break
				}
			}
			if valid {
				continue
			}

			for _, whitelisted := range importPrefixWhitelist {
				if strings.HasPrefix(importPath, whitelisted) {
					valid = true
					break
				}
			}
			if valid {
				continue
			}

			errs = multierr.Append(errs, fmt.Errorf("import %q is not in the whitelist", importPath))
		}
	}

	// rewrite imports
	for _, paragraph := range imports {
		for _, importSpec := range paragraph {
			importPath := strings.TrimPrefix(strings.TrimSuffix(importSpec.Path.Value, `"`), `"`)

			// std package
			if importPath == gnoStdPkgBefore {
				if !astutil.RewriteImport(fset, f, gnoStdPkgBefore, gnoStdPkgAfter) {
					errs = multierr.Append(errs, fmt.Errorf("failed to replace the %q package with %q", gnoStdPkgBefore, gnoStdPkgAfter))
				}
			}

			// p/pkg packages
			if strings.HasPrefix(importPath, gnoPackagePrefixBefore) {
				target := gnoPackagePrefixAfter + strings.TrimPrefix(importPath, gnoPackagePrefixBefore)

				if !astutil.RewriteImport(fset, f, importPath, target) {
					errs = multierr.Append(errs, fmt.Errorf("failed to replace the %q package with %q", importPath, target))
				}

			}

			// r/realm packages
			if strings.HasPrefix(importPath, gnoRealmPkgsPrefixBefore) {
				target := gnoRealmPkgsPrefixAfter + strings.TrimPrefix(importPath, gnoRealmPkgsPrefixBefore)

				if !astutil.RewriteImport(fset, f, importPath, target) {
					errs = multierr.Append(errs, fmt.Errorf("failed to replace the %q package with %q", importPath, target))
				}

			}
		}
	}

	// custom handler
	node := astutil.Apply(f,
		// pre
		func(c *astutil.Cursor) bool {
			// do things here
			return true
		},
		// post
		func(c *astutil.Cursor) bool {
			// and here
			return true
		},
	)

	return node, errs
}
