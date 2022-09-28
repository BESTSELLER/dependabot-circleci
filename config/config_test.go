package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestReadRepoConfigAsExpected tests if the config is loaded as expected
func TestReadRepoConfigAsExpected(t *testing.T) {
	// arrange
	var jsonExpected = []byte(`{
		"target-branch": "branch",
		"reviewers": ["reviewer"],
		"assignees": ["assignee"],
		"labels": ["labels"],
		"directory": "directory",
		"schedule": "daily",
	}`)

	// act
	repoConfig, err := ReadRepoConfig(jsonExpected)
	if err != nil {
		t.Fatal(err)
	}

	err = repoConfig.IsValid()
	if err != nil {
		t.Fatal(err)
	}

	// assert
	assert.Equal(t, "daily", repoConfig.Schedule)

}

// TestReadRepoConfigScheduleNotAsExpected tests if schedule is set correct
func TestReadRepoConfigScheduleNotAsExpected(t *testing.T) {
	// arrange
	var jsonExpected = []byte(`{
		"target-branch": "branch",
		"reviewers": ["reviewer"],
		"assignees": ["assignee"],
		"labels": ["labels"],
		"directory": "directory",
		"schedule": "wrongvalue",
	}`)

	// act
	repoConfig, err := ReadRepoConfig(jsonExpected)
	if err != nil {
		t.Fatal(err)
	}

	err = repoConfig.IsValid()

	// assert
	assert.NotNil(t, err, "error expected")

}
