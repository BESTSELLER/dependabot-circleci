package circleci

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog/log"
)

func getTestCases() map[string]*string {
	path := "../.test_cases"
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatal().Err(err)
	}
	result := make(map[string]*string, len(files))
	for _, f := range files {
		fileName := f.Name()
		ext := strings.ToLower(filepath.Ext(fileName))
		if ext != ".yml" && ext != ".yaml" {
			continue
		}
		filePath := filepath.Join(path, fileName)
		fmt.Println(f.Name())

		content, _ := os.ReadFile(filePath)
		contentString := string(content)
		result[fileName] = &contentString
	}

	return result
}
func TestGetUpdates(t *testing.T) {
	tests := getTestCases()
	SHA := "ABC"
	updates := map[string]Update{}

	for k, v := range tests {
		t.Run(k, func(t *testing.T) {
			ScanFileUpdates(&updates, v, &k, &SHA)
		})
	}
}
