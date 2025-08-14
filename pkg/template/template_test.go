//nolint:goconst
package template

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"testing"

	"github.com/pastdev/clconf/v3/pkg/secret"
)

const fakeValue = "bar"

func makeTestSubfolder(
	t *testing.T,
	temp string,
	subPath string,
	perms os.FileMode, //nolint:unparam
) {
	path := filepath.Join(temp, subPath)
	err := MkdirAllNoUmask(path, perms)
	if err != nil {
		t.Fatalf("Error making temp sub dir %q: %v", path, err)
	}
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Error stating folder %q after creation: %v", path, err)
	}
	if runtime.GOOS != "windows" && stat.Mode()&0777 != perms {
		t.Fatalf("Created folder %q does not have proper permissions after creation [%o != %o]",
			path, stat.Mode()&0777, perms)
	}
}

func writeTestFile(t *testing.T, temp string, subPath string, perms os.FileMode) {
	path := filepath.Join(temp, subPath)
	content := []byte("{{ getv \"/foo\" }}")
	err := os.WriteFile(path, content, perms)
	if err != nil {
		t.Fatalf("Error making temp file %q: %v", path, err)
	}
	err = os.Chmod(path, perms)
	if err != nil {
		t.Fatalf("Error setting mode on file %q after creation: %v", path, err)
	}
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Error stating file %q after creation: %v", path, err)
	}
	if runtime.GOOS != "windows" && stat.Mode() != perms {
		t.Fatalf("Created file %q does not have proper permissions after creation [%o != %o]",
			path, stat.Mode(), perms)
	}
}

func buildTestFolder(t *testing.T) string {
	extension := ".clconf"
	temp, err := os.MkdirTemp("", "")
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
	defer func() { _ = os.RemoveAll(temp) }()

	paths, err := findTemplates(filepath.Join(temp, subPath), extension)
	if err != nil {
		t.Fatalf("Error running findTemplates (%s): %v", message, err)
	}

	if len(expected) == 0 {
		if len(paths) != 0 {
			t.Errorf("Paths wasn't empty when it was supposed to be (%s): %v", message, paths)
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
		filepath.Join("subdir1", "subsubdir1", "yes_subdir1subsubdir1.sh.clconf"),
		filepath.Join("subdir1", "subsubdir2", "yes_subdir1subsubdir2.sh.clconf"),
		filepath.Join("subdir1", "yes_subdir1.sh.clconf"),
		filepath.Join("subdir2", "yes_subdir2.sh.clconf"),
		"yes_basedir.html.clconf",
		"yes_basedir1.html.clconf",
		"yes_basedir2.html.clconf",
	})
}

func TestFindTemplatesWithoutExtension(t *testing.T) {
	testFindTemplates(t, "Without Extension", "", "", false, []string{
		"no_basedir.html",
		"no_basedir1.html",
		filepath.Join("subdir1", "no_subdir1.sh"),
		filepath.Join("subdir1", "subsubdir1", "no_subdir1subsubdir1.sh"),
		filepath.Join("subdir1", "subsubdir1", "yes_subdir1subsubdir1.sh.clconf"),
		filepath.Join("subdir1", "subsubdir2", "no_subdir1subsubdir2.sh"),
		filepath.Join("subdir1", "subsubdir2", "yes_subdir1subsubdir2.sh.clconf"),
		filepath.Join("subdir1", "yes_subdir1.sh.clconf"),
		filepath.Join("subdir2", "no_subdir2.sh"),
		filepath.Join("subdir2", "yes_subdir2.sh.clconf"),
		"yes_basedir.html.clconf",
		"yes_basedir1.html.clconf",
		"yes_basedir2.html.clconf",
	})
}

func TestFindTemplatesSingleFileMatchingExtension(t *testing.T) {
	testFindTemplates(t, "Single File Matching Extension", ".clconf", "subdir1/subsubdir1/yes_subdir1subsubdir1.sh.clconf", true, []string{
		filepath.Join("subdir1", "subsubdir1", "yes_subdir1subsubdir1.sh.clconf"),
	})
}

func TestFindTemplatesSingleFileNotMatchingExtension(t *testing.T) {
	testFindTemplates(t, "Single File Not Matching Extension", ".clconf", "subdir1/subsubdir1/no_subdir1subsubdir1.sh", true, []string{
		filepath.Join("subdir1", "subsubdir1", "no_subdir1subsubdir1.sh"),
	})
}

func TestFindTemplatesSingleEmptyFolder(t *testing.T) {
	testFindTemplates(t, "Empty Folder", ".clconf", "emptydir", true, []string{})
}

func defaultContext() (TemplateOptions, *secret.SecretAgent) {
	return TemplateOptions{
		CopyTemplatePerms: true,
		Flatten:           false,
		KeepExistingPerms: false,
		Rm:                false,
		DirMode:           os.FileMode(0753), // We use a wonky value to look for it later
		Extension:         ".clconf",
	}, secret.NewSecretAgent([]byte(""))
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func checkFile(t *testing.T, path string, expectedPerms os.FileMode) {
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Error reading %q: %v", path, err)
	}

	stat, _ := os.Stat(path)
	if runtime.GOOS != "windows" && stat.Mode() != expectedPerms {
		t.Fatalf("File %q has mode %o that does not match expected value %o", path, stat.Mode(), expectedPerms)
	}

	if string(content) != fakeValue {
		t.Fatalf("File %q content %q does not match expected %q", path, content, fakeValue)
	}
}

func TestProcessTemplateInPlace(t *testing.T) {
	extension := ".clconf"
	temp := buildTestFolder(t)
	defer func() { _ = os.RemoveAll(temp) }()

	options, secretAgent := defaultContext()

	value := map[interface{}]interface{}{"foo": fakeValue}

	template := filepath.Join(temp, "yes_basedir.html"+extension)
	t.Run("In place, Default options", func(t *testing.T) {
		_, err := processTemplate(pathWithRelative{
			full: template,
			rel:  filepath.Base(template),
		}, "", value, secretAgent, options)
		if err != nil {
			t.Fatalf("processTemplate reported error: %v", err)
		}

		checkFile(t, filepath.Join(temp, "yes_basedir.html"), 0646)

		if !exists(template) {
			t.Errorf("Template went missing when it wasn't supposed to!")
		}
	})

	t.Run("In place, Fixed file mode", func(t *testing.T) {
		options.CopyTemplatePerms = false
		options.FileMode = 0610
		_, err := processTemplate(pathWithRelative{
			full: template,
			rel:  filepath.Base(template),
		}, "", value, secretAgent, options)
		if err != nil {
			t.Fatalf("processTemplate reported error: %v", err)
		}

		checkFile(t, filepath.Join(temp, "yes_basedir.html"), 0610)
	})

	t.Run("In place, Keep existing perms, rm template", func(t *testing.T) {
		options.Rm = true
		options.FileMode = 0601
		options.KeepExistingPerms = true
		_, err := processTemplate(pathWithRelative{
			full: template,
			rel:  filepath.Base(template),
		}, "", value, secretAgent, options)
		if err != nil {
			t.Fatalf("processTemplate reported error: %v", err)
		}

		checkFile(t, filepath.Join(temp, "yes_basedir.html"), 0610)
		if exists(template) {
			t.Errorf("Template went missing when it wasn't supposed to!")
		}
	})

	t.Run("In place, Single noext file, options with ext", func(t *testing.T) {
		template = filepath.Join(temp, "subdir1", "subsubdir2", "no_subdir1subsubdir2.sh")
		_, err := processTemplate(pathWithRelative{
			full: template,
			rel:  filepath.Base(template),
		}, "", value, secretAgent, options)
		if err != nil {
			t.Fatalf("processTemplate reported error: %v", err)
		}

		checkFile(t, filepath.Join(temp, "subdir1", "subsubdir2", "no_subdir1subsubdir2.sh"), 0777)
	})
}

func TestProcessTemplateFolder(t *testing.T) {
	extension := ".clconf"
	temp := buildTestFolder(t)
	defer func() { _ = os.RemoveAll(temp) }()

	dest := filepath.Join(temp, "dest")
	err := os.Mkdir(dest, 0750)
	if err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	options, secretAgent := defaultContext()

	value := map[interface{}]interface{}{"foo": fakeValue}

	t.Run("With dest, default options", func(t *testing.T) {
		template := filepath.Join(temp, "yes_basedir.html"+extension)
		_, err := processTemplate(pathWithRelative{
			full: template,
			rel:  filepath.Base(template),
		}, dest, value, secretAgent, options)
		if err != nil {
			t.Fatalf("processTemplate reported error: %v", err)
		}

		checkFile(t, filepath.Join(dest, "yes_basedir.html"), 0646)

		if !exists(template) {
			t.Errorf("Template went missing when it wasn't supposed to!")
		}
	})

	t.Run("With dest, rm enabled", func(t *testing.T) {
		options.Rm = true
		template := filepath.Join(temp, "subdir1", "subsubdir1", "no_subdir1subsubdir1.sh")
		_, err := processTemplate(pathWithRelative{
			full: template,
			rel:  filepath.Join("subdir1", "subsubdir1", "no_subdir1subsubdir1.sh"),
		}, dest, value, secretAgent, options)
		if err != nil {
			t.Fatalf("processTemplate reported error: %v", err)
		}

		checkFile(t, filepath.Join(dest, "subdir1", "subsubdir1", "no_subdir1subsubdir1.sh"), 0775)

		if exists(template) {
			t.Errorf("Template still exists when it wasn't supposed to!")
		}
	})
}

func TestProcessTemplateKeepEmpty(t *testing.T) {
	temp := buildTestFolder(t)
	defer func() { _ = os.RemoveAll(temp) }()

	dest := filepath.Join(temp, "dest")
	err := os.Mkdir(dest, 0750)
	if err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	options, secretAgent := defaultContext()

	value := map[interface{}]interface{}{"foo": ""}

	t.Run("Keep Empty True Rm False With Dest", func(t *testing.T) {
		subPath := filepath.Join("subdir1", "subsubdir1", "no_subdir1subsubdir1.sh")
		template := filepath.Join(temp, subPath)
		options.KeepEmpty = true
		options.Rm = false
		_, err := processTemplate(pathWithRelative{
			full: template,
			rel:  subPath,
		}, dest, value, secretAgent, options)
		if err != nil {
			t.Fatalf("processTemplate reported error: %v", err)
		}

		if !exists(filepath.Join(dest, subPath)) {
			t.Errorf("Result is missing, should be present!")
		}
		stat, _ := os.Stat(filepath.Join(dest, subPath))
		if stat.Size() != 0 {
			t.Errorf("Result does not have zero size!")
		}
	})

	t.Run("Keep Empty False Rm False With Dest", func(t *testing.T) {
		subPath := filepath.Join("subdir1", "subsubdir1", "no_subdir1subsubdir1.sh")
		template := filepath.Join(temp, subPath)
		options.KeepEmpty = false
		options.Rm = false
		_, err := processTemplate(pathWithRelative{
			full: template,
			rel:  subPath,
		}, dest, value, secretAgent, options)
		if err != nil {
			t.Fatalf("processTemplate reported error: %v", err)
		}

		if exists(filepath.Join(dest, subPath)) {
			t.Errorf("Result is present, should be missing!")
		}
	})

	t.Run("Keep Empty False Rm True With Dest", func(t *testing.T) {
		subPath := filepath.Join("subdir1", "subsubdir1", "no_subdir1subsubdir1.sh")
		template := filepath.Join(temp, subPath)
		options.KeepEmpty = false
		options.Rm = true
		_, err := processTemplate(pathWithRelative{
			full: template,
			rel:  subPath,
		}, dest, value, secretAgent, options)
		if err != nil {
			t.Fatalf("processTemplate reported error: %v", err)
		}

		if exists(filepath.Join(dest, subPath)) {
			t.Errorf("Result is present, should be missing!")
		}
		if exists(template) {
			t.Errorf("Template still exists when it wasn't supposed to!")
		}
	})

	t.Run("Keep Empty False Rm False InPlace", func(t *testing.T) {
		subPath := "no_basedir.html"
		template := filepath.Join(temp, subPath)
		options.KeepEmpty = false
		options.Rm = false // The template should still go away because we're doing it in place
		_, err := processTemplate(pathWithRelative{
			full: template,
			rel:  subPath,
		}, "", value, secretAgent, options)
		if err != nil {
			t.Fatalf("processTemplate reported error: %v", err)
		}

		if exists(filepath.Join(dest, subPath)) {
			t.Errorf("Result is present, should be missing!")
		}
		if exists(template) {
			t.Errorf("Template still exists when it wasn't supposed to!")
		}
	})

	t.Run("Keep Empty False Rm True InPlace", func(t *testing.T) {
		subPath := "no_basedir1.html"
		template := filepath.Join(temp, subPath)
		options.KeepEmpty = false
		options.Rm = true
		_, err := processTemplate(pathWithRelative{
			full: template,
			rel:  subPath,
		}, "", value, secretAgent, options)
		if err != nil {
			t.Fatalf("processTemplate reported error: %v", err)
		}

		if exists(filepath.Join(dest, subPath)) {
			t.Errorf("Result is present, should be missing!")
		}
		if exists(template) {
			t.Errorf("Template still exists when it wasn't supposed to!")
		}
	})
}

func TestProcessTemplateFolderFlatten(t *testing.T) {
	temp := buildTestFolder(t)
	defer func() { _ = os.RemoveAll(temp) }()

	dest := filepath.Join(temp, "dest")

	options, secretAgent := defaultContext()

	value := map[interface{}]interface{}{"foo": fakeValue}

	t.Run("With dest, flatten", func(t *testing.T) {
		options.Flatten = true
		template := filepath.Join(temp, "subdir1", "subsubdir1", "no_subdir1subsubdir1.sh")
		_, err := processTemplate(pathWithRelative{
			full: template,
			rel:  filepath.Join("subdir1", "subsubdir1", "no_subdir1subsubdir1.sh"),
		}, dest, value, secretAgent, options)
		if err != nil {
			t.Fatalf("processTemplate reported error: %v", err)
		}

		checkFile(t, filepath.Join(dest, "no_subdir1subsubdir1.sh"), 0775)
	})
}

func TestProcessTemplatesWithExtension(t *testing.T) {
	temp := buildTestFolder(t)
	defer func() { _ = os.RemoveAll(temp) }()

	dest := filepath.Join(temp, "dest")
	err := os.Mkdir(dest, 0750)
	if err != nil {
		t.Fatalf("unable to create temp dest dir: %v", err)
	}

	options, secretAgent := defaultContext()

	value := map[interface{}]interface{}{"foo": fakeValue}

	_, err = ProcessTemplates([]string{temp}, dest, value, secretAgent, options)
	if err != nil {
		t.Errorf("TestProcessTemplatesWithExtension: Error processing templates: %v", err)
	}
	checkFile(t, filepath.Join(dest, "yes_basedir.html"), 0646)
	checkFile(t, filepath.Join(dest, "yes_basedir1.html"), 0640)
	checkFile(t, filepath.Join(dest, "yes_basedir2.html"), 0640)
	checkFile(t, filepath.Join(dest, "subdir1", "yes_subdir1.sh"), 0770)
	checkFile(t, filepath.Join(dest, "subdir1", "subsubdir1", "yes_subdir1subsubdir1.sh"), 0775)
	checkFile(t, filepath.Join(dest, "subdir1", "subsubdir2", "yes_subdir1subsubdir2.sh"), 0777)
	checkFile(t, filepath.Join(dest, "subdir2", "yes_subdir2.sh"), 0777)

	if exists(filepath.Join(dest, "no_basedir.html")) {
		t.Errorf("File exists when it shouldn't")
	}
	if exists(filepath.Join(dest, "no_basedir1.html")) {
		t.Errorf("File exists when it shouldn't")
	}
	if exists(filepath.Join(dest, "subdir1", "no_subdir1.sh")) {
		t.Errorf("File exists when it shouldn't")
	}
	if exists(filepath.Join(dest, "subdir1", "subsubdir1", "no_subdir1subsubdir1.sh")) {
		t.Errorf("File exists when it shouldn't")
	}
	if exists(filepath.Join(dest, "subdir1", "subsubdir2", "no_subdir1subsubdir2.sh")) {
		t.Errorf("File exists when it shouldn't")
	}
	if exists(filepath.Join(dest, "subdir2", "no_subdir2.sh")) {
		t.Errorf("File exists when it shouldn't")
	}
}
