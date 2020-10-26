package circleci

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func getTestCases() map[string]*yaml.Node {
	path := "../.test_cases"
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal().Err(err)
	}
	result := make(map[string]*yaml.Node, len(files))
	for _, f := range files {
		fileName := f.Name()
		ext := strings.ToLower(filepath.Ext(fileName))
		if ext != ".yml" && ext != ".yaml" {
			continue
		}
		filePath := filepath.Join(path, fileName)
		fmt.Println(f.Name())

		content, _ := ioutil.ReadFile(filePath)
		var cciconfig yaml.Node
		err = yaml.Unmarshal(content, &cciconfig)
		if err != nil {
			continue
		}
		result[fileName] = &cciconfig
	}

	return result
}
func TestGetUpdates(t *testing.T) {
	tests := getTestCases()

	for k, v := range tests {
		t.Run(k, func(t *testing.T) {
			GetUpdates(v)
			// if got := GetUpdates(v); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("GetUpdates() = %v, want %v", got, tt.want)
			// }
		})
	}
}
