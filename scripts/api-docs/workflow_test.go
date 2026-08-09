package main

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
)

type documentedLibrary struct {
	Version string `json:"version"`
	SHA     string `json:"sha"`
}

func TestDocumentedLibraryManifestIsValid(t *testing.T) {
	library := loadDocumentedLibrary(t)
	if !regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`).MatchString(library.Version) {
		t.Errorf("invalid documented version %q", library.Version)
	}
	if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(library.SHA) {
		t.Errorf("invalid documented SHA %q", library.SHA)
	}
}

func TestInstallGuidePinsDocumentedRelease(t *testing.T) {
	library := loadDocumentedLibrary(t)
	body, err := os.ReadFile("../../src/content/docs/en/get_started/01_installation.md")
	if err != nil {
		t.Fatal(err)
	}
	guide := string(body)
	for _, required := range []string{
		"go get github.com/verity-bdd/verity-bdd@" + library.Version,
		library.SHA,
	} {
		if !strings.Contains(guide, required) {
			t.Errorf("installation guide missing %q", required)
		}
	}
}

func loadDocumentedLibrary(t *testing.T) documentedLibrary {
	t.Helper()
	body, err := os.ReadFile("../../documented-library.json")
	if err != nil {
		t.Fatal(err)
	}
	var library documentedLibrary
	if err := json.Unmarshal(body, &library); err != nil {
		t.Fatal(err)
	}
	return library
}

func TestDeployWorkflowUsesManifestRevisionAtomically(t *testing.T) {
	body, err := os.ReadFile("../../.github/workflows/deploy.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(body)
	for _, required := range []string{
		"documented-library.json",
		"id: documented-library",
		"steps.documented-library.outputs.sha",
		"steps.documented-library.outputs.version",
		"fetch-depth: 0",
		"fetch-tags: true",
		"Dispatch SHA $LIBRARY_SHA does not match documented SHA $DOCUMENTED_SHA",
	} {
		if !strings.Contains(workflow, required) {
			t.Errorf("workflow missing manifest contract %q", required)
		}
	}
}

func TestWritingFirstTestUsesCurrentRequestAPI(t *testing.T) {
	body, err := os.ReadFile("../../src/content/docs/en/get_started/02_writing-your-first-test.md")
	if err != nil {
		t.Fatal(err)
	}
	guide := string(body)
	for _, stale := range []string{"api.NewResponseHeader(", "api.NewJSONPath(", "api.NewRequestBuilder("} {
		if strings.Contains(guide, stale) {
			t.Errorf("guide contains removed call %q", stale)
		}
	}
	for _, current := range []string{"api.LastResponseHeader(", "api.LastResponseBodyAtJSONPath(", "api.RequestFor("} {
		if !strings.Contains(guide, current) {
			t.Errorf("guide missing current call %q", current)
		}
	}
	if !strings.Contains(guide, "if err := builder.WithJSONBody") || !strings.Contains(guide, "request, err := builder.Build()") {
		t.Error("guide must show WithJSONBody error handling before Build")
	}
}

func TestDeployWorkflowCompilesCheckedExamplesBeforeAstroBuild(t *testing.T) {
	body, err := os.ReadFile("../../.github/workflows/deploy.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(body)
	check := strings.Index(workflow, "scripts/check-docs-examples.sh")
	build := strings.Index(workflow, "npm run build")
	if check < 0 || build < 0 || check > build {
		t.Fatal("checked Go examples must run before the Astro build")
	}
}

func TestScreenplayGuideDocumentsActorsTerminalLifecycle(t *testing.T) {
	body, err := os.ReadFile("../../src/content/docs/en/core_concepts/1_screenplay.md")
	if err != nil {
		t.Fatal(err)
	}
	guide := string(body)
	for _, required := range []string{
		"test.Actors()", "fresh, non-nil snapshot", "case-sensitive lexical order",
		"modifying the returned slice", "ActorCalled` panics", "default-ability factories",
		"safe for concurrent use", "empty snapshot",
	} {
		if !strings.Contains(guide, required) {
			t.Errorf("screenplay guide missing actor lifecycle contract %q", required)
		}
	}
}

func TestBehaviorContractsAreDocumented(t *testing.T) {
	cases := map[string][]string{
		"../../src/content/docs/en/get_started/08_reusable_tasks.md": {
			"stops at the first child error", "nested child activities",
		},
		"../../src/content/docs/en/core_concepts/2_assertions.md": {
			"task itself is fail-fast", "poll-first", "pre-cancelled context", "CheckingEvery` requires a duration greater than zero", "panics at construction time",
		},
		"../../src/content/docs/en/get_started/10_abilities.md": {
			"WhoCan` appends", "first assignable match", "pointer and value forms", "caller or default-ability factory",
		},
		"../../src/content/docs/en/guides/11_notes.md": {
			"caller or a default-ability factory", "same instance to multiple actors",
		},
		"../../src/content/docs/en/core_concepts/1_screenplay.md": {
			"nil callback", "panics only when the interaction executes",
		},
		"../../src/content/docs/en/guides/21_attachments.md": {
			"NewActivityTracker", "Finish(err, attachments...)", "ordinary `Actor.AttemptsTo`",
		},
	}
	for path, required := range cases {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, fragment := range required {
			if !strings.Contains(string(body), fragment) {
				t.Errorf("%s missing contract %q", path, fragment)
			}
		}
	}
}

func TestAttachmentDocsDoNotClaimAllStepAttachmentsAreEmpty(t *testing.T) {
	staleByPath := map[string][]string{
		"../../src/content/docs/en/guides/21_attachments.md": {
			"produces attachments only for the test result passed to `OnTestFinish`",
			"all current step results, carry none",
		},
		"../../src/content/docs/en/guides/31_custom_reporter.md": {
			"only `OnTestFinish` receives produced attachments",
			"Step results currently have no attachments",
		},
	}
	for path, staleFragments := range staleByPath {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, stale := range staleFragments {
			if strings.Contains(string(body), stale) {
				t.Errorf("%s contains stale attachment claim %q", path, stale)
			}
		}
	}
}

func TestSectionRootsAndCorrectExternalLinks(t *testing.T) {
	configBody, err := os.ReadFile("../../astro.config.mjs")
	if err != nil {
		t.Fatal(err)
	}
	config := string(configBody)
	for _, route := range []string{"/en/get_started/", "/en/core_concepts/", "/en/guides/", "/en/examples/", "/en/api/"} {
		if !strings.Contains(config, `"`+route+`"`) {
			t.Errorf("missing redirect for section root %s", route)
		}
	}
	workflowBody, err := os.ReadFile("../../.github/workflows/deploy.yml")
	if err != nil {
		t.Fatal(err)
	}
	for _, artifact := range []string{"dist/en/get_started/index.html", "dist/en/core_concepts/index.html", "dist/en/guides/index.html", "dist/en/examples/index.html", "dist/en/api/index.html"} {
		if !strings.Contains(string(workflowBody), artifact) {
			t.Errorf("workflow does not check section artifact %s", artifact)
		}
	}
	screenplayBody, err := os.ReadFile("../../src/content/docs/en/core_concepts/1_screenplay.md")
	if err != nil {
		t.Fatal(err)
	}
	screenplay := string(screenplayBody)
	for _, stale := range []string{"understanding-screenplay-(part-1)", "https://blog.mattwynne.net/"} {
		if strings.Contains(screenplay, stale) {
			t.Errorf("screenplay guide contains stale external URL %q", stale)
		}
	}
	for _, current := range []string{"https://cucumber.io/blog/bdd/understanding-screenplay-part-1/", "https://mattwynne.net/"} {
		if !strings.Contains(screenplay, current) {
			t.Errorf("screenplay guide missing corrected external URL %q", current)
		}
	}
}

func TestGeneratedAPIMarkdownIsIgnored(t *testing.T) {
	body, err := os.ReadFile("../../.gitignore")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "src/content/docs/en/api/") {
		t.Fatal("generated API Markdown directory is not ignored")
	}
}

func TestDeployWorkflowEnforcesPRLeastPrivilegeAndImmutableActions(t *testing.T) {
	body, err := os.ReadFile("../../.github/workflows/deploy.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(body)
	for _, required := range []string{
		"  pull_request:\n    branches: [main]",
		"  build:\n    name: Build\n    permissions:\n      contents: read",
		"  deploy:\n    name: Deploy\n    if: github.event_name != 'pull_request'",
		"      pages: write",
		"      id-token: write",
	} {
		if !strings.Contains(workflow, required) {
			t.Errorf("workflow missing policy fragment %q", required)
		}
	}
	buildStart := strings.Index(workflow, "  build:")
	deployStart := strings.Index(workflow, "  deploy:")
	if buildStart < 0 || deployStart < 0 || deployStart <= buildStart {
		t.Fatal("could not identify build and deploy jobs")
	}
	build := workflow[buildStart:deployStart]
	for _, forbidden := range []string{"pages: write", "id-token: write"} {
		if strings.Contains(build, forbidden) {
			t.Errorf("build job has forbidden permission %q", forbidden)
		}
	}
	usesLine := regexp.MustCompile(`(?m)^\s*uses:\s*([^\s#]+)(?:\s+#\s*(v[^\s]+))?\s*$`)
	matches := usesLine.FindAllStringSubmatch(workflow, -1)
	if len(matches) == 0 {
		t.Fatal("workflow has no actions")
	}
	fullPin := regexp.MustCompile(`^[^@]+@[0-9a-f]{40}$`)
	for _, match := range matches {
		if !fullPin.MatchString(match[1]) {
			t.Errorf("action is not pinned to a full immutable SHA: %s", match[1])
		}
		if match[2] == "" {
			t.Errorf("pinned action lacks a retained version comment: %s", match[1])
		}
	}
}
