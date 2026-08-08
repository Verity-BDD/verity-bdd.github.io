package main

import (
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestRenderAliasPreservesPublicStructContract(t *testing.T) {
	model := mustLoadFixtureModel(t)
	pkg := model.packageByImportPath("example.com/aliasmod/facade")
	shape, err := model.renderAlias(pkg, "Item", pkg.aliases["Item"])
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"// Name is the externally visible name and costs $5.",
		"Name string `json:\"name\"`",
		"// State is the current lifecycle state.",
		"State State `json:\"state\"`",
		"// Embedded contributes its exported fields.",
		"Embedded",
	} {
		if !strings.Contains(shape, want) {
			t.Errorf("shape missing %q:\n%s", want, shape)
		}
	}
	if strings.Contains(shape, "Embedded Embedded") {
		t.Errorf("embedded field lost anonymous syntax:\n%s", shape)
	}
	qualifiers := regexp.MustCompile(`shared_[0-9a-f]{8}\.Token`).FindAllString(shape, -1)
	if len(qualifiers) != 2 || qualifiers[0] == qualifiers[1] {
		t.Errorf("same-name packages need distinct stable qualifiers, got %v:\n%s", qualifiers, shape)
	}
}

func TestInternalSelectorRewriteUsesIdentifierBoundariesAndFailsClosed(t *testing.T) {
	model := mustLoadFixtureModel(t)
	pkg := model.packageByImportPath("example.com/aliasmod/facade")
	_, err := model.rewriteInternalSelectors("impl.Item impl.ItemSuffix", pkg)
	if err == nil || !strings.Contains(err.Error(), ".ItemSuffix") {
		t.Fatalf("expected the complete adjacent identifier to fail closed, got %v", err)
	}
}

func TestRenderAliasPreservesInheritedInterfaceMethodComment(t *testing.T) {
	model := mustLoadFixtureModel(t)
	pkg := model.packageByImportPath("example.com/aliasmod/facade")
	shape, err := model.renderAlias(pkg, "Composite", pkg.aliases["Composite"])
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"// Ping verifies that the component is responsive.", "Ping() error"} {
		if !strings.Contains(shape, want) {
			t.Errorf("interface shape missing %q:\n%s", want, shape)
		}
	}
}

func TestRealGomarkdocGenerationPreservesContractAndPolicy(t *testing.T) {
	gomarkdoc := requireGomarkdoc(t)
	libDir, sha := cleanFixtureRepository(t)
	outDir := filepath.Join(t.TempDir(), "api")
	if err := generateWithInventory(libDir, outDir, gomarkdoc, sha, []string{"facade"}); err != nil {
		t.Fatal(err)
	}
	gotBytes, err := os.ReadFile(filepath.Join(outDir, "facade.md"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(gotBytes)
	for _, want := range []string{
		"Library commit: `" + sha + "`",
		"Name string `json:\"name\"`",
		"Name is the externally visible name and costs $5.",
		"const PriceLabel = \"costs $5\"",
		"const TypedPriceLabel string = \"costs $5\"",
		"const TypedCount int = 5",
		"const TypedSwitch bool = true",
		"const RenamedPriceLabel = \"renamed costs $5\"",
		"const LiteralPriceLabel = \"literal costs $5\"",
		"GroupedLabel = \"costs $5\"",
		"GroupedCount = 5",
		"GroupedSwitch = true",
		"PriceLabel is a standalone forwarded constant whose value contains $5.",
		"Embedded contributes its exported fields.",
		"// Ping verifies that the component is responsive.",
		"Rename changes the item's public name.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("real gomarkdoc output missing %q:\n%s", want, got)
		}
	}
	for _, forbidden := range []string{"hidden.", "impl.", "/internal/", "secret string"} {
		if strings.Contains(got, forbidden) {
			t.Errorf("real gomarkdoc output contains forbidden %q:\n%s", forbidden, got)
		}
	}
	assertGeneratedConstantFencesAreGo(t, got)
	assertMarkdownInventory(t, outDir, []string{"facade.md"})
}

func assertGeneratedConstantFencesAreGo(t *testing.T, markdown string) {
	t.Helper()
	fences := regexp.MustCompile("(?s)```go\\n(.*?)\\n```").FindAllStringSubmatch(markdown, -1)
	checked := 0
	for _, fence := range fences {
		declarations := fence[1]
		if !regexp.MustCompile(`(?m)^const(?:[ 	(]|$)`).MatchString(declarations) {
			continue
		}
		checked++
		source := []byte("package generated\n\n" + declarations + "\n")
		if _, err := parser.ParseFile(token.NewFileSet(), "generated.go", source, parser.AllErrors); err != nil {
			t.Errorf("generated constant fence is not valid Go: %v\n%s", err, declarations)
			continue
		}
		if _, err := format.Source(source); err != nil {
			t.Errorf("gofmt rejected generated constant fence: %v\n%s", err, declarations)
		}
	}
	if checked == 0 {
		t.Fatal("real gomarkdoc output contained no constant declaration fences")
	}
}

func TestReplaceOutputTreeIgnoresBackupCleanupFailureAfterInstall(t *testing.T) {
	root := t.TempDir()
	stage := filepath.Join(root, "stage")
	destination := filepath.Join(root, "api")
	if err := os.MkdirAll(stage, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stage, "new.md"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destination, "old.md"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := replaceOutputTreeWithRemove(stage, destination, func(string) error { return fs.ErrPermission }); err != nil {
		t.Fatalf("cleanup failure after installation must be best-effort, got %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(destination, "new.md")); err != nil || string(got) != "new" {
		t.Fatalf("new output was not installed: content=%q err=%v", got, err)
	}
	if _, err := os.Stat(stage + "-last-good"); err != nil {
		t.Fatalf("failed cleanup should leave the backup available: %v", err)
	}
}

func TestUnsupportedInternalTypeFailsClosedAndPreservesLastGoodOutput(t *testing.T) {
	gomarkdoc := requireGomarkdoc(t)
	libDir, _ := copyFixture(t)
	hiddenPath := filepath.Join(libDir, "internal", "hidden", "hidden.go")
	body, err := os.ReadFile(hiddenPath)
	if err != nil {
		t.Fatal(err)
	}
	mutated := strings.Replace(string(body), "type Item struct {", "type PrivateShape struct{}\n\ntype Item struct {\n\tLeaked PrivateShape", 1)
	if mutated == string(body) {
		t.Fatal("fixture mutation was a no-op")
	}
	if err := os.WriteFile(hiddenPath, []byte(mutated), 0o644); err != nil {
		t.Fatal(err)
	}
	sha := commitFixture(t, libDir)
	outDir := filepath.Join(t.TempDir(), "api")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(outDir, "last-good.md")
	if err := os.WriteFile(sentinel, []byte("last good"), 0o644); err != nil {
		t.Fatal(err)
	}
	err = generateWithInventory(libDir, outDir, gomarkdoc, sha, []string{"facade"})
	if err == nil || !strings.Contains(err.Error(), "unsupported internal type") {
		t.Fatalf("expected unsupported internal type failure, got %v", err)
	}
	got, readErr := os.ReadFile(sentinel)
	if readErr != nil || string(got) != "last good" {
		t.Fatalf("last-good output was not preserved: content=%q err=%v", got, readErr)
	}
	assertMarkdownInventory(t, outDir, []string{"last-good.md"})
}

func TestMissingGomarkdocPreservesLastGoodOutput(t *testing.T) {
	libDir, sha := cleanFixtureRepository(t)
	outDir := filepath.Join(t.TempDir(), "api")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(outDir, "last-good.md")
	if err := os.WriteFile(sentinel, []byte("last good"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := generateWithInventory(libDir, outDir, filepath.Join(t.TempDir(), "missing-gomarkdoc"), sha, []string{"facade"})
	if err == nil {
		t.Fatal("expected missing gomarkdoc failure")
	}
	if got, readErr := os.ReadFile(sentinel); readErr != nil || string(got) != "last good" {
		t.Fatalf("last-good output was not preserved: content=%q err=%v", got, readErr)
	}
}

func TestGenerationRejectsFalseProvenanceAndDirtySource(t *testing.T) {
	gomarkdoc := requireGomarkdoc(t)
	libDir, sha := cleanFixtureRepository(t)
	outDir := filepath.Join(t.TempDir(), "api")
	for name, tc := range map[string]struct {
		supplied string
		dirty    bool
		want     string
	}{
		"uppercase SHA": {strings.ToUpper(sha), false, "lowercase 40-character"},
		"wrong SHA":     {strings.Repeat("0", 40), false, "does not match"},
		"dirty tree":    {sha, true, "not clean"},
	} {
		t.Run(name, func(t *testing.T) {
			if tc.dirty {
				if err := os.WriteFile(filepath.Join(libDir, "dirty.tmp"), []byte("dirty"), 0o644); err != nil {
					t.Fatal(err)
				}
				defer os.Remove(filepath.Join(libDir, "dirty.tmp"))
			}
			err := generateWithInventory(libDir, outDir, gomarkdoc, tc.supplied, []string{"facade"})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got %v", tc.want, err)
			}
		})
	}
}

func TestDiscoverPackagesRequiresExactPublicInventory(t *testing.T) {
	got, err := discoverPackages(filepath.Join("testdata", "aliasmod"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].RelativePath != "facade" {
		t.Fatalf("got %#v, want only facade", got)
	}
}

func TestWriteOutputsRejectsInventoryDrift(t *testing.T) {
	for name, tc := range map[string]struct {
		actual   []packageInfo
		expected []string
		want     string
	}{
		"missing":    {[]packageInfo{{RelativePath: "facade"}}, []string{"facade", "missing"}, "missing expected"},
		"unexpected": {[]packageInfo{{RelativePath: "facade"}, {RelativePath: "extra"}}, []string{"facade"}, "unexpected public"},
	} {
		t.Run(name, func(t *testing.T) {
			err := validateInventory(tc.actual, tc.expected)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q inventory error, got %v", tc.want, err)
			}
		})
	}
}

func mustLoadFixtureModel(t *testing.T) *apiModel {
	t.Helper()
	model, err := loadModel(filepath.Join("testdata", "aliasmod"))
	if err != nil {
		t.Fatal(err)
	}
	return model
}

func requireGomarkdoc(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("gomarkdoc")
	if err != nil {
		t.Skip("gomarkdoc integration dependency is not installed")
	}
	return path
}

func cleanFixtureRepository(t *testing.T) (string, string) {
	t.Helper()
	dir, _ := copyFixture(t)
	return dir, commitFixture(t, dir)
}

func copyFixture(t *testing.T) (string, string) {
	t.Helper()
	src := filepath.Join("testdata", "aliasmod")
	dst := filepath.Join(t.TempDir(), "aliasmod")
	err := filepath.WalkDir(src, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, body, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return dst, ""
}

func commitFixture(t *testing.T, dir string) string {
	t.Helper()
	runGit(t, dir, "init", "-q")
	runGit(t, dir, "config", "user.email", "api-docs@example.invalid")
	runGit(t, dir, "config", "user.name", "API Docs Test")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-qm", "fixture")
	return strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func assertMarkdownInventory(t *testing.T, dir string, expected []string) {
	t.Helper()
	var actual []string
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			actual = append(actual, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(actual, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("output inventory got %v, want %v", actual, expected)
	}
}

func TestMain(m *testing.M) {
	os.Setenv("GODEBUG", "gotypesalias=1")
	os.Exit(m.Run())
}
