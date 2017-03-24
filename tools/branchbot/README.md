## branchbot

`branchbot` is a tool to make cherry picking PRs across branches easier.

## Install

```
$ make branchbot
```

This will drop a `branchbot` binary into this folder.

To push a new `editions/branchbot` image (currently private) you can do:

```
$ make push
```

## Prod Deployment

The current production `branchbot` is running on https://cloud.docker.com using
the new Swarm mode support.

Administration of the swarm can be done from within the drop-down in Docker for
Mac (and Windows?).

The organization it is deployed to is
[`editions`](https://hub.docker.com/r/editions/) and the name of the swarm is
`taskswarm`.

You can see the `docker-compose.yml` used to generate the `docker service` for
it in this directory.

To deploy, it needs two secrets (currently already set in `taskswarm`):

1. `gh-cherrypick-token` -- a GitHub API token for the
   [editionsbot](https://github.com/editionsbot) (to query and modify PRs)
2. `gh-ssh-key` -- a SSH private key corresponding to a public key registered
   for [editionsbot](https://github.com/editionsbot). (for `git push`/`pull`)

## Usage

`branchbot` will continually scan this repo for PRs with the label prefix
`process/cherrypick-<release>` where `<release>` is currently `17.03` or `17.04`
(more will be added later).

If it finds a PR with this label, it will:

1. Cherry-pick the commits from this PR into the corresponding `17.03` etc.
   branch and push a new branch with these cherry-picks to GitHub. The
   cherry-pick will always favor changes from the cherry-picked commits over
   changes in the target branch, so any conflicts will always be resolved in favor
   of the upstream. (It's not clear if this is the most practical behavior in
   actual usage yet, we're still experimenting).
2. Push a PR, including pings to the relevant maintainers, to the
   `docker/editions` repository.
3. Remove the `process/cherrypick-*` label from the cherry-picked PR.
