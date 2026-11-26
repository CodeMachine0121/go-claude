package claude

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const modulePath = "github.com/CodeMachine0121/go-claude"

// productionFiles lists every Go file compiled into a program that imports this
// library: no tests, no examples, no integration tests.
func productionFiles(t *testing.T) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (path == "examples" || path == "test" || strings.HasPrefix(entry.Name(), ".")) && path != "." {
			return filepath.SkipDir
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}
		return nil
	})
	require.NoError(t, err)
	require.NotEmpty(t, files)
	return files
}

func TestProductionCodeAddsNoThirdPartyDependencies(t *testing.T) {
	for _, file := range productionFiles(t) {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		require.NoError(t, err)
		for _, spec := range parsed.Imports {
			imported, err := strconv.Unquote(spec.Path.Value)
			require.NoError(t, err)
			firstElement, _, _ := strings.Cut(imported, "/")
			isStandardLibrary := !strings.Contains(firstElement, ".")
			isThisModule := imported == modulePath || strings.HasPrefix(imported, modulePath+"/")
			assert.Truef(t, isStandardLibrary || isThisModule, "%s imports third-party %q", file, imported)
		}
	}
}

func TestProductionCodeNeverTalksToTheNetwork(t *testing.T) {
	for _, file := range productionFiles(t) {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		require.NoError(t, err)
		for _, spec := range parsed.Imports {
			imported, err := strconv.Unquote(spec.Path.Value)
			require.NoError(t, err)
			assert.Falsef(t, imported == "net" || strings.HasPrefix(imported, "net/"), "%s imports %q", file, imported)
		}
	}
}

func TestProductionCodeNeverReadsStoredSignInSecrets(t *testing.T) {
	forbidden := []string{".credentials.json", "find-generic-password", "Claude Code-credentials", "keychain"}
	for _, file := range productionFiles(t) {
		source, err := os.ReadFile(file)
		require.NoError(t, err)
		for _, marker := range forbidden {
			assert.NotContainsf(t, strings.ToLower(string(source)), strings.ToLower(marker), "%s refers to stored sign-in secrets", file)
		}
	}
}
