# Contribution Guide

* [Before you get started](#before-you-get-started)
    * [Code of Conduct](#code-of-conduct)
* [Your First Contribution](#your-first-contribution)
    * [Find a good first topic](#find-a-good-first-topic)
* [Setting up your development environment](#setting-up-your-development-environment)
    * [Fork the project](#fork-the-project)
    * [Clone the project](#clone-the-project)
    * [New branch for a new code](#new-branch-for-a-new-code)
    * [Test](#test)
    * [Commit and push](#commit-and-push)
    * [Create a Pull Request](#create-a-pull-request)
    * [Sign the CLA](#sign-the-cla)
    * [Get a code review](#get-a-code-review)

## Before you get started

### Code of Conduct

Please make sure to read and observe our [Code of Conduct](./CODE_OF_CONDUCT.md).

### Developer Certificate of Origin - DCO

Please make sure to read and observe our [DCO](./DCO).

## Your First Contribution

### Find a good first topic

You can start by finding an existing issue with the
[good first issue](https://gitlab.com/WaldirBorbaJR/kvstok/-/issues) or [help wanted](https://gitlab.com/WaldirBorbaJR/kvstok/-/issues) labels. These issues are well suited for new contributors.


## Setting up your development environment

KVStoK uses [`Go Modules`](https://github.com/golang/go/wiki/Modules)
to manage dependencies. The version of Go should be **1.17** or above.

### Fork the project

- Visit https://gitlab.com/WaldirBorbaJR/kvstok
- Click the `Fork` button (top right) to create a fork of the repository

### Clone the project

```sh
$ git clone https://gitlab.com/$GITLAB_USER/kvstok
$ cd kvstok
$ git remote add upstream git@gitlab.com:waldirborbajr/kvstok.git

# Never push to the upstream master
git remote set-url --push upstream no_push
```

### New branch for a new code

Get your local master up to date:

```sh
$ git fetch upstream
$ git checkout master
$ git rebase upstream/master
```

Create a new branch from the master:

```sh
$ git checkout -b my_new_feature
```

And now you can finally add your changes to project.

### Test

Build and run all tests:

```sh
$ ./test.sh
```

### Commit and push

Commit your changes:

```sh
$ git commit
```

When the changes are ready to review:

```sh
$ git push origin my_new_feature
```

### Create a Pull Request

Just open `https://gitlab.com/$GITLAB_USER/kvstok/pull/new/my_new_feature` and
fill the PR description.

### Sign the CLA

Click the **Sign in with Gitlab to agree** button to sign the CLA. [An example](https://cla-assistant.io/kvstok/badger?pullRequest=1377).

### Get a code review

If your pull request (PR) is opened, it will be assigned to one or more
reviewers. Those reviewers will do a code review.

To address review comments, you should commit the changes to the same branch of
the PR on your fork.
