package session

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// EncodeProjectDir replaces special characters with dashes, matching
// the encoding Claude Code uses for project directories.
func EncodeProjectDir(cwd string) string {
	replacer := strings.NewReplacer("/", "-", ".", "-", "_", "-")
	return replacer.Replace(cwd)
}

// ListSessionFiles returns all .jsonl files in projectDir sorted by
// modification time descending (most recent first).
func ListSessionFiles(projectDir string) ([]string, error) {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonl") {
			files = append(files, filepath.Join(projectDir, e.Name()))
		}
	}

	sort.Slice(files, func(i, j int) bool {
		si, _ := os.Stat(files[i])
		sj, _ := os.Stat(files[j])
		return si.ModTime().After(sj.ModTime())
	})

	return files, nil
}

// FindActiveSession returns the most recent .jsonl session file in the
// given project directory, or an empty string and false if none exist.
func FindActiveSession(projectDir string) (string, bool) {
	files, err := ListSessionFiles(projectDir)
	if err != nil || len(files) == 0 {
		return "", false
	}

	return files[0], true
}
