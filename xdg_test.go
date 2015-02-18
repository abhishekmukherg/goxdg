package xdg

import "testing"

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
