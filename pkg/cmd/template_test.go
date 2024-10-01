package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTemplateCmd(t *testing.T) {
	temp := t.TempDir()

	testDataPath := filepath.Join("..", "..", "testdata")
	cmd := rootCmd()
	cmd.SetArgs([]string{"template",
		"--yaml", filepath.Join(testDataPath, "testconfig.yml"),
		testDataPath, temp})

	if err := cmd.Execute(); err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	resultPath := filepath.Join(temp, "testtemplate.txt")
	actual, err := os.ReadFile(resultPath)
	if err != nil {
		t.Errorf("Error reading %q: %v", resultPath, err)
	}
	expected := "db.pastdev.com"
	if string(actual) != expected {
		t.Errorf("Content of %q was not as expected, %q != %q", resultPath, actual, expected)
	}
}

func TestTemplateCmdDelims(t *testing.T) {
	temp := t.TempDir()

	testDataPath := filepath.Join("..", "..", "testdata")
	cmd := rootCmd()
	cmd.SetArgs([]string{"template",
		"--yaml", filepath.Join(testDataPath, "testconfig.yml"),
		"--left-delimiter", "<<",
		"--right-delimiter", ">>",
		testDataPath, temp})

	if err := cmd.Execute(); err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	resultPath := filepath.Join(temp, "testtemplatedelims.txt")
	actual, err := os.ReadFile(resultPath)
	if err != nil {
		t.Errorf("Error reading %q: %v", resultPath, err)
	}
	expected := "db.pastdev.com"
	if string(actual) != expected {
		t.Errorf("Content of %q was not as expected, %q != %q", resultPath, actual, expected)
	}
}
