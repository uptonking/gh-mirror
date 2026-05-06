package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigSupportsMixedRepoEntries(t *testing.T) {
	data := []byte(`{
		"sleepDuration": 1,
		"repos": [
			"owner/repo",
			"ssh://git@code.haverbeke.berlin/codemirror/merge",
			"https://code.haverbeke.berlin/codemirror/merge.git",
			{"url": "ssh://git@code.haverbeke.berlin/codemirror/merge.git", "path": "custom/codemirror-merge"}
		]
	}`)

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if got, want := len(cfg.Repos), 4; got != want {
		t.Fatalf("len(cfg.Repos) = %d, want %d", got, want)
	}
	if got, want := cfg.Repos[0].Repo, "owner/repo"; got != want {
		t.Fatalf("cfg.Repos[0].Repo = %q, want %q", got, want)
	}
	if got, want := cfg.Repos[1].URL, "ssh://git@code.haverbeke.berlin/codemirror/merge"; got != want {
		t.Fatalf("cfg.Repos[1].URL = %q, want %q", got, want)
	}
	if got, want := cfg.Repos[2].URL, "https://code.haverbeke.berlin/codemirror/merge.git"; got != want {
		t.Fatalf("cfg.Repos[2].URL = %q, want %q", got, want)
	}
	if got, want := cfg.Repos[3].Path, "custom/codemirror-merge"; got != want {
		t.Fatalf("cfg.Repos[3].Path = %q, want %q", got, want)
	}
}

func TestRepoEntryMarshalKeepsLegacyString(t *testing.T) {
	data, err := json.Marshal(RepoEntry{Repo: "owner/repo"})
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	if got, want := string(data), `"owner/repo"`; got != want {
		t.Fatalf("json.Marshal = %s, want %s", got, want)
	}
}

func TestRepoEntryMarshalKeepsStringURL(t *testing.T) {
	data, err := json.Marshal(RepoEntry{URL: "https://code.haverbeke.berlin/codemirror/merge", serializedAsString: true})
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	if got, want := string(data), `"https://code.haverbeke.berlin/codemirror/merge"`; got != want {
		t.Fatalf("json.Marshal = %s, want %s", got, want)
	}
}

func TestRepoEntryMarshalKeepsObjectWithPath(t *testing.T) {
	data, err := json.Marshal(RepoEntry{URL: "https://code.haverbeke.berlin/codemirror/merge", Path: "custom/codemirror-merge"})
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	if got, want := string(data), `{"url":"https://code.haverbeke.berlin/codemirror/merge","path":"custom/codemirror-merge"}`; got != want {
		t.Fatalf("json.Marshal = %s, want %s", got, want)
	}
}

func TestResolveGitHubShorthand(t *testing.T) {
	repo, err := (RepoEntry{Repo: "owner/repo"}).Resolve()
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	if got, want := repo.CloneURL, "git@github.com:owner/repo.git"; got != want {
		t.Fatalf("CloneURL = %q, want %q", got, want)
	}
	if got, want := repo.DisplayPath(), "owner/repo"; got != want {
		t.Fatalf("DisplayPath = %q, want %q", got, want)
	}
}

func TestResolveCustomCloneURLs(t *testing.T) {
	tests := []struct {
		name          string
		repo          RepoEntry
		wantPath      string
		wantRemoteKey string
	}{
		{
			name:          "scp style ssh",
			repo:          RepoEntry{URL: "git@gitlab.com:fdroid/fdroidclient.git"},
			wantPath:      "fdroid/fdroidclient",
			wantRemoteKey: "gitlab.com|fdroid/fdroidclient",
		},
		{
			name:          "ssh scheme with git suffix",
			repo:          RepoEntry{URL: "ssh://git@code.haverbeke.berlin/codemirror/merge.git"},
			wantPath:      "codemirror/merge",
			wantRemoteKey: "code.haverbeke.berlin|codemirror/merge",
		},
		{
			name:          "ssh scheme without git suffix",
			repo:          RepoEntry{URL: "ssh://git@code.haverbeke.berlin/codemirror/merge"},
			wantPath:      "codemirror/merge",
			wantRemoteKey: "code.haverbeke.berlin|codemirror/merge",
		},
		{
			name:          "https with git suffix",
			repo:          RepoEntry{URL: "https://code.haverbeke.berlin/codemirror/merge.git"},
			wantPath:      "codemirror/merge",
			wantRemoteKey: "code.haverbeke.berlin|codemirror/merge",
		},
		{
			name:          "https without git suffix",
			repo:          RepoEntry{URL: "https://code.haverbeke.berlin/codemirror/merge"},
			wantPath:      "codemirror/merge",
			wantRemoteKey: "code.haverbeke.berlin|codemirror/merge",
		},
		{
			name:          "prosemirror repo path",
			repo:          RepoEntry{URL: "ssh://git@code.haverbeke.berlin/prosemirror/prosemirror-view.git"},
			wantPath:      "prosemirror/prosemirror-view",
			wantRemoteKey: "code.haverbeke.berlin|prosemirror/prosemirror-view",
		},
		{
			name: "path override",
			repo: RepoEntry{
				URL:  "ssh://git@code.haverbeke.berlin/codemirror/merge.git",
				Path: "custom/codemirror-merge",
			},
			wantPath:      "custom/codemirror-merge",
			wantRemoteKey: "code.haverbeke.berlin|codemirror/merge",
		},
		{
			name:          "distinct ssh user",
			repo:          RepoEntry{URL: "ssh://alice@code.haverbeke.berlin/codemirror/merge"},
			wantPath:      "codemirror/merge",
			wantRemoteKey: "code.haverbeke.berlin|codemirror/merge|user=alice",
		},
		{
			name:          "distinct ssh port",
			repo:          RepoEntry{URL: "ssh://git@code.haverbeke.berlin:2222/codemirror/merge"},
			wantPath:      "codemirror/merge",
			wantRemoteKey: "code.haverbeke.berlin|codemirror/merge|port=2222",
		},
		{
			name:          "default https port ignored",
			repo:          RepoEntry{URL: "https://code.haverbeke.berlin:443/codemirror/merge"},
			wantPath:      "codemirror/merge",
			wantRemoteKey: "code.haverbeke.berlin|codemirror/merge",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolved, err := test.repo.Resolve()
			if err != nil {
				t.Fatalf("Resolve returned error: %v", err)
			}
			if got := resolved.DisplayPath(); got != test.wantPath {
				t.Fatalf("DisplayPath = %q, want %q", got, test.wantPath)
			}
			if got := resolved.RemoteKey; got != test.wantRemoteKey {
				t.Fatalf("RemoteKey = %q, want %q", got, test.wantRemoteKey)
			}
		})
	}
}

func TestResolveSSHAndHTTPSShareDefaultIdentity(t *testing.T) {
	sshRepo, err := (RepoEntry{URL: "ssh://git@code.haverbeke.berlin/codemirror/merge.git"}).Resolve()
	if err != nil {
		t.Fatalf("Resolve ssh returned error: %v", err)
	}
	httpsRepo, err := (RepoEntry{URL: "https://code.haverbeke.berlin/codemirror/merge"}).Resolve()
	if err != nil {
		t.Fatalf("Resolve https returned error: %v", err)
	}

	if sshRepo.RemoteKey != httpsRepo.RemoteKey {
		t.Fatalf("RemoteKey mismatch: %q vs %q", sshRepo.RemoteKey, httpsRepo.RemoteKey)
	}
	if sshRepo.DisplayPath() != httpsRepo.DisplayPath() {
		t.Fatalf("DisplayPath mismatch: %q vs %q", sshRepo.DisplayPath(), httpsRepo.DisplayPath())
	}
}

func TestResolveDifferentHostsSharePathButNotIdentity(t *testing.T) {
	sshRepo, err := (RepoEntry{URL: "ssh://git@code.haverbeke.berlin/codemirror/merge.git"}).Resolve()
	if err != nil {
		t.Fatalf("Resolve ssh returned error: %v", err)
	}
	otherRepo, err := (RepoEntry{URL: "https://gitlab.com/codemirror/merge.git"}).Resolve()
	if err != nil {
		t.Fatalf("Resolve other returned error: %v", err)
	}

	if sshRepo.DisplayPath() != otherRepo.DisplayPath() {
		t.Fatalf("DisplayPath mismatch: %q vs %q", sshRepo.DisplayPath(), otherRepo.DisplayPath())
	}
	if sshRepo.RemoteKey == otherRepo.RemoteKey {
		t.Fatalf("RemoteKey unexpectedly matched: %q", sshRepo.RemoteKey)
	}
}

func TestBundledConfigExamplesParse(t *testing.T) {
	files, err := filepath.Glob("gh-mirror-config-examples/*.json")
	if err != nil {
		t.Fatalf("filepath.Glob returned error: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected config examples")
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("os.ReadFile returned error: %v", err)
			}

			var cfg Config
			if err := json.Unmarshal(data, &cfg); err != nil {
				t.Fatalf("json.Unmarshal returned error: %v", err)
			}

			for _, repo := range cfg.Repos {
				if _, err := repo.Resolve(); err != nil {
					t.Fatalf("Resolve(%q) returned error: %v", repo.Display(), err)
				}
			}
		})
	}
}

func TestResolveRejectsInvalidEntries(t *testing.T) {
	tests := []RepoEntry{
		{URL: ""},
		{URL: "not-a-valid-clone-url"},
		{URL: "git@gitlab.com:fdroid/fdroidclient.git", Path: "../outside"},
		{URL: "git@gitlab.com:fdroid/fdroidclient.git", Path: "custom/../outside"},
		{URL: "ssh://git@code.haverbeke.berlin"},
	}

	for _, test := range tests {
		if _, err := test.Resolve(); err == nil {
			t.Fatalf("Resolve(%#v) expected error", test)
		}
	}
}
