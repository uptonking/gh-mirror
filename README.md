# gh-mirror

> forked from [ntns/gh-mirror](https://github.com/ntns/gh-mirror)

- A simple tool to clone/sync Git repositories locally with a `gh-mirror.json` config file.
# use cases
- Keeping a local backup of Git repositories
  - easy to backup/sync with a config file
  - supports GitHub shorthand and custom clone URLs
# installation

 `go install github.com/uptonking/gh-mirror`

```sh
# run
go run .
# build
go build -o gh-mirror .
```

# usage

```
$ gh-mirror -init                # init gh-mirror directory and config
$ gh-mirror -add username/repo   # add GitHub shorthand
$ gh-mirror -add git@gitlab.com:group/repo.git
$ gh-mirror -add ssh://git@code.haverbeke.berlin/codemirror/merge
$ gh-mirror -add https://code.haverbeke.berlin/codemirror/merge.git
$ gh-mirror -list                # list repos added to gh-mirror
$ gh-mirror                      # mirror and update repos
```

- You can use gh-mirror from any directory. 
- GitHub shorthand repos are saved at `~/gh-mirror/username/repo`.
- Custom clone URLs default to the remote repo path, such as `~/gh-mirror/prosemirror/prosemirror-view`.
- If two different remotes would land on the same folder, set an explicit `path`.

# config

Legacy GitHub shorthand still works:

```json
{
 "sleepDuration": 1,
 "repos": [
  "owner/repo"
 ]
}
```

Custom clone URLs can be stored as strings:

```json
{
 "sleepDuration": 1,
 "repos": [
  "owner/repo",
  "ssh://git@code.haverbeke.berlin/codemirror/merge",
  "https://code.haverbeke.berlin/codemirror/merge.git",
  "git@gitlab.com:fdroid/fdroidclient.git"
 ]
}
```

Use an object only when you need to override the local path:

```json
{
 "sleepDuration": 1,
 "repos": [
  {
   "url": "https://code.haverbeke.berlin/codemirror/merge",
   "path": "custom/codemirror-merge"
  }
 ]
}
```

# notes
- backup directory is not configurable. default is `~/gh-mirror`.
- concurrent sync/update is not supported
# how it works

New repositories are cloned using the configured clone URL and local path.

Existing repositories are kept up to date using `git pull --rebase`

That's it!
