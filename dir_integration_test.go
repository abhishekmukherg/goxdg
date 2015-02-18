package xdg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/check.v1"
)

type TempDirSuite struct {
	tempDir                            string
	m                                  mockEnv
	fileInHome, fileInLocal, fileInLib *os.File
}

func Test(t *testing.T) { check.TestingT(t) }

var _ = check.Suite(&TempDirSuite{})

func (s *TempDirSuite) SetUpTest(c *check.C) {
	s.tempDir = c.MkDir()

	dataDirs := []string{filepath.Join(s.tempDir, "local"), filepath.Join(s.tempDir, "lib")}
	s.m = mockEnv{
		"XDG_DATA_HOME": filepath.Join(s.tempDir, "home"),
		"XDG_DATA_DIRS": strings.Join(dataDirs, ":"),
	}
	os.Mkdir(s.m["XDG_DATA_HOME"], 0700)
	for _, d := range dataDirs {
		os.Mkdir(d, 0700)
	}

	var err error
	s.fileInHome, err = os.Create(filepath.Join(s.m["XDG_DATA_HOME"], "fileInHome"))
	if err != nil {
		c.Fatal("Failed to create test files:", err)
	}

	s.fileInLocal, err = os.Create(filepath.Join(dataDirs[0], "fileInLocal"))
	if err != nil {
		c.Fatal("Failed to create test files:", err)
	}

	s.fileInLib, err = os.Create(filepath.Join(dataDirs[0], "fileInLib"))
	if err != nil {
		c.Fatal("Failed to create test files:", err)
	}

	for _, f := range []*os.File{s.fileInHome, s.fileInLocal, s.fileInLib} {
		err = f.Close()
		if err != nil {
			c.Fatal("Failed to create test files:", err)
		}
	}

}

func (s *TempDirSuite) TestOpen(c *check.C) {
	readEnv := Data.fromEnvironment(s.m.Getenv)
	runTest := func(filename string, expected *os.File) {
		openedFile, err := readEnv.Open(filename)
		if err != nil {
			c.Errorf("Open(%q) = err(%q) but was expecting %q", filename, err, expected)
			return
		}
		if openedFile.Name() != expected.Name() {
			c.Errorf("Open(%q) = %q but was expected %q", filename, openedFile.Name(), expected.Name())
		}
	}
	runTest("fileInHome", s.fileInHome)
	runTest("fileInLocal", s.fileInLocal)
	runTest("fileInLib", s.fileInLib)
	openedFile, err := readEnv.Open("doesNotExist")
	if err == nil {
		c.Errorf("Found a file that doesn't exist?", openedFile.Name())
	}
}

func (s *TempDirSuite) TestMkdir(c *check.C) {
	readEnv := Data.fromEnvironment(s.m.Getenv)
	dirname, err := readEnv.Mkdir("asdf", 0700)

	c.Assert(err, check.IsNil) // Should be able to readEnv.Mkdir("asdf")
	if dirname != filepath.Join(s.m["XDG_DATA_HOME"], "asdf") {
		c.Error("Apparently made the wrong dir:", dirname)
	}
}

func (s *TempDirSuite) TestMkdirBlank(c *check.C) {
	readEnv := Data.fromEnvironment(s.m.Getenv)
	dirname, err := readEnv.Mkdir("", 0700)
	// Should fail to create directory with Mkdir("")
	c.Assert(err, check.FitsTypeOf, &os.PathError{})
	// Should say no such file or directory when name blank to match os.Mkdir
	c.Assert(err, check.ErrorMatches, "mkdir : no such file or directory")
	c.Assert(dirname, check.Equals, "")

	dirname, err = readEnv.MkdirAll("", 0700)
	// Should fail to create directory with MkdirAll("")
	c.Assert(err, check.FitsTypeOf, &os.PathError{})
	// Should say no such file or directory when name blank to match os.MkdirAll
	c.Assert(err, check.ErrorMatches, "mkdir : no such file or directory")
	c.Assert(dirname, check.Equals, "")
}

func (s *TempDirSuite) TestMkdirAll(c *check.C) {
	readEnv := Data.fromEnvironment(s.m.Getenv)
	dirname, err := readEnv.MkdirAll("asdf/qwerty", 0700)

	c.Assert(err, check.IsNil) // Failed to create dir with MkdirAll

	// Check if we created the correct directory
	c.Assert(dirname, check.Equals, filepath.Join(s.m["XDG_DATA_HOME"], "asdf/qwerty"))
}
