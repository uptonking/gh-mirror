package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

type RepoEntry struct {
	Repo string
	URL  string `json:"url,omitempty"`
	Path string `json:"path,omitempty"`

	serializedAsString bool
}

type repoEntryJSON struct {
	URL  string `json:"url"`
	Path string `json:"path,omitempty"`
}

type ResolvedRepo struct {
	CloneURL  string
	LocalPath string
	RemoteKey string
}

type cloneTarget struct {
	Scheme   string
	User     string
	Host     string
	Port     string
	RepoPath string
}

var reSCPStyleCloneURL = regexp.MustCompile(`^(?:([^@/:]+)@)?([^:/]+):/?(.+)$`)

func (r RepoEntry) IsGitHubShorthand() bool {
	return r.Repo != ""
}

func (r *RepoEntry) UnmarshalJSON(data []byte) error {
	var shorthand string
	if err := json.Unmarshal(data, &shorthand); err == nil {
		shorthand = strings.TrimSpace(shorthand)
		r.Repo = ""
		r.URL = ""
		r.Path = ""
		r.serializedAsString = true

		switch {
		case isValidRepoPath(shorthand):
			r.Repo = shorthand
		case shorthand != "":
			if _, err := parseCloneURL(shorthand); err == nil {
				r.URL = shorthand
				return nil
			}
			r.Repo = shorthand
		default:
			r.Repo = shorthand
		}

		return nil
	}

	var raw repoEntryJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	r.Repo = ""
	r.URL = strings.TrimSpace(raw.URL)
	r.Path = strings.TrimSpace(raw.Path)
	r.serializedAsString = false
	return nil
}

func (r RepoEntry) MarshalJSON() ([]byte, error) {
	if r.IsGitHubShorthand() {
		return json.Marshal(r.Repo)
	}
	if r.Path == "" && r.serializedAsString {
		return json.Marshal(r.URL)
	}

	return json.Marshal(repoEntryJSON{
		URL:  r.URL,
		Path: r.Path,
	})
}

func (r RepoEntry) Resolve() (ResolvedRepo, error) {
	if r.IsGitHubShorthand() {
		if !isValidRepoPath(r.Repo) {
			return ResolvedRepo{}, fmt.Errorf("repo path is invalid: %v", r.Repo)
		}

		target := cloneTarget{
			Scheme:   "ssh",
			User:     "git",
			Host:     "github.com",
			RepoPath: r.Repo,
		}
		return ResolvedRepo{
			CloneURL:  fmt.Sprintf("git@github.com:%s.git", r.Repo),
			LocalPath: filepath.Clean(r.Repo),
			RemoteKey: target.IdentityKey(),
		}, nil
	}

	if r.URL == "" {
		return ResolvedRepo{}, fmt.Errorf("clone url is required")
	}

	target, err := parseCloneURL(r.URL)
	if err != nil {
		return ResolvedRepo{}, err
	}

	localPath := strings.TrimSpace(r.Path)
	if localPath == "" {
		localPath = target.DeriveLocalPath()
	}
	if err := validateRelativePath(localPath); err != nil {
		return ResolvedRepo{}, err
	}

	return ResolvedRepo{
		CloneURL:  r.URL,
		LocalPath: filepath.Clean(localPath),
		RemoteKey: target.IdentityKey(),
	}, nil
}

func (r RepoEntry) Display() string {
	if r.IsGitHubShorthand() {
		return r.Repo
	}

	resolved, err := r.Resolve()
	if err != nil {
		if r.Path != "" {
			return fmt.Sprintf("%s <- %s", toDisplayPath(r.Path), r.URL)
		}
		return r.URL
	}

	return fmt.Sprintf("%s <- %s", resolved.DisplayPath(), resolved.CloneURL)
}

func (r RepoEntry) SortKey() string {
	if resolved, err := r.Resolve(); err == nil {
		return resolved.DisplayPath()
	}
	if r.IsGitHubShorthand() {
		return r.Repo
	}
	if r.Path != "" {
		return toDisplayPath(r.Path)
	}
	return r.URL
}

func (r ResolvedRepo) DisplayPath() string {
	return toDisplayPath(r.LocalPath)
}

func parseRepoArg(input string) (RepoEntry, error) {
	input = strings.TrimSpace(input)
	if isValidRepoPath(input) {
		return RepoEntry{Repo: input}, nil
	}
	if _, err := parseCloneURL(input); err == nil {
		return RepoEntry{URL: input, serializedAsString: true}, nil
	}
	return RepoEntry{}, fmt.Errorf("repo must be username/repo or a valid clone url: %v", input)
}

func parseCloneURL(raw string) (cloneTarget, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return cloneTarget{}, fmt.Errorf("clone url is empty")
	}

	if !strings.Contains(raw, "://") {
		matches := reSCPStyleCloneURL.FindStringSubmatch(raw)
		if matches != nil {
			repoPath := normalizeRemoteRepoPath(matches[3])
			if err := validateRemoteRepoPath(repoPath); err != nil {
				return cloneTarget{}, err
			}
			return cloneTarget{
				Scheme:   "ssh",
				User:     strings.TrimSpace(matches[1]),
				Host:     normalizeHost(matches[2]),
				RepoPath: repoPath,
			}, nil
		}
	}

	u, err := url.Parse(raw)
	if err != nil {
		return cloneTarget{}, fmt.Errorf("clone url is invalid: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return cloneTarget{}, fmt.Errorf("clone url is invalid: %v", raw)
	}

	repoPath := normalizeRemoteRepoPath(u.Path)
	if err := validateRemoteRepoPath(repoPath); err != nil {
		return cloneTarget{}, err
	}

	return cloneTarget{
		Scheme:   strings.ToLower(u.Scheme),
		User:     u.User.Username(),
		Host:     normalizeHost(u.Hostname()),
		Port:     u.Port(),
		RepoPath: repoPath,
	}, nil
}

func (t cloneTarget) DeriveLocalPath() string {
	parts := splitPathSegments(t.RepoPath)
	return filepath.Join(parts...)
}

func (t cloneTarget) IdentityKey() string {
	key := []string{t.Host, t.RepoPath}
	if user := t.identityUser(); user != "" {
		key = append(key, "user="+user)
	}
	if port := t.identityPort(); port != "" {
		key = append(key, "port="+port)
	}
	return strings.Join(key, "|")
}

func (t cloneTarget) identityUser() string {
	if t.Scheme != "ssh" {
		return ""
	}
	user := strings.TrimSpace(t.User)
	if user == "" || user == "git" {
		return ""
	}
	return user
}

func (t cloneTarget) identityPort() string {
	port := strings.TrimSpace(t.Port)
	if port == "" || port == defaultPortForScheme(t.Scheme) {
		return ""
	}
	return port
}

func normalizeRemoteRepoPath(repoPath string) string {
	repoPath = strings.TrimSpace(repoPath)
	repoPath = strings.TrimPrefix(repoPath, "/")
	repoPath = strings.TrimSuffix(repoPath, "/")
	repoPath = strings.TrimSuffix(repoPath, ".git")
	return repoPath
}

func normalizeHost(host string) string {
	return strings.ToLower(strings.TrimSpace(host))
}

func defaultPortForScheme(scheme string) string {
	switch scheme {
	case "http":
		return "80"
	case "https":
		return "443"
	case "ssh":
		return "22"
	default:
		return ""
	}
}

func validateRelativePath(localPath string) error {
	rawPath := strings.TrimSpace(localPath)
	if rawPath == "" {
		return fmt.Errorf("local path is required")
	}

	for _, part := range splitPathSegments(rawPath) {
		if part == "." || part == ".." {
			return fmt.Errorf("local path is invalid: %v", localPath)
		}
	}

	localPath = filepath.Clean(rawPath)
	if localPath == "." {
		return fmt.Errorf("local path is required")
	}
	if filepath.IsAbs(localPath) {
		return fmt.Errorf("local path must be relative: %v", localPath)
	}

	for _, part := range splitPathSegments(localPath) {
		if part == "" || part == "." || part == ".." {
			return fmt.Errorf("local path is invalid: %v", localPath)
		}
	}
	return nil
}

func validateRemoteRepoPath(repoPath string) error {
	if repoPath == "" {
		return fmt.Errorf("clone url must include a repository path")
	}
	for _, part := range splitPathSegments(repoPath) {
		if part == "" || part == "." || part == ".." {
			return fmt.Errorf("clone url has an invalid repository path: %v", repoPath)
		}
	}
	return nil
}

func splitPathSegments(path string) []string {
	return strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == '\\'
	})
}

func toDisplayPath(path string) string {
	return filepath.ToSlash(path)
}
