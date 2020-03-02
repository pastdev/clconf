package clconf

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

const fakeValue = "bar"

type relExpectedPath struct {
	subPath string
	ext     string
}

func makeTestSubfolder(t *testing.T, temp string, subPath string, perms os.FileMode) {
	path := filepath.Join(temp, subPath)
	err := MkdirAllNoUmask(path, perms)
	if err != nil {
		t.Fatalf("Error making temp sub dir %q: %v", path, err)
	}
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Error stating folder %q after creation: %v", path, err)
	}
	if stat.Mode()&0777 != perms {
		t.Fatalf("Created folder %q does not have proper permissions after creation [%o != %o]",
			path, stat.Mode()&0777, perms)
	}
}

func writeTestFile(t *testing.T, temp string, subPath string, perms os.FileMode) {
	path := filepath.Join(temp, subPath)
	content := []byte("{{ getv \"/foo\" }}")
	err := ioutil.WriteFile(path, content, perms)
	if err != nil {
		t.Fatalf("Error making temp file %q: %v", path, err)
	}
	os.Chmod(path, perms)
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Error stating file %q after creation: %v", path, err)
	}
	if stat.Mode() != perms {
		t.Fatalf("Created file %q does not have proper permissions after creation [%o != %o]",
			path, stat.Mode(), perms)
	}
}

func buildTestFolder(t *testing.T) string {
	extension := ".clconf"
	temp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Error making temp folder %v", err)
	}
	makeTestSubfolder(t, temp, "subdir1/subsubdir1", 0775)
	makeTestSubfolder(t, temp, "subdir1/subsubdir2", 0775)
	makeTestSubfolder(t, temp, "subdir2", 0775)
	makeTestSubfolder(t, temp, "emptydir", 0775)

	writeTestFile(t, temp, "yes_basedir.html"+extension, 0646)
	writeTestFile(t, temp, "yes_basedir1.html"+extension, 0640)
	writeTestFile(t, temp, "yes_basedir2.html"+extension, 0640)
	writeTestFile(t, temp, "no_basedir.html", 0641)
	writeTestFile(t, temp, "no_basedir1.html", 0640)
	writeTestFile(t, temp, "subdir1/yes_subdir1.sh"+extension, 0770)
	writeTestFile(t, temp, "subdir1/no_subdir1.sh", 0770)
	writeTestFile(t, temp, "subdir1/subsubdir1/yes_subdir1subsubdir1.sh"+extension, 0775)
	writeTestFile(t, temp, "subdir1/subsubdir1/no_subdir1subsubdir1.sh", 0775)
	writeTestFile(t, temp, "subdir1/subsubdir2/yes_subdir1subsubdir2.sh"+extension, 0777)
	writeTestFile(t, temp, "subdir1/subsubdir2/no_subdir1subsubdir2.sh", 0777)
	writeTestFile(t, temp, "subdir2/yes_subdir2.sh"+extension, 0777)
	writeTestFile(t, temp, "subdir2/no_subdir2.sh", 0777)

	return temp
}

func normalizePaths(paths []pathWithRelative) {
	sort.SliceStable(paths, func(i, j int) bool {
		return paths[i].full < paths[j].full || paths[i].rel < paths[j].rel
	})
}

func testFindTemplates(t *testing.T, message string, extension string, subPath string, expectedFlat bool, expected []string) {
	temp := buildTestFolder(t)
	defer os.RemoveAll(temp)

	paths, err := findTemplates(filepath.Join(temp, subPath), extension)
	if err != nil {
		t.Fatalf("Error running findTemplates (%s): %v", message, err)
	}

	if len(expected) == 0 {
		if len(paths) != 0 {
			t.Errorf("Paths wasn't empty when it was supposed to be: %v", paths)
		}
		return
	}

	expectedTemplates := []pathWithRelative{}
	for _, exp := range expected {
		relPath := exp
		if expectedFlat {
			relPath = filepath.Base(relPath)
		}
		expectedTemplates = append(expectedTemplates, pathWithRelative{
			full: filepath.Join(temp, exp),
			rel:  relPath,
		})
	}

	normalizePaths(paths)
	if !reflect.DeepEqual(paths, expectedTemplates) {
		t.Errorf("TestFindTemplates %s [%v] != [%v]", message, paths, expectedTemplates)
	}
}

func TestFindTemplatesWithExtension(t *testing.T) {
	testFindTemplates(t, "With Extension", ".clconf", "", false, []string{
		"subdir1/subsubdir1/yes_subdir1subsubdir1.sh.clconf",
		"subdir1/subsubdir2/yes_subdir1subsubdir2.sh.clconf",
		"subdir1/yes_subdir1.sh.clconf",
		"subdir2/yes_subdir2.sh.clconf",
		"yes_basedir.html.clconf",
		"yes_basedir1.html.clconf",
		"yes_basedir2.html.clconf",
	})
}

func TestFindTemplatesWithoutExtension(t *testing.T) {
	testFindTemplates(t, "Without Extension", "", "", false, []string{
		"no_basedir.html",
		"no_basedir1.html",
		"subdir1/no_subdir1.sh",
		"subdir1/subsubdir1/no_subdir1subsubdir1.sh",
		"subdir1/subsubdir1/yes_subdir1subsubdir1.sh.clconf",
		"subdir1/subsubdir2/no_subdir1subsubdir2.sh",
		"subdir1/subsubdir2/yes_subdir1subsubdir2.sh.clconf",
		"subdir1/yes_subdir1.sh.clconf",
		"subdir2/no_subdir2.sh",
		"subdir2/yes_subdir2.sh.clconf",
		"yes_basedir.html.clconf",
		"yes_basedir1.html.clconf",
		"yes_basedir2.html.clconf",
	})
}

func TestFindTemplatesSingleFileMatchingExtension(t *testing.T) {
	testFindTemplates(t, "Single File Matching Extension", ".clconf", "subdir1/subsubdir1/yes_subdir1subsubdir1.sh.clconf", true, []string{
		"subdir1/subsubdir1/yes_subdir1subsubdir1.sh.clconf",
	})
}

func TestFindTemplatesSingleFileNotMatchingExtension(t *testing.T) {
	testFindTemplates(t, "Single File Not Matching Extension", ".clconf", "subdir1/subsubdir1/no_subdir1subsubdir1.sh", true, []string{
		"subdir1/subsubdir1/no_subdir1subsubdir1.sh",
	})
}

func TestFindTemplatesSingleEmptyFolder(t *testing.T) {
	testFindTemplates(t, "Empty Folder", ".clconf", "emptydir", true, []string{})
}

func defaultContext() (*TemplateSettings, *SecretAgent) {
	return &TemplateSettings{
		Flatten:           false,
		KeepExistingPerms: false,
		Rm:                false,
		DirMode:           "753", //We use a wonky value to look for it later
		Extension:         ".clconf",
	}, NewSecretAgent([]byte(""))
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func checkFile(t *testing.T, path string, expectedPerms os.FileMode) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Error reading %q: %v", path, err)
	}

	stat, _ := os.Stat(path)
	if stat.Mode() != expectedPerms {
		t.Fatalf("File %q has mode %o that does not match expected value %o", path, stat.Mode(), expectedPerms)
	}

	if string(content) != fakeValue {
		t.Fatalf("File %q content %q does not match expected %q", path, content, fakeValue)
	}
}

func TestProcessTemplateInPlace(t *testing.T) {
	extension := ".clconf"
	temp := buildTestFolder(t)
	defer os.RemoveAll(temp)

	context, secretAgent := defaultContext()

	value := map[interface{}]interface{}{"foo": fakeValue}

	template := filepath.Join(temp, "yes_basedir.html"+extension)
	_, err := context.processTemplate(pathWithRelative{
		full: template,
		rel:  filepath.Base(template),
	}, "", value, secretAgent)
	if err != nil {
		t.Fatalf("processTemplate reported error: %v", err)
	}

	checkFile(t, filepath.Join(temp, "yes_basedir.html"), 0646)

	if !exists(template) {
		t.Errorf("Template went missing when it wasn't supposed to!")
	}

	context.FileMode = "610"
	_, err = context.processTemplate(pathWithRelative{
		full: template,
		rel:  filepath.Base(template),
	}, "", value, secretAgent)
	if err != nil {
		t.Fatalf("processTemplate reported error: %v", err)
	}

	checkFile(t, filepath.Join(temp, "yes_basedir.html"), 0610)

	context.Rm = true
	context.FileMode = "601"
	context.KeepExistingPerms = true
	_, err = context.processTemplate(pathWithRelative{
		full: template,
		rel:  filepath.Base(template),
	}, "", value, secretAgent)
	if err != nil {
		t.Fatalf("processTemplate reported error: %v", err)
	}

	checkFile(t, filepath.Join(temp, "yes_basedir.html"), 0610)
	if exists(template) {
		t.Errorf("Template went missing when it wasn't supposed to!")
	}

	template = filepath.Join(temp, "subdir1/subsubdir2/no_subdir1subsubdir2.sh")
	_, err = context.processTemplate(pathWithRelative{
		full: template,
		rel:  filepath.Base(template),
	}, "", value, secretAgent)
	if err != nil {
		t.Fatalf("processTemplate reported error: %v", err)
	}

	checkFile(t, filepath.Join(temp, "subdir1/subsubdir2/no_subdir1subsubdir2.sh"), 0777)
}

func TestProcessTemplateFolder(t *testing.T) {
	extension := ".clconf"
	temp := buildTestFolder(t)
	defer os.RemoveAll(temp)

	dest := filepath.Join(temp, "dest")
	os.Mkdir(dest, 0755)

	context, secretAgent := defaultContext()

	value := map[interface{}]interface{}{"foo": fakeValue}

	template := filepath.Join(temp, "yes_basedir.html"+extension)
	_, err := context.processTemplate(pathWithRelative{
		full: template,
		rel:  filepath.Base(template),
	}, dest, value, secretAgent)
	if err != nil {
		t.Fatalf("processTemplate reported error: %v", err)
	}

	checkFile(t, filepath.Join(dest, "yes_basedir.html"), 0646)

	if !exists(template) {
		t.Errorf("Template went missing when it wasn't supposed to!")
	}

	context.Rm = true
	template = filepath.Join(temp, "subdir1/subsubdir1/no_subdir1subsubdir1.sh")
	_, err = context.processTemplate(pathWithRelative{
		full: template,
		rel:  "subdir1/subsubdir1/no_subdir1subsubdir1.sh",
	}, dest, value, secretAgent)
	if err != nil {
		t.Fatalf("processTemplate reported error: %v", err)
	}

	checkFile(t, filepath.Join(dest, "subdir1/subsubdir1/no_subdir1subsubdir1.sh"), 0775)

	if exists(template) {
		t.Errorf("Template still exists when it wasn't supposed to!")
	}

}

func TestProcessTemplateFolderFlatten(t *testing.T) {
	temp := buildTestFolder(t)
	defer os.RemoveAll(temp)

	dest := filepath.Join(temp, "dest")
	os.Mkdir(dest, 0755)

	context, secretAgent := defaultContext()

	value := map[interface{}]interface{}{"foo": fakeValue}

	context.Flatten = true
	template := filepath.Join(temp, "subdir1/subsubdir1/no_subdir1subsubdir1.sh")
	_, err := context.processTemplate(pathWithRelative{
		full: template,
		rel:  "subdir1/subsubdir1/no_subdir1subsubdir1.sh",
	}, dest, value, secretAgent)
	if err != nil {
		t.Fatalf("processTemplate reported error: %v", err)
	}

	checkFile(t, filepath.Join(dest, "no_subdir1subsubdir1.sh"), 0775)
}

func TestProcessTemplatesWithExtension(t *testing.T) {
	temp := buildTestFolder(t)
	defer os.RemoveAll(temp)

	dest := filepath.Join(temp, "dest")
	os.Mkdir(dest, 0755)

	context, secretAgent := defaultContext()

	value := map[interface{}]interface{}{"foo": fakeValue}

	_, err := context.ProcessTemplates([]string{temp}, dest, value, secretAgent)
	if err != nil {
		t.Errorf("Error processing templates: %v", err)
	}
	checkFile(t, filepath.Join(dest, "yes_basedir.html"), 0646)
	checkFile(t, filepath.Join(dest, "yes_basedir1.html"), 0640)
	checkFile(t, filepath.Join(dest, "yes_basedir2.html"), 0640)
	checkFile(t, filepath.Join(dest, "subdir1/yes_subdir1.sh"), 0770)
	checkFile(t, filepath.Join(dest, "subdir1/subsubdir1/yes_subdir1subsubdir1.sh"), 0775)
	checkFile(t, filepath.Join(dest, "subdir1/subsubdir2/yes_subdir1subsubdir2.sh"), 0777)
	checkFile(t, filepath.Join(dest, "subdir2/yes_subdir2.sh"), 0777)

	if exists(filepath.Join(dest, "no_basedir.html")) {
		t.Errorf("File exists when it shouldn't")
	}
	if exists(filepath.Join(dest, "no_basedir1.html")) {
		t.Errorf("File exists when it shouldn't")
	}
	if exists(filepath.Join(dest, "subdir1/no_subdir1.sh")) {
		t.Errorf("File exists when it shouldn't")
	}
	if exists(filepath.Join(dest, "subdir1/subsubdir1/no_subdir1subsubdir1.sh")) {
		t.Errorf("File exists when it shouldn't")
	}
	if exists(filepath.Join(dest, "subdir1/subsubdir2/no_subdir1subsubdir2.sh")) {
		t.Errorf("File exists when it shouldn't")
	}
	if exists(filepath.Join(dest, "subdir2/no_subdir2.sh")) {
		t.Errorf("File exists when it shouldn't")
	}

}
