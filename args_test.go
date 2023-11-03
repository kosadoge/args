package args

import (
	"testing"
)

func TestParseName_Success(t *testing.T) {
	testcases := []struct {
		input string
		name  string
		alias string
	}{
		{"mode", "mode", ""},
		{"m", "", "m"},
		{"mode,m", "mode", "m"},
		{"m,mode", "mode", "m"},
	}

	for _, tc := range testcases {
		t.Run(tc.input, func(t *testing.T) {
			name, alias, err := parseName(tc.input)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if name != tc.name {
				t.Fatalf("name not equals [expect=%s, actual=%s]", tc.name, name)
			}
			if alias != tc.alias {
				t.Fatalf("alias not equals [expect=%s, actual=%s]", tc.alias, alias)
			}
		})
	}
}

func TestParseName_Failure(t *testing.T) {
	testcases := []string{
		"-mode",
		"m ode",
		"m=ode",
		"mode,m,apple",
		"m,m",
		"mode,-m",
		"mode,      m",
	}

	for _, tc := range testcases {
		t.Run(tc, func(t *testing.T) {
			name, alias, err := parseName(tc)
			if err == nil {
				t.Fatalf("error is nil [name=%s, alias=%s]", name, alias)
			}
		})
	}
}
