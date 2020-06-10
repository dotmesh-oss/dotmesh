# Releasing dotmesh

Our release process is automated through Gitlab pipelines. The file
`.gitlab-ci.yml` in the root directory of this repo contains all the
mechanisms that enable this.

## Branches and versions

Releases are made from branches matching the pattern `release-X.Y`, eg
`release-0.1`. Every commit on such a branch is a release version, and
the version string will be `release-X.Y.0` for the first commit on the
branch (which is the commit it shares with the parent branch, which
should be `master`), or `release-X.Y.Z`, where `Z` is the number of
commits not in common with the master branch (e.g. merge commits).

You can check the generated version string from any given git state by doing this:

```
$ (cd cmd/versioner/; go run versioner.go)
release-0.1.1
```

To do a new version release, create a branch called
`release-X.Y` (`git checkout -b release-X.Y`) where X and Y are major and minor versions for your release; to do a patch release, you would merge into the version branch,
like so:

```
$ git checkout release-x.y
$ git pull origin release-x.y
$ git merge --no-ff master
$ (cd cmd/versioner/; go run versioner.go)
release-x.y.z
$ git push origin release-x.y
```

e.g patching version 0.1 would be `git checkout release-0.1` etc.

## The build artefacts

The products of a build fall into three camps.

### Docker images

The backend components of Dotmesh are all distributed as Docker
images. The `build` stage in the Gitlab CI pipelines push them
directory to `quay.io`, with the full Git hash of the commit being
built as the tag. There is no `:latest`, nor use of version numbers as
Docker image tags, because we don't publish the docker image names -
they (with their tags) are hardcoded into the binaries and YAML we
distribute.

As such, there is no distinction between release and non-release
builds; they all go into `quay.io` identified by their Git hash. It
keeps it simple, but we'll have fun one day when we decide to start
garbage collecting old master builds and we need to distinguish them
from release builds we want to keep!

### Client Binaries

We build `dm` client binaries for various platforms. They are uploaded
to `get.dotmesh.io`, with relative paths of the form `PLATFORM/dm`
(`PLATFORM`=`Linux` or `Darwin` at the time of writing). This binary
has the git hash of the Docker images it needs to pull down embedded
inside it, so will automatically refer to the correct server images.

### Kubernetes YAML

We also build YAML files to install Dotmesh into Kubernetes, which is
also uploaded to `get.dotmesh.io`, with a relative path of
`yaml/*.yaml` as they are platform-independent.

## Automated deployment

The `deploy` stage drives automatic continuous deployment, and which
jobs activate depend on the name of the branch.

### `master`

Master builds trigger the `deploy_master_build` job. This copies the
build artefacts to `get.dotmesh.io/unstable/master`, overwriting
whatever was there previously. We don't keep old master builds around.

The job is written such that, if we change it to trigger on any
non-release branch rather than just `master`, it would also do the
same for other branches - putting their latest builds in
`get.dotmesh.io/unstable/BRANCH`.

### `release-.*`

Release branch builds get a version computed by running the versioner,
found in-tree at `cmd/versioner`; this generates a human-readable
version string based on the branch name and the commit number in the
branch, eg `release-0.1.0`.

The artefacts are deployed to `get.dotmesh.io/VERSION`. As the version
changes with every commit, this creates a continuous record of every
release build.

## Making it live

However, we do not publish the above `get.dotmesh.io` URLs, or run
their builds in production. Additional steps are required to mark a
build as the "latest stable build" that appears at the root of
`get.dotmesh.io`, which is linked to from our documentation, These are manual
jobs in the `manual_deploy` stage.

### `mark_release_as_stable`

This is only available for builds from `release-*` branches. If
triggered, it moves the symlinks from the root of `get.dotmesh.io/...`
to point to `get.dotmesh.io/VERSION/...`, thereby making the published
URLs now point to this version.

## How to Make a Release

### Decide if it's a point release (keeping the first two parts of the version the same) or not

If we update the first two parts of the version, we need a new release
branch. Point releases are just a new commit on the same branch.

#### Major release X.Y.0

Create a branch called `test-X.Y` from `master`.

#### Minor release X.Y.Z

Create a branch called `test-X.Y` from `release-X.Y`, and run `git merge origin/master`.

### Test

Smoke test your `test-X.Y` locally, and if it passes, push the branch
to github so that CI has a go at it.

### Release it

If it works, it's time to make it official.

For a major release, create a new `release-X.Y` branch from `test-X.Y`.

For a minor release, fast-forward `release-X.Y` up to `test-X.Y`.

Push `release-X.Y` and delete `test-X.Y`.

### Updating the releases on Github

We direct people here to see the release history:

https://github.com/dotmesh-oss/dotmesh/releases

This currently needs to be manually updated.

 * Create a new release tag in the github UI. This opens up a window to enter details.
 * Call the tag `release-X.Y.Z` and pick the correct release branch
 * Write a description and release notes, by copying the pattern from an existing tag.
 * Press the button to create the release

Try the latest binary on https://dotmesh.com/try-dotmesh/ with a dm
version to check that it's all deployed correctly.

## Pushing live version of docs

There might be docs issues that talk about as yet unreleased features.  These
issues should be in the `blocked` column of the kanban board.

Once the release is complete - open [the pipeline for the docs
repo](http://gitlab.dotmesh.io:9999/dotmesh/docs-sync/pipelines) and
click the `deploy to production` job on the latest pipeline run.

Do this once the release is complete - now the docs and the released software
should be lining up!

### Re-releasing
In the event of a failure during the release process, for example, CI failures, or an embarassing bug fix on the same tag, it is possible to re-release by triggering the pipeline on http://gitlab.dotmesh.io:9999/dotmesh/dotmesh-sync/pipelines on the commit on the release branch. If the commit is on a pre-exisiting release tag, it will preserve the version of the tag.

In rare cases where there is a requirement to re-do a release with different code, this can be done by resetting the branch and preserving the relese-x.y.z tag on the commit and force-pushing. GitLab doesn't handle syncing force pushes, so you will need to temporarily disable branch protection for your user and force push to dotmesh-sync after doing this.
