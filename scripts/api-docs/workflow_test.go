package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

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
