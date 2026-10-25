package main

import (
	"bytes"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestValidateTranslationFlags(t *testing.T) {
	for _, tt := range []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"empty", nil, false},
		{"short conflict", []string{"install", "-q", "-v"}, true},
		{"long conflict", []string{"install", "--quiet", "--verbose"}, true},
		{"mixed conflict", []string{"-q", "install", "--verbose"}, true},
		{"reverse mixed conflict", []string{"--verbose", "install", "-q"}, true},
		{"other mixed conflict", []string{"install", "--quiet", "-v"}, true},
		{"quiet only", []string{"install", "-q", "--quiet", "pkg"}, false},
		{"verbose only", []string{"install", "-v", "--verbose", "pkg"}, false},
		{"bundled short conflict", []string{"install", "-qv"}, true},
		{"reverse bundled short conflict", []string{"install", "-vq"}, true},
		{"multi-v and quiet conflict", []string{"install", "-vvv", "-q"}, true},
		{"multi-q and verbose conflict", []string{"install", "-qq", "-v"}, true},
		{"bundled multi-v only", []string{"install", "-vv", "pkg"}, false},
		{"bundled multi-q only", []string{"install", "-qq", "pkg"}, false},
		{"literal operands", []string{"install", "--", "-q", "-v"}, false},
		{"literal verbose", []string{"install", "-q", "--", "--verbose"}, false},
		{"conflict before separator", []string{"install", "-q", "-v", "--"}, true},
		{"option values and URL", []string{"install", "--index-url", "https://example.org/-q/-v", "--config-settings=--quiet", "pkg"}, false},
		{"attached requirement option with q and v in filename", []string{"install", "-rrequirements-dev.txt"}, false},
		{"attached constraint option with q and v in filename", []string{"install", "-cconstraints-quick-view.txt"}, false},
		{"attached editable option with q and v in path", []string{"install", "-e./pkg-quick-view"}, false},
		{"attached find-links option with q and v", []string{"install", "-fhttps://repo.org/q-and-v"}, false},
		{"attached trusted-host option with q and v", []string{"install", "-thost-with-q-v"}, false},
		{"attached build-dir option with q and v", []string{"install", "-bbuild-dir-qv"}, false},
		{"attached extra-index option with q and v", []string{"install", "-iextra-qv"}, false},
		{"non-alpha short option with digits", []string{"install", "-1", "-2"}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTranslationFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, want rejection %v", err, tt.wantErr)
			}
			if err != nil && (!errors.Is(err, errUsage) || !errors.Is(err, ErrInvalidArgs)) {
				t.Fatalf("lost typed usage error: %v", err)
			}
		})
	}
}

func TestCLIRejectsTranslationFlagConflicts(t *testing.T) {
	// A missing uv proves validation happens before executable discovery.
	t.Setenv("UVPIP_UV", filepath.Join(t.TempDir(), "missing"))
	for _, flags := range [][]string{{"-q", "-v"}, {"--quiet", "--verbose"}, {"-v", "--quiet"}} {
		t.Run(strings.Join(flags, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			args := append([]string{"install", "private-package-value"}, flags...)
			if code := runCLI(args, nil, &stdout, &stderr); code != 2 {
				t.Fatalf("exit = %d, want 2: %s", code, &stderr)
			}
			if stdout.Len() != 0 || !strings.Contains(stderr.String(), "cannot be combined") || strings.Contains(stderr.String(), "private-package-value") {
				t.Fatalf("unexpected output: stdout=%q stderr=%q", &stdout, &stderr)
			}
		})
	}
}

func TestTranslateToUV(t *testing.T) {
	tests := []struct {
		name     string
		in, want []string
	}{
		{"empty", nil, []string{"pip"}},
		{"install", []string{"install", "requests"}, []string{"pip", "install", "requests"}},
		{"requirements", []string{"install", "-r", "a b.txt"}, []string{"pip", "install", "-r", "a b.txt"}},
		{"editable", []string{"install", "-e", "."}, []string{"pip", "install", "-e", "."}},
		{"upgrade flag", []string{"install", "--upgrade", "pkg"}, []string{"pip", "install", "--upgrade", "pkg"}},
		{"upgrade", []string{"upgrade", "pkg", "--pre"}, []string{"pip", "install", "--upgrade", "pkg", "--pre"}},
		{"uninstall", []string{"uninstall", "pkg"}, []string{"pip", "uninstall", "pkg"}},
		{"confirmations", []string{"uninstall", "-y", "a", "--yes", "b", "-y"}, []string{"pip", "uninstall", "a", "b"}},
		{"separator", []string{"uninstall", "--yes", "--", "-y", "--yes"}, []string{"pip", "uninstall", "--", "-y", "--yes"}},
		{"similar flags", []string{"uninstall", "--yes=no", "-yy", "x-y", ""}, []string{"pip", "uninstall", "--yes=no", "-yy", "x-y", ""}},
		{"only flag", []string{"uninstall", "-y"}, []string{"pip", "uninstall"}},
		{"remove alias", []string{"REMOVE", "--yes", "pkg"}, []string{"pip", "uninstall", "pkg"}},
		{"list", []string{"list", "--format=json"}, []string{"pip", "list", "--format=json"}},
		{"show", []string{"show", "a", "b"}, []string{"pip", "show", "a", "b"}},
		{"freeze", []string{"freeze"}, []string{"pip", "freeze"}},
		{"check", []string{"check"}, []string{"pip", "check"}},
		{"cache dir", []string{"cache", "dir"}, []string{"cache", "dir"}},
		{"cache list", []string{"cache", "list", "*"}, []string{"cache", "list", "*"}},
		{"cache empty", []string{"cache"}, []string{"cache"}},
		{"unknown", []string{"download", "-y", "pkg"}, []string{"pip", "download", "-y", "pkg"}},
		{"global flags", []string{"-v", "install", "pkg"}, []string{"pip", "-v", "install", "pkg"}},
		{"help", []string{"--help"}, []string{"pip", "--help"}},
		{"case preserved", []string{"INSTALL", "Pkg"}, []string{"pip", "INSTALL", "Pkg"}},
		{"literal args", []string{"install", "a;b", "$(cmd)", "a\"b", "x y", ""}, []string{"pip", "install", "a;b", "$(cmd)", "a\"b", "x y", ""}},
		{"wheel", []string{"wheel", "pkg"}, []string{"pip", "wheel", "pkg"}},
		{"debug", []string{"debug"}, []string{"pip", "debug"}},
		{"quiet flag", []string{"install", "-q", "pkg"}, []string{"pip", "install", "-q", "pkg"}},
		{"isolated flag", []string{"install", "--isolated", "pkg"}, []string{"pip", "install", "--isolated", "pkg"}},
		{"retries flag", []string{"install", "--retries", "5", "pkg"}, []string{"pip", "install", "--retries", "5", "pkg"}},
		{"no-input flag", []string{"install", "--no-input", "pkg"}, []string{"pip", "install", "--no-input", "pkg"}},
		{"dry-run flag", []string{"install", "--dry-run", "pkg"}, []string{"pip", "install", "--dry-run", "pkg"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := append([]string(nil), tt.in...)
			got := translateToUV(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %#v, want %#v", got, tt.want)
			}
			got[0] = "mutated"
			if !reflect.DeepEqual(tt.in, before) {
				t.Fatal("translation aliases or mutates input")
			}
		})
	}
}
