package xdg

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockEnv map[string]string

func (m mockEnv) Getenv(key string) string { return m[key] }

func TestReadDataEnvironment(t *testing.T) {
	tests := []struct {
		home, dataHome, dataDirs string
		expected                 []string
	}{
		{expected: []string{"/usr/local/share", "/usr/share"}},
		{home: "/tmp/magic", expected: []string{"/tmp/magic/.local/share", "/usr/local/share", "/usr/share"}},
		{dataHome: "/tmp/magic", expected: []string{"/tmp/magic", "/usr/local/share", "/usr/share"}},
		{dataDirs: "/tmp/magic", expected: []string{"/tmp/magic"}},
		{home: "/tmp/home", dataDirs: "/tmp/magic", expected: []string{"/tmp/home/.local/share", "/tmp/magic"}},
		{
			home:     "/tmp/home",
			dataHome: "/tmp/dataHome",
			expected: []string{"/tmp/dataHome", "/usr/local/share", "/usr/share"},
		},
		{dataHome: "/tmp/dataHome", dataDirs: "/tmp/dataDir", expected: []string{"/tmp/dataHome", "/tmp/dataDir"}},
		{dataDirs: "/tmp/dataDir1:/tmp/dataDir2", expected: []string{"/tmp/dataDir1", "/tmp/dataDir2"}},
	}

	for _, data := range tests {
		m := mockEnv{}
		if data.home != "" {
			m["HOME"] = data.home
		}
		if data.dataHome != "" {
			m["XDG_DATA_HOME"] = data.dataHome
		}
		if data.dataDirs != "" {
			m["XDG_DATA_DIRS"] = data.dataDirs
		}
		readEnv := Data.fromEnvironment(m.Getenv)
		if len(readEnv) != len(data.expected) {
			t.Errorf("Expected Data.fromEnvironment(%q) = %q to equal %q", m, readEnv, data.expected)
		}
		for i, exp := range data.expected {
			if readEnv[i] != exp {
				t.Errorf("Expected Data.fromEnvironment(%q) = %q to equal %q", m, readEnv, data.expected)
			}
		}
	}
}

func TestOpen(t *testing.T) {

	tempDir, err := ioutil.TempDir(os.TempDir(), "goxdgtest")
	if err != nil {
		t.Skip("Failed to create temporary directory for test:", err)
	}
	defer os.RemoveAll(tempDir)

	dataDirs := []string{filepath.Join(tempDir, "local"), filepath.Join(tempDir, "lib")}
	m := mockEnv{
		"XDG_DATA_HOME": filepath.Join(tempDir, "home"),
		"XDG_DATA_DIRS": strings.Join(dataDirs, ":"),
	}
	os.Mkdir(m["XDG_DATA_HOME"], 0700)
	for _, d := range dataDirs {
		os.Mkdir(d, 0700)
	}

	fileInHome, err := os.Create(filepath.Join(m["XDG_DATA_HOME"], "fileInHome"))
	if err != nil {
		t.Fatal("Failed to create test files:", err)
	}

	fileInLocal, err := os.Create(filepath.Join(dataDirs[0], "fileInLocal"))
	if err != nil {
		t.Fatal("Failed to create test files:", err)
	}

	fileInLib, err := os.Create(filepath.Join(dataDirs[0], "fileInLib"))
	if err != nil {
		t.Fatal("Failed to create test files:", err)
	}

	for _, f := range []*os.File{fileInHome, fileInLocal, fileInLib} {
		err = f.Close()
		if err != nil {
			t.Fatal("Failed to create test files:", err)
		}
	}
	readEnv := Data.fromEnvironment(m.Getenv)
	runTest := func(filename string, expected *os.File) {
		openedFile, err := readEnv.Open(filename)
		if err != nil {
			t.Errorf("Open(%q) = %q but was expecting %q", filename, err, expected)
			return
		}
		if openedFile.Name() != expected.Name() {
			t.Errorf("Open(%q) = %q but was expected %q", filename, openedFile.Name(), expected.Name())
		}
	}
	runTest("fileInHome", fileInHome)
	runTest("fileInLocal", fileInLocal)
	runTest("fileInLib", fileInLib)
	openedFile, err := readEnv.Open("doesNotExist")
	if err == nil {
		t.Errorf("Found a file that doesn't exist?", openedFile.Name())
	}
}

func TestMkdir(t *testing.T) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "goxdgtest")
	if err != nil {
		t.Skip("Failed to create temporary directory for test:", err)
	}
	defer os.RemoveAll(tempDir)

	dataDirs := []string{filepath.Join(tempDir, "local"), filepath.Join(tempDir, "lib")}
	m := mockEnv{
		"XDG_DATA_HOME": filepath.Join(tempDir, "home"),
		"XDG_DATA_DIRS": strings.Join(dataDirs, ":"),
	}
	os.Mkdir(m["XDG_DATA_HOME"], 0700)
	for _, d := range dataDirs {
		os.Mkdir(d, 0700)
	}

	readEnv := Data.fromEnvironment(m.Getenv)
	dirname, err := readEnv.Mkdir("asdf", 0700)
	if err != nil {
		t.Error("Failed to create directory", err)
	}
	if dirname != filepath.Join(m["XDG_DATA_HOME"], "asdf") {
		t.Errorf("Apparently made the wrong dir:", dirname)
	}
}

func TestMkdirAll(t *testing.T) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "goxdgtest")
	if err != nil {
		t.Skip("Failed to create temporary directory for test:", err)
	}
	defer os.RemoveAll(tempDir)

	dataDirs := []string{filepath.Join(tempDir, "local"), filepath.Join(tempDir, "lib")}
	m := mockEnv{
		"XDG_DATA_HOME": filepath.Join(tempDir, "home"),
		"XDG_DATA_DIRS": strings.Join(dataDirs, ":"),
	}
	os.Mkdir(m["XDG_DATA_HOME"], 0700)
	for _, d := range dataDirs {
		os.Mkdir(d, 0700)
	}

	readEnv := Data.fromEnvironment(m.Getenv)
	dirname, err := readEnv.MkdirAll("asdf/qwerty", 0700)
	if err != nil {
		t.Error("Failed to create directory", err)
	}
	if dirname != filepath.Join(m["XDG_DATA_HOME"], "asdf/qwerty") {
		t.Errorf("Apparently made the wrong dir:", dirname)
	}
}
