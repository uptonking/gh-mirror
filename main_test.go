package main

import "testing"

func TestValidRepoPath(t *testing.T) {
	var tests = []struct {
		repo  string
		valid bool
	}{
		{"", false},
		{"ntns", false},
		{"ntns/gh-mirror", true},
		{"NTNS/GH-mirror", true},
		{"ntns/gh-mirror-0", true},
		{"ntns/gh!mirror", false},
		{"ntns!/gh-mirror", false},
		{"ntns/gh_mirror", true},
		{"ntns/gh.mirror", true},
	}

	for _, test := range tests {
		got := isValidRepoPath(test.repo)
		want := test.valid
		if got != want {
			t.Errorf("repo %v, got %v, wanted %v", test.repo, got, want)
		}
	}
}

func TestParseRepoArg(t *testing.T) {
	tests := []struct {
		input string
		want  RepoEntry
	}{
		{"ntns/gh-mirror", RepoEntry{Repo: "ntns/gh-mirror"}},
		{"git@gitlab.com:fdroid/fdroidclient.git", RepoEntry{URL: "git@gitlab.com:fdroid/fdroidclient.git", serializedAsString: true}},
		{"ssh://git@code.haverbeke.berlin/codemirror/merge.git", RepoEntry{URL: "ssh://git@code.haverbeke.berlin/codemirror/merge.git", serializedAsString: true}},
		{"ssh://git@code.haverbeke.berlin/codemirror/merge", RepoEntry{URL: "ssh://git@code.haverbeke.berlin/codemirror/merge", serializedAsString: true}},
		{"https://code.haverbeke.berlin/codemirror/merge", RepoEntry{URL: "https://code.haverbeke.berlin/codemirror/merge", serializedAsString: true}},
	}

	for _, test := range tests {
		got, err := parseRepoArg(test.input)
		if err != nil {
			t.Fatalf("parseRepoArg(%q) returned error: %v", test.input, err)
		}
		if got != test.want {
			t.Fatalf("parseRepoArg(%q) = %#v, want %#v", test.input, got, test.want)
		}
	}
}
