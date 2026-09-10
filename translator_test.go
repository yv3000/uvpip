package main

import (
	"reflect"
	"testing"
)

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
