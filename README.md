# gh-mirror

> forked from [ntns/gh-mirror](https://github.com/ntns/gh-mirror)

- A simple tool to clone/sync Github repositories locally with a `gh-mirror.json` config file.
# use cases
- Keeping a local backup of Github repositories
  - easy to backup/sync with a config file
  - support public repos on github
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
$ gh-mirror -add username/repo   # add repo to gh-mirror
$ gh-mirror -list                # list repos added to gh-mirror
$ gh-mirror                      # mirror and update repos
```

- You can use gh-mirror from any directory. 
- Repos will always be saved at `~/gh-mirror/username/repo` .

# notes
- backup directory is not configurable. default is `~/gh-mirror`.
- concurrent sync/update is not supported
# how it works

New repositories are cloned using `git clone git@github.com:username/repo.git username/repo `

Existing repositories are kept up to date using `git pull --rebase`

That's it!
