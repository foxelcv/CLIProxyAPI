package interactions_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoeForge/moego/regexp"
)

func TestInteractionsTranslatorsDoNotImportGeminiTranslators(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	scanDirs := []string{
		"internal/translator/openai/interactions",
		"internal/translator/claude/interactions",
		"internal/translator/codex/interactions",
		"internal/translator/antigravity/interactions",
	}
	forbidden := regexp.MustCompile(`"github\.com/router-for-me/CLIProxyAPI/v8/internal/translator/[^"]*/gemini[^"]*"`)
	var violations []string
	for _, scanDir := range scanDirs {
		root := filepath.Join(repoRoot, scanDir)
		errWalk := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			data, errRead := os.ReadFile(path)
			if errRead != nil {
				return errRead
			}
			if hasForbiddenGeminiImport(forbidden, data, scanDir) {
				rel, errRel := filepath.Rel(repoRoot, path)
				if errRel != nil {
					rel = path
				}
				violations = append(violations, rel)
			}
			return nil
		})
		if errWalk != nil {
			t.Fatalf("scan %s: %v", scanDir, errWalk)
		}
	}
	if len(violations) > 0 {
		t.Fatalf("non-Gemini Interactions translators import Gemini translators: %s", strings.Join(violations, ", "))
	}
}

// sameProviderGeminiHelpers lists provider helper packages an Interactions
// translator may import even though their path ends in /gemini. Upstream
// (ed885bdc) reuses Antigravity's own request sanitizers from
// translator/antigravity/gemini inside the Antigravity Interactions translator;
// that shares provider wire-format helpers and does not route Interactions
// requests through a Gemini translator.
var sameProviderGeminiHelpers = map[string]string{
	"internal/translator/antigravity/interactions": `"github.com/router-for-me/CLIProxyAPI/v8/internal/translator/antigravity/gemini"`,
}

func hasForbiddenGeminiImport(forbidden *regexp.Regexp, data []byte, scanDir string) bool {
	allowed := sameProviderGeminiHelpers[scanDir]
	for _, match := range forbidden.FindAll(data, -1) {
		if allowed != "" && string(match) == allowed {
			continue
		}
		return true
	}
	return false
}
