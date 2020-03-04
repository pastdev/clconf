package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestTemplateCmd(t *testing.T) {
	temp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Error making temp folder %v", err)
	}
	defer os.RemoveAll(temp)

	testDataPath := filepath.Join("..", "testdata")
	rootCmd.SetArgs([]string{"template",
		"--yaml", filepath.Join(testDataPath, "testconfig.yml"),
		testDataPath, temp})

	if err = rootCmd.Execute(); err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	resultPath := filepath.Join(temp, "testtemplate.txt")
	actual, err := ioutil.ReadFile(resultPath)
	if err != nil {
		t.Errorf("Error reading %q: %v", resultPath, err)
	}
	expected := "db.pastdev.com"
	if string(actual) != expected {
		t.Errorf("Content of %q was not as expected, %q != %q", resultPath, actual, expected)
	}
}
