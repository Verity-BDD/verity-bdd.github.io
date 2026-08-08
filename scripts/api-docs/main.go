package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

var expectedPackages = []string{
	".",
	"verity_abilities/api",
	"verity_abilities/take_notes",
	"verity_abilities/wait",
	"verity_answerable",
	"verity_expectations",
	"verity_expectations/ensure",
	"verity_reporting",
	"verity_reporting/allure_reporter",
	"verity_reporting/console_reporter",
}

var lowercaseSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)
var selectorPattern = regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\.([A-Za-z_][A-Za-z0-9_]*)\b`)

type packageInfo struct {
	ImportPath   string
	RelativePath string
	Dir          string
}

type packageModel struct {
	pkg     *packages.Package
	aliases map[string]*types.Alias
}

type publicRef struct {
	pkgPath string
	pkgName string
	name    string
}

type apiModel struct {
	packages       map[string]*packageModel
	publicByType   map[string]publicRef
	methodDocs     map[string]string
	objectDocs     map[token.Pos]string
	qualifierNames map[string]string
}

func main() {
	libDir := flag.String("lib-dir", "", "path to the checked-out library")
	outDir := flag.String("output-dir", "", "API Markdown output directory")
	gomarkdoc := flag.String("gomarkdoc", "gomarkdoc", "path to gomarkdoc v1.1.0")
	sha := flag.String("library-sha", "", "selected lowercase library commit")
	flag.Parse()
	if *libDir == "" || *outDir == "" || *sha == "" {
		fatal(errors.New("--lib-dir, --output-dir, and --library-sha are required"))
	}
	if err := generate(*libDir, *outDir, *gomarkdoc, *sha); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "api-docs:", err)
	os.Exit(1)
}

func generate(libDir, outDir, gomarkdoc, sha string) error {
	return generateWithInventory(libDir, outDir, gomarkdoc, sha, expectedPackages)
}

func generateWithInventory(libDir, outDir, gomarkdoc, sha string, expected []string) error {
	if err := verifyLibrarySource(libDir, sha); err != nil {
		return err
	}
	pkgs, err := discoverPackages(libDir)
	if err != nil {
		return err
	}
	if err := validateInventory(pkgs, expected); err != nil {
		return err
	}
	model, err := loadModel(libDir)
	if err != nil {
		return err
	}

	outDir, err = filepath.Abs(outDir)
	if err != nil {
		return err
	}
	parent := filepath.Dir(outDir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(parent, "."+filepath.Base(outDir)+"-stage-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)

	for _, pkg := range pkgs {
		args := []string{"./" + pkg.RelativePath}
		if pkg.RelativePath == "." {
			args = []string{"./"}
		}
		cmd := exec.Command(gomarkdoc, args...)
		cmd.Dir = libDir
		raw, err := cmd.Output()
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				return fmt.Errorf("gomarkdoc %s: %w\n%s", pkg.RelativePath, err, exitErr.Stderr)
			}
			return fmt.Errorf("gomarkdoc %s: %w", pkg.RelativePath, err)
		}
		pm := model.packageByImportPath(pkg.ImportPath)
		if pm == nil {
			return fmt.Errorf("loaded model missing expected package %s", pkg.ImportPath)
		}
		body, err := rewriteMarkdown(string(raw), pm, model, sha)
		if err != nil {
			return fmt.Errorf("rewrite %s: %w", pkg.RelativePath, err)
		}
		title, relOut := outputName(pkg.RelativePath)
		frontmatter := fmt.Sprintf("---\ntitle: %s\ndescription: API reference for %s (library %s)\n---\n\n", title, title, sha)
		path := filepath.Join(stage, filepath.FromSlash(relOut))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(frontmatter+body), 0o644); err != nil {
			return err
		}
	}
	if err := validateOutputTree(stage, pkgs, sha); err != nil {
		return err
	}
	if err := replaceOutputTree(stage, outDir); err != nil {
		return err
	}
	for _, pkg := range pkgs {
		_, relOut := outputName(pkg.RelativePath)
		fmt.Printf("generated %s from %s\n", relOut, sha)
	}
	return nil
}

func verifyLibrarySource(libDir, sha string) error {
	if !lowercaseSHA.MatchString(sha) {
		return errors.New("--library-sha must be a lowercase 40-character hexadecimal commit SHA")
	}
	head, err := gitOutput(libDir, "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("read library HEAD: %w", err)
	}
	if head != sha {
		return fmt.Errorf("--library-sha %s does not match library HEAD %s", sha, head)
	}
	status, err := gitOutput(libDir, "status", "--porcelain", "--untracked-files=normal")
	if err != nil {
		return fmt.Errorf("read library worktree status: %w", err)
	}
	if status != "" {
		return errors.New("library source worktree is not clean; refusing false provenance")
	}
	return nil
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func outputName(relative string) (title, output string) {
	if relative == "." {
		return "verity-bdd", "verity-bdd.md"
	}
	return relative, relative + ".md"
}

func validateOutputTree(root string, pkgs []packageInfo, sha string) error {
	expected := make(map[string]bool, len(pkgs))
	for _, pkg := range pkgs {
		_, rel := outputName(pkg.RelativePath)
		expected[rel] = true
	}
	seen := map[string]bool{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !expected[rel] {
			return fmt.Errorf("unexpected generated output %s", rel)
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if len(body) == 0 || !bytes.Contains(body, []byte("Library commit: `"+sha+"`")) {
			return fmt.Errorf("generated output %s is empty or missing provenance", rel)
		}
		if bytes.Contains(body, []byte("/internal/")) {
			return fmt.Errorf("generated output %s exposes an internal import path", rel)
		}
		seen[rel] = true
		return nil
	})
	if err != nil {
		return err
	}
	for rel := range expected {
		if !seen[rel] {
			return fmt.Errorf("missing generated output %s", rel)
		}
	}
	if len(seen) != len(expected) {
		return fmt.Errorf("generated %d outputs, expected %d", len(seen), len(expected))
	}
	return nil
}

func replaceOutputTree(stage, destination string) error {
	return replaceOutputTreeWithRemove(stage, destination, os.RemoveAll)
}

func replaceOutputTreeWithRemove(stage, destination string, removeBackup func(string) error) error {
	backup := stage + "-last-good"
	_, statErr := os.Stat(destination)
	hadDestination := statErr == nil
	if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return statErr
	}
	if hadDestination {
		if err := os.Rename(destination, backup); err != nil {
			return fmt.Errorf("preserve last-good output: %w", err)
		}
	}
	if err := os.Rename(stage, destination); err != nil {
		if hadDestination {
			if rollbackErr := os.Rename(backup, destination); rollbackErr != nil {
				return fmt.Errorf("install generated output: %w (rollback also failed: %v)", err, rollbackErr)
			}
		}
		return fmt.Errorf("install generated output: %w", err)
	}
	if hadDestination {
		_ = removeBackup(backup)
	}
	return nil
}

func discoverPackages(moduleDir string) ([]packageInfo, error) {
	abs, err := filepath.Abs(moduleDir)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = abs
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list: %w", err)
	}
	dec := json.NewDecoder(bytes.NewReader(out))
	var result []packageInfo
	for dec.More() {
		var listed struct {
			Dir        string
			ImportPath string
			Name       string
			GoFiles    []string
			Module     *struct{ Dir string }
		}
		if err := dec.Decode(&listed); err != nil {
			return nil, err
		}
		if listed.Module == nil || filepath.Clean(listed.Module.Dir) != filepath.Clean(abs) || len(listed.GoFiles) == 0 {
			continue
		}
		rel, err := filepath.Rel(abs, listed.Dir)
		if err != nil {
			return nil, err
		}
		rel = filepath.ToSlash(rel)
		parts := strings.Split(rel, "/")
		if contains(parts, "internal") || contains(parts, "examples") || listed.Name == "main" {
			continue
		}
		result = append(result, packageInfo{ImportPath: listed.ImportPath, RelativePath: rel, Dir: listed.Dir})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].RelativePath < result[j].RelativePath })
	return result, nil
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func validateInventory(actual []packageInfo, expected []string) error {
	actualSet := make(map[string]bool, len(actual))
	for _, pkg := range actual {
		actualSet[pkg.RelativePath] = true
	}
	expectedSet := make(map[string]bool, len(expected))
	for _, pkg := range expected {
		expectedSet[pkg] = true
	}
	var problems []string
	for _, pkg := range expected {
		if !actualSet[pkg] {
			problems = append(problems, "missing expected package "+pkg)
		}
	}
	for _, pkg := range actual {
		if !expectedSet[pkg.RelativePath] {
			problems = append(problems, "unexpected public package "+pkg.RelativePath)
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("public package inventory drift:\n  %s", strings.Join(problems, "\n  "))
	}
	return nil
}

func loadModel(moduleDir string) (*apiModel, error) {
	abs, err := filepath.Abs(moduleDir)
	if err != nil {
		return nil, err
	}
	cfg := &packages.Config{
		Dir: abs,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedDeps | packages.NeedTypes |
			packages.NeedTypesInfo | packages.NeedSyntax,
		Env: append(os.Environ(), "GODEBUG=gotypesalias=1"),
	}
	roots, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, err
	}
	if packages.PrintErrors(roots) > 0 {
		return nil, errors.New("package loading failed")
	}
	all := collectPackages(roots)
	model := &apiModel{
		packages:       map[string]*packageModel{},
		publicByType:   map[string]publicRef{},
		methodDocs:     map[string]string{},
		objectDocs:     map[token.Pos]string{},
		qualifierNames: collisionSafeQualifiers(all),
	}
	for path, pkg := range all {
		pm := &packageModel{pkg: pkg, aliases: map[string]*types.Alias{}}
		for _, name := range pkg.Types.Scope().Names() {
			obj, ok := pkg.Types.Scope().Lookup(name).(*types.TypeName)
			if !ok || !obj.Exported() || !obj.IsAlias() {
				continue
			}
			alias, ok := obj.Type().(*types.Alias)
			if ok {
				pm.aliases[name] = alias
			}
		}
		model.packages[path] = pm
	}
	for path, pm := range model.packages {
		if isInternalImportPath(path) || strings.Contains(path, "/examples/") {
			continue
		}
		for name, alias := range pm.aliases {
			if named, ok := types.Unalias(alias).(*types.Named); ok {
				model.publicByType[typeKey(named.Obj())] = publicRef{pkgPath: path, pkgName: pm.pkg.Name, name: name}
			}
		}
	}
	model.indexDocs(all)
	return model, nil
}

func collectPackages(roots []*packages.Package) map[string]*packages.Package {
	result := map[string]*packages.Package{}
	var visit func(*packages.Package)
	visit = func(pkg *packages.Package) {
		if pkg == nil || result[pkg.PkgPath] != nil {
			return
		}
		result[pkg.PkgPath] = pkg
		for _, dep := range pkg.Imports {
			visit(dep)
		}
	}
	for _, root := range roots {
		visit(root)
	}
	return result
}

func collisionSafeQualifiers(all map[string]*packages.Package) map[string]string {
	byName := map[string][]string{}
	for path, pkg := range all {
		if pkg.Types != nil {
			byName[pkg.Types.Name()] = append(byName[pkg.Types.Name()], path)
		}
	}
	result := map[string]string{}
	for name, paths := range byName {
		sort.Strings(paths)
		for _, path := range paths {
			qualifier := name
			if len(paths) > 1 {
				sum := sha256.Sum256([]byte(path))
				qualifier += "_" + hex.EncodeToString(sum[:4])
			}
			result[path] = qualifier
		}
	}
	return result
}

func (m *apiModel) indexDocs(all map[string]*packages.Package) {
	for _, pkg := range all {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(node ast.Node) bool {
				switch node := node.(type) {
				case *ast.FuncDecl:
					if node.Recv != nil && node.Doc != nil {
						if fn, ok := pkg.TypesInfo.Defs[node.Name].(*types.Func); ok {
							m.methodDocs[methodKey(fn)] = node.Doc.Text()
						}
					}
				case *ast.Field:
					doc := fieldDoc(node)
					if doc == "" {
						break
					}
					m.objectDocs[node.Pos()] = doc
					for _, name := range node.Names {
						m.objectDocs[name.Pos()] = doc
					}
				}
				return true
			})
		}
	}
}

func fieldDoc(field *ast.Field) string {
	if field.Doc != nil {
		return field.Doc.Text()
	}
	if field.Comment != nil {
		return field.Comment.Text()
	}
	return ""
}

func (m *apiModel) packageByImportPath(path string) *packageModel { return m.packages[path] }

func typeKey(obj *types.TypeName) string {
	if obj == nil || obj.Pkg() == nil {
		return ""
	}
	return obj.Pkg().Path() + "." + obj.Name()
}

func methodKey(fn *types.Func) string {
	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.Recv() == nil {
		return ""
	}
	recv := sig.Recv().Type()
	if ptr, ok := recv.(*types.Pointer); ok {
		recv = ptr.Elem()
	}
	named, ok := recv.(*types.Named)
	if !ok {
		return ""
	}
	return typeKey(named.Obj()) + "." + fn.Name()
}

func rewriteMarkdown(markdown string, pkg *packageModel, model *apiModel, sha string) (string, error) {
	const marker = "<!-- Code generated by gomarkdoc. DO NOT EDIT -->"
	if !strings.Contains(markdown, marker) {
		return "", errors.New("gomarkdoc output missing generated marker")
	}
	markdown = strings.Replace(markdown, marker, "<!-- Code generated by gomarkdoc v1.1.0 and scripts/api-docs. DO NOT EDIT -->\n\n> Library commit: `"+sha+"`", 1)
	aliasNames := make([]string, 0, len(pkg.aliases))
	for name := range pkg.aliases {
		aliasNames = append(aliasNames, name)
	}
	sort.Strings(aliasNames)
	for _, name := range aliasNames {
		alias := pkg.aliases[name]
		shape, err := model.renderAlias(pkg, name, alias)
		if err != nil {
			return "", err
		}
		re := regexp.MustCompile(`type ` + regexp.QuoteMeta(name) + ` = [^\n]+`)
		if !re.MatchString(markdown) {
			return "", fmt.Errorf("gomarkdoc output missing alias declaration for %s", name)
		}
		markdown = re.ReplaceAllStringFunc(markdown, func(string) string { return shape })
		methods, err := model.renderMethods(pkg, name, alias)
		if err != nil {
			return "", err
		}
		if methods != "" {
			anchor := `<a name="` + name + `"></a>`
			start := strings.Index(markdown, anchor)
			if start < 0 {
				return "", fmt.Errorf("gomarkdoc output missing anchor for %s", name)
			}
			decl := strings.Index(markdown[start:], shape)
			if decl < 0 {
				return "", fmt.Errorf("rewritten declaration missing for %s", name)
			}
			fenceEnd := strings.Index(markdown[start+decl+len(shape):], "```")
			if fenceEnd < 0 {
				return "", fmt.Errorf("declaration fence missing for %s", name)
			}
			insertAt := start + decl + len(shape) + fenceEnd + len("```")
			markdown = markdown[:insertAt] + methods + markdown[insertAt:]
		}
	}
	var err error
	markdown, err = model.rewriteValues(markdown, pkg)
	if err != nil {
		return "", err
	}
	markdown, err = model.rewriteInternalSelectors(markdown, pkg)
	if err != nil {
		return "", err
	}
	if strings.Contains(markdown, "/internal/") {
		return "", errors.New("generated Markdown still exposes an internal import path")
	}
	return markdown, nil
}

func (m *apiModel) renderAlias(current *packageModel, name string, alias *types.Alias) (string, error) {
	named, ok := types.Unalias(alias).(*types.Named)
	if !ok {
		formatted, err := m.formatType(types.Unalias(alias), current)
		return "type " + name + " = " + formatted, err
	}
	switch underlying := named.Underlying().(type) {
	case *types.Struct:
		var b strings.Builder
		fmt.Fprintf(&b, "type %s struct {", name)
		for i := 0; i < underlying.NumFields(); i++ {
			field := underlying.Field(i)
			if !field.Exported() {
				continue
			}
			formatted, err := m.formatType(field.Type(), current)
			if err != nil {
				return "", fmt.Errorf("field %s.%s: %w", name, field.Name(), err)
			}
			writeIndentedComment(&b, m.objectDocs[field.Pos()])
			b.WriteString("\n    ")
			if !field.Anonymous() {
				b.WriteString(field.Name() + " ")
			}
			b.WriteString(formatted)
			if tag := underlying.Tag(i); tag != "" {
				b.WriteString(" `" + tag + "`")
			}
		}
		b.WriteString("\n}")
		return b.String(), nil
	case *types.Interface:
		underlying.Complete()
		var b strings.Builder
		fmt.Fprintf(&b, "type %s interface {", name)
		for i := 0; i < underlying.NumMethods(); i++ {
			method := underlying.Method(i)
			if !method.Exported() {
				continue
			}
			formatted, err := m.formatType(method.Type(), current)
			if err != nil {
				return "", fmt.Errorf("interface method %s.%s: %w", name, method.Name(), err)
			}
			writeIndentedComment(&b, m.objectDocs[method.Pos()])
			fmt.Fprintf(&b, "\n    %s%s", method.Name(), strings.TrimPrefix(formatted, "func"))
		}
		b.WriteString("\n}")
		return b.String(), nil
	default:
		formatted, err := m.formatType(underlying, current)
		return "type " + name + " " + formatted, err
	}
}

func writeIndentedComment(b *strings.Builder, doc string) {
	for _, line := range strings.Split(strings.TrimSpace(doc), "\n") {
		if line != "" {
			b.WriteString("\n    // " + line)
		}
	}
}

func (m *apiModel) formatType(typ types.Type, current *packageModel) (string, error) {
	qualifierPaths := map[string]string{}
	formatted := types.TypeString(typ, func(pkg *types.Package) string {
		if pkg == nil || pkg.Path() == current.pkg.PkgPath {
			return ""
		}
		qualifier := m.qualifierNames[pkg.Path()]
		if qualifier == "" {
			qualifier = pkg.Name()
		}
		qualifierPaths[qualifier] = pkg.Path()
		return qualifier
	})
	var formatErr error
	formatted = selectorPattern.ReplaceAllStringFunc(formatted, func(selector string) string {
		parts := selectorPattern.FindStringSubmatch(selector)
		path := qualifierPaths[parts[1]]
		if !isInternalImportPath(path) {
			return selector
		}
		ref, ok := m.publicByType[path+"."+parts[2]]
		if !ok {
			formatErr = fmt.Errorf("unsupported internal type %s.%s has no public alias", path, parts[2])
			return selector
		}
		return m.publicName(current, ref)
	})
	if formatErr != nil {
		return "", formatErr
	}
	return formatted, nil
}

func (m *apiModel) publicName(current *packageModel, ref publicRef) string {
	if ref.pkgPath == current.pkg.PkgPath {
		return ref.name
	}
	qualifier := m.qualifierNames[ref.pkgPath]
	if qualifier == "" {
		qualifier = ref.pkgName
	}
	return qualifier + "." + ref.name
}

func isInternalImportPath(path string) bool {
	for _, part := range strings.Split(path, "/") {
		if part == "internal" {
			return true
		}
	}
	return false
}

func (m *apiModel) renderMethods(current *packageModel, aliasName string, alias *types.Alias) (string, error) {
	named, ok := types.Unalias(alias).(*types.Named)
	if !ok {
		return "", nil
	}
	seen := map[string]bool{}
	var methods []*types.Func
	for _, typ := range []types.Type{named, types.NewPointer(named)} {
		set := types.NewMethodSet(typ)
		for i := 0; i < set.Len(); i++ {
			fn, ok := set.At(i).Obj().(*types.Func)
			if !ok || !fn.Exported() || seen[fn.Name()] {
				continue
			}
			seen[fn.Name()] = true
			methods = append(methods, fn)
		}
	}
	sort.Slice(methods, func(i, j int) bool { return methods[i].Name() < methods[j].Name() })
	var b strings.Builder
	for _, fn := range methods {
		sig := fn.Type().(*types.Signature)
		recvPrefix := ""
		if _, ok := sig.Recv().Type().(*types.Pointer); ok {
			recvPrefix = "*"
		}
		formatted, err := m.formatType(sig, current)
		if err != nil {
			return "", fmt.Errorf("method %s.%s: %w", aliasName, fn.Name(), err)
		}
		formatted = strings.TrimPrefix(formatted, "func")
		fmt.Fprintf(&b, "\n\n<a name=\"%s.%s\"></a>\n### func (%s%s) %s\n\n```go\nfunc (%s %s%s) %s%s\n```\n", aliasName, fn.Name(), recvPrefix, aliasName, fn.Name(), strings.ToLower(aliasName[:1]), recvPrefix, aliasName, fn.Name(), formatted)
		if doc := strings.TrimSpace(m.methodDocs[methodKey(fn)]); doc != "" {
			b.WriteString("\n" + doc + "\n")
		}
	}
	return b.String(), nil
}

func (m *apiModel) rewriteValues(markdown string, current *packageModel) (string, error) {
	for _, name := range current.pkg.Types.Scope().Names() {
		obj := current.pkg.Types.Scope().Lookup(name)
		switch obj := obj.(type) {
		case *types.Const:
			if !obj.Exported() {
				continue
			}
			line := name
			basic, isBasic := obj.Type().(*types.Basic)
			if !isBasic || basic.Info()&types.IsUntyped == 0 {
				typ, err := m.formatType(obj.Type(), current)
				if err != nil {
					return "", fmt.Errorf("constant %s: %w", name, err)
				}
				line += " " + typ
			}
			line += " = " + obj.Val().ExactString()
			re := regexp.MustCompile(`(?m)^([ 	]*)(const[ 	]+)?` + regexp.QuoteMeta(name) + `(?:[ 	]+[^=\n]+)?[ 	]*=[ 	]*[A-Za-z_][A-Za-z0-9_]*\.[A-Za-z_][A-Za-z0-9_]*[ 	]*$`)
			markdown = re.ReplaceAllStringFunc(markdown, func(declaration string) string {
				parts := re.FindStringSubmatch(declaration)
				if parts[2] != "" {
					return parts[1] + "const " + line
				}
				return parts[1] + line
			})
		case *types.Var:
			if !obj.Exported() || obj.IsField() {
				continue
			}
			typ, err := m.formatType(obj.Type(), current)
			if err != nil {
				return "", fmt.Errorf("variable %s: %w", name, err)
			}
			re := regexp.MustCompile(`var ` + regexp.QuoteMeta(name) + ` = [A-Za-z_][A-Za-z0-9_]*\.[^\n]+`)
			markdown = re.ReplaceAllStringFunc(markdown, func(string) string { return "var " + name + " " + typ })
		}
	}
	return markdown, nil
}

func (m *apiModel) rewriteInternalSelectors(markdown string, current *packageModel) (string, error) {
	aliases := map[string]string{}
	for _, file := range current.pkg.Syntax {
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if !isInternalImportPath(path) {
				continue
			}
			alias := ""
			if imp.Name != nil {
				alias = imp.Name.Name
			} else if target := m.packages[path]; target != nil {
				alias = target.pkg.Name
			}
			if alias != "" && alias != "_" && alias != "." {
				aliases[alias] = path
			}
		}
	}
	var rewriteErr error
	markdown = selectorPattern.ReplaceAllStringFunc(markdown, func(selector string) string {
		parts := selectorPattern.FindStringSubmatch(selector)
		path, internal := aliases[parts[1]]
		if !internal {
			return selector
		}
		ref, ok := m.publicByType[path+"."+parts[2]]
		if !ok {
			rewriteErr = fmt.Errorf("unsupported internal type %s.%s has no public alias", path, parts[2])
			return selector
		}
		return m.publicName(current, ref)
	})
	if rewriteErr != nil {
		return "", rewriteErr
	}
	return markdown, nil
}
