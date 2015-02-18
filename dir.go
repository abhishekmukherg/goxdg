// Implementation of file handling as per the XDG specification
//
// This is a simple (mostly experimentation) for me to play with go. This
// package will ideally help file handling as per the XDG specification.
// This specification allows the user to override where an application
// will store data files and such
package xdg

import (
	"os"
	"path/filepath"
	"strings"
)

type XdgDirectory struct {
	userEnvVar, systemEnvVar   string
	userDefault, systemDefault string
}

type DirPath []string

var (
	Data XdgDirectory = XdgDirectory{
		userEnvVar:    "XDG_DATA_HOME",
		systemEnvVar:  "XDG_DATA_DIRS",
		userDefault:   "$HOME/.local/share",
		systemDefault: "/usr/local/share:/usr/share"}
	Config = XdgDirectory{
		userEnvVar:    "XDG_CONFIG_HOME",
		systemEnvVar:  "XDG_CONFIG_DIRS",
		userDefault:   "$HOME/.config",
		systemDefault: "/etc/xdg"}
)

const homeEnvVar = "HOME"

type EnvGetter func(string) string

func (dir *XdgDirectory) fromEnvironment(env EnvGetter) DirPath {
	results := make([]string, 0, 3)
	userDir := env(dir.userEnvVar)
	if userDir != "" {
		results = append(results, userDir)
	} else {
		home := env(homeEnvVar)
		// This is a hack, we're guessing that HOME is the only one that's needed
		if home != "" {
			results = append(results, os.Expand(dir.userDefault, env))
		}
	}

	dataDirs := env(dir.systemEnvVar)
	if dataDirs == "" {
		dataDirs = os.Expand(dir.systemDefault, env)
	}
	results = append(results, strings.Split(dataDirs, ":")...)

	return DirPath(results)
}

func (dp DirPath) Open(name string) (file *os.File, err error) {
	for _, path := range dp {
		fullFilename := filepath.Join(path, name)
		file, err = os.Open(fullFilename)
		if err == nil || !os.IsNotExist(err) {
			return
		}
	}
	return nil, os.ErrNotExist
}

func (dp DirPath) runFirstFilename(name string, action func(string) interface{}) (string, interface{}) {
	if len(dp) == 0 {
		return "", nil
	}
	dir := dp[0]
	fullName := filepath.Join(dir, name)
	return fullName, action(fullName)
}

func (dp DirPath) MkdirAll(name string, perm os.FileMode) (dirname string, err error) {
	if name == "" {
		return "", os.MkdirAll("", perm)
	}
	dirname, errInt := dp.runFirstFilename(name, func(fullName string) interface{} {
		return os.MkdirAll(fullName, perm)
	})
	if errInt != nil {
		err = errInt.(error)
	}
	return
}

func (dp DirPath) Mkdir(name string, perm os.FileMode) (dirname string, err error) {
	if name == "" {
		return "", os.Mkdir("", perm)
	}
	dirname, errInt := dp.runFirstFilename(name, func(fullName string) interface{} {
		return os.Mkdir(fullName, perm)
	})
	if errInt != nil {
		err = errInt.(error)
	}
	return
}

func (dp DirPath) Create(name string) (file *os.File, err error) {
	parent := filepath.Dir(name)
	if parent != "." {
		_, err := dp.MkdirAll(name, 0700)
		if err != nil && !os.IsExist(err) {
			return nil, err
		}
	}
	type retType struct {
		file *os.File
		err  error
	}
	_, ret := dp.runFirstFilename(name, func(fullName string) interface{} {
		file, err := os.Create(fullName)
		return retType{file: file, err: err}
	})
	retCast := ret.(retType)
	return retCast.file, retCast.err
}
