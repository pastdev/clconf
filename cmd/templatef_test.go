package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/pastdev/clconf/v2/clconf"
)

const fakeValue = "bar"

func makeTestSubfolder(t *testing.T, temp string, subPath string, perms os.FileMode) {
	path := filepath.Join(temp, subPath)
	err := os.MkdirAll(path, perms)
	if err != nil {
		t.Fatalf("Error making temp sub dir %q: %v", path, err)
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

func buildTestFolder(t *testing.T, extension string) string {
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
		return paths[i].fullPath < paths[j].fullPath || paths[i].relPath < paths[j].relPath
	})
}

func TestFindTemplatesWithExtension(t *testing.T) {
	extension := ".clconf"
	temp := buildTestFolder(t, extension)
	defer os.RemoveAll(temp)

	paths, err := findTemplates(temp, extension)
	if err != nil {
		t.Fatalf("Error running findTemplates: %v", err)
	}

	normalizePaths(paths)

	expected := []pathWithRelative{
		{fullPath: filepath.Join(temp, "subdir1/subsubdir1/yes_subdir1subsubdir1.sh"+extension), relPath: "subdir1/subsubdir1/yes_subdir1subsubdir1.sh" + extension},
		{fullPath: filepath.Join(temp, "subdir1/subsubdir2/yes_subdir1subsubdir2.sh"+extension), relPath: "subdir1/subsubdir2/yes_subdir1subsubdir2.sh" + extension},
		{fullPath: filepath.Join(temp, "subdir1/yes_subdir1.sh"+extension), relPath: "subdir1/yes_subdir1.sh" + extension},
		{fullPath: filepath.Join(temp, "subdir2/yes_subdir2.sh"+extension), relPath: "subdir2/yes_subdir2.sh" + extension},
		{fullPath: filepath.Join(temp, "yes_basedir.html"+extension), relPath: "yes_basedir.html" + extension},
		{fullPath: filepath.Join(temp, "yes_basedir1.html"+extension), relPath: "yes_basedir1.html" + extension},
		{fullPath: filepath.Join(temp, "yes_basedir2.html"+extension), relPath: "yes_basedir2.html" + extension},
	}

	if !reflect.DeepEqual(paths, expected) {
		t.Errorf("Path didn't match expected [%v] != [%v]", paths, expected)
	}
}

func TestFindTemplatesWithoutExtension(t *testing.T) {
	extension := ""
	temp := buildTestFolder(t, extension)
	defer os.RemoveAll(temp)

	paths, err := findTemplates(temp, extension)
	if err != nil {
		t.Fatalf("Error running findTemplates: %v", err)
	}

	normalizePaths(paths)

	expected := []pathWithRelative{
		{fullPath: filepath.Join(temp, "no_basedir.html"), relPath: "no_basedir.html"},
		{fullPath: filepath.Join(temp, "no_basedir1.html"), relPath: "no_basedir1.html"},
		{fullPath: filepath.Join(temp, "subdir1/no_subdir1.sh"), relPath: "subdir1/no_subdir1.sh"},
		{fullPath: filepath.Join(temp, "subdir1/subsubdir1/no_subdir1subsubdir1.sh"), relPath: "subdir1/subsubdir1/no_subdir1subsubdir1.sh"},
		{fullPath: filepath.Join(temp, "subdir1/subsubdir1/yes_subdir1subsubdir1.sh"+extension), relPath: "subdir1/subsubdir1/yes_subdir1subsubdir1.sh" + extension},
		{fullPath: filepath.Join(temp, "subdir1/subsubdir2/no_subdir1subsubdir2.sh"), relPath: "subdir1/subsubdir2/no_subdir1subsubdir2.sh"},
		{fullPath: filepath.Join(temp, "subdir1/subsubdir2/yes_subdir1subsubdir2.sh"+extension), relPath: "subdir1/subsubdir2/yes_subdir1subsubdir2.sh" + extension},
		{fullPath: filepath.Join(temp, "subdir1/yes_subdir1.sh"+extension), relPath: "subdir1/yes_subdir1.sh" + extension},
		{fullPath: filepath.Join(temp, "subdir2/no_subdir2.sh"), relPath: "subdir2/no_subdir2.sh"},
		{fullPath: filepath.Join(temp, "subdir2/yes_subdir2.sh"+extension), relPath: "subdir2/yes_subdir2.sh" + extension},
		{fullPath: filepath.Join(temp, "yes_basedir.html"+extension), relPath: "yes_basedir.html" + extension},
		{fullPath: filepath.Join(temp, "yes_basedir1.html"+extension), relPath: "yes_basedir1.html" + extension},
		{fullPath: filepath.Join(temp, "yes_basedir2.html"+extension), relPath: "yes_basedir2.html" + extension},
	}

	if !reflect.DeepEqual(paths, expected) {
		t.Errorf("Path didn't match expected [%v] != [%v]", paths, expected)
	}
}

func TestFindTemplatesSingleFileMatchingExtension(t *testing.T) {
	extension := ".clconf"
	temp := buildTestFolder(t, extension)
	defer os.RemoveAll(temp)

	paths, err := findTemplates(filepath.Join(temp, "subdir1/subsubdir1/yes_subdir1subsubdir1.sh"+extension), extension)
	if err != nil {
		t.Fatalf("Error running findTemplates: %v", err)
	}

	normalizePaths(paths)

	expected := []pathWithRelative{
		{fullPath: filepath.Join(temp, "subdir1/subsubdir1/yes_subdir1subsubdir1.sh"+extension), relPath: "yes_subdir1subsubdir1.sh" + extension},
	}

	if !reflect.DeepEqual(paths, expected) {
		t.Errorf("Path didn't match expected [%v] != [%v]", paths, expected)
	}
}

func TestFindTemplatesSingleFileNotMatchingExtension(t *testing.T) {
	extension := ".clconf"
	temp := buildTestFolder(t, extension)
	defer os.RemoveAll(temp)

	paths, err := findTemplates(filepath.Join(temp, "subdir1/subsubdir1/no_subdir1subsubdir1.sh"), extension)
	if err != nil {
		t.Fatalf("Error running findTemplates: %v", err)
	}

	normalizePaths(paths)

	expected := []pathWithRelative{
		{fullPath: filepath.Join(temp, "subdir1/subsubdir1/no_subdir1subsubdir1.sh"), relPath: "no_subdir1subsubdir1.sh"},
	}

	if !reflect.DeepEqual(paths, expected) {
		t.Errorf("Path didn't match expected [%v] != [%v]", paths, expected)
	}
}

func TestFindTemplatesSingleEmptyFolder(t *testing.T) {
	extension := ".clconf"
	temp := buildTestFolder(t, extension)
	defer os.RemoveAll(temp)

	paths, err := findTemplates(filepath.Join(temp, "emptydir"), extension)
	if err != nil {
		t.Fatalf("Error running findTemplates: %v", err)
	}

	normalizePaths(paths)

	if len(paths) != 0 {
		t.Errorf("Paths wasn't empty when it was supposed to be: %v", paths)
	}
}

func defaultContext() (*templatefContext, *clconf.SecretAgent) {
	return &templatefContext{
		inPlace:           false,
		flatten:           false,
		keepExistingPerms: false,
		rm:                false,
		dirMode:           "753", //We use a wonky value to look for it later
		extension:         ".clconf",
	}, clconf.NewSecretAgent([]byte(""))
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func checkFile(t *testing.T, path string, expectedPerms os.FileMode) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Error reading %q: %v", path, err)
	}

	stat, _ := os.Stat(path)
	if stat.Mode() != expectedPerms {
		return fmt.Errorf("File %q has mode %o that does not match expected value %o", path, stat.Mode(), expectedPerms)
	}

	if string(content) != fakeValue {
		return fmt.Errorf("File %q content %q does not match expected %q", path, content, fakeValue)
	}
	return nil
}

func TestProcessTemplateInPlace(t *testing.T) {
	extension := ".clconf"
	temp := buildTestFolder(t, extension)
	defer os.RemoveAll(temp)

	context, secretAgent := defaultContext()

	value := map[interface{}]interface{}{"foo": fakeValue}

	template := filepath.Join(temp, "yes_basedir.html"+extension)
	err := context.processTemplate(pathWithRelative{
		fullPath: template,
		relPath:  filepath.Base(template),
	}, "", value, secretAgent)
	if err != nil {
		t.Fatalf("processTemplate reported error: %v", err)
	}

	err = checkFile(t, filepath.Join(temp, "yes_basedir.html"), 0646)
	if err != nil {
		t.Errorf("yes_basdir.html isn't proper: %v", err)
	}

	if !exists(template) {
		t.Errorf("Template went missing when it wasn't supposed to!")
	}

	context.fileMode = "610"
	err = context.processTemplate(pathWithRelative{
		fullPath: template,
		relPath:  filepath.Base(template),
	}, "", value, secretAgent)
	if err != nil {
		t.Fatalf("processTemplate reported error: %v", err)
	}

	err = checkFile(t, filepath.Join(temp, "yes_basedir.html"), 0610)
	if err != nil {
		t.Errorf("yes_basdir.html isn't proper: %v", err)
	}

	context.rm = true
	context.fileMode = "601"
	context.keepExistingPerms = true
	err = context.processTemplate(pathWithRelative{
		fullPath: template,
		relPath:  filepath.Base(template),
	}, "", value, secretAgent)
	if err != nil {
		t.Fatalf("processTemplate reported error: %v", err)
	}

	err = checkFile(t, filepath.Join(temp, "yes_basedir.html"), 0610)
	if err != nil {
		t.Errorf("yes_basdir.html isn't proper: %v", err)
	}
	if exists(template) {
		t.Errorf("Template went missing when it wasn't supposed to!")
	}

	template = filepath.Join(temp, "subdir1/subsubdir2/no_subdir1subsubdir2.sh")
	err = context.processTemplate(pathWithRelative{
		fullPath: template,
		relPath:  filepath.Base(template),
	}, "", value, secretAgent)
	if err != nil {
		t.Fatalf("processTemplate reported error: %v", err)
	}

	err = checkFile(t, filepath.Join(temp, "subdir1/subsubdir2/no_subdir1subsubdir2.sh"), 0777)
	if err != nil {
		t.Errorf("no_subdir1subsubdir2.sh isn't proper: %v", err)
	}
}

func TestProcessTemplateFolder(t *testing.T) {
	extension := ".clconf"
	temp := buildTestFolder(t, extension)
	defer os.RemoveAll(temp)

	dest := filepath.Join(temp, "dest")
	os.Mkdir(dest, 0755)

	context, secretAgent := defaultContext()

	value := map[interface{}]interface{}{"foo": fakeValue}

	template := filepath.Join(temp, "yes_basedir.html"+extension)
	err := context.processTemplate(pathWithRelative{
		fullPath: template,
		relPath:  filepath.Base(template),
	}, dest, value, secretAgent)
	if err != nil {
		t.Fatalf("processTemplate reported error: %v", err)
	}

	err = checkFile(t, filepath.Join(dest, "yes_basedir.html"), 0646)
	if err != nil {
		t.Errorf("yes_basdir.html isn't proper: %v", err)
	}

	if !exists(template) {
		t.Errorf("Template went missing when it wasn't supposed to!")
	}

	context.rm = true
	template = filepath.Join(temp, "subdir1/subsubdir1/no_subdir1subsubdir1.sh")
	err = context.processTemplate(pathWithRelative{
		fullPath: template,
		relPath:  "subdir1/subsubdir1/no_subdir1subsubdir1.sh",
	}, dest, value, secretAgent)
	if err != nil {
		t.Fatalf("processTemplate reported error: %v", err)
	}

	err = checkFile(t, filepath.Join(dest, "subdir1/subsubdir1/no_subdir1subsubdir1.sh"), 0775)
	if err != nil {
		t.Errorf("yes_basdir.html isn't proper: %v", err)
	}

	if exists(template) {
		t.Errorf("Template still exists when it wasn't supposed to!")
	}

}

func TestProcessTemplateFolderFlatten(t *testing.T) {
	extension := ".clconf"
	temp := buildTestFolder(t, extension)
	defer os.RemoveAll(temp)

	dest := filepath.Join(temp, "dest")
	os.Mkdir(dest, 0755)

	context, secretAgent := defaultContext()

	value := map[interface{}]interface{}{"foo": fakeValue}

	context.flatten = true
	template := filepath.Join(temp, "subdir1/subsubdir1/no_subdir1subsubdir1.sh")
	err := context.processTemplate(pathWithRelative{
		fullPath: template,
		relPath:  "subdir1/subsubdir1/no_subdir1subsubdir1.sh",
	}, dest, value, secretAgent)
	if err != nil {
		t.Fatalf("processTemplate reported error: %v", err)
	}

	err = checkFile(t, filepath.Join(dest, "no_subdir1subsubdir1.sh"), 0775)
	if err != nil {
		t.Errorf("yes_basdir.html isn't proper: %v", err)
	}
}

func TestProcessTemplatesNoExtension(t *testing.T) {
	extension := ".clconf"
	temp := buildTestFolder(t, extension)
	defer os.RemoveAll(temp)

	dest := filepath.Join(temp, "dest")
	os.Mkdir(dest, 0755)

	context, secretAgent := defaultContext()

	value := map[interface{}]interface{}{"foo": fakeValue}

	err := context.processTemplates([]string{temp}, dest, value, secretAgent)
	if err != nil {
		t.Errorf("Error processing templates: %v", err)
	}
	err = checkFile(t, filepath.Join(dest, "yes_basedir.html"), 0646)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = checkFile(t, filepath.Join(dest, "yes_basedir1.html"), 0640)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = checkFile(t, filepath.Join(dest, "yes_basedir2.html"), 0640)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = checkFile(t, filepath.Join(dest, "subdir1/yes_subdir1.sh"), 0770)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = checkFile(t, filepath.Join(dest, "subdir1/subsubdir1/yes_subdir1subsubdir1.sh"), 0775)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = checkFile(t, filepath.Join(dest, "subdir1/subsubdir2/yes_subdir1subsubdir2.sh"), 0777)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = checkFile(t, filepath.Join(dest, "subdir2/yes_subdir2.sh"), 0777)
	if err != nil {
		t.Errorf("%v", err)
	}
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
