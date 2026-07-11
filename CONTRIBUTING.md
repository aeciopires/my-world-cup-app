<!-- TOC -->

- [Contributing](#contributing)
- [Tip](#tip)

<!-- TOC -->

# Contributing

Your contribution is very welcome!

Follow the steps below whenever you want to improve the content of this repository.

- Install the following packages: `git`, `go` (see the version in [README.md](README.md#software-requirements)), `make`, `docker`, `helm`, `helm-docs`, and a text editor of your choice.
- Fork this repository. See this tutorial: https://help.github.com/en/github/getting-started-with-github/fork-a-repo
- Configure authentication on your GitHub account to use the SSH protocol instead of HTTP. Watch this tutorial to learn how to configure it: https://help.github.com/en/github/authenticating-to-github/adding-a-new-ssh-key-to-your-github-account
- Clone the repository resulting from the fork to your computer.
- Add the URL of the upstream repository with the following command:

```bash
git remote -v
git remote add upstream git@github.com:aeciopires/my-world-cup-app.git
git remote -v
```

- Create a branch using the pattern:

```bash
git checkout -b BRANCH_NAME
```

- Make sure you are on the correct branch with the following command:

```bash
git branch
```

- The branch in use has a `*` before its name.
- Make the necessary changes.
- **If your change should ship as a new release**, bump the version in the root [`VERSION`](VERSION) file (plain semver text, e.g. `1.3.0`, no `v` prefix). It's the single source of truth for the release version: `make docker-build`, `make docker-build-multiarch`, and `make docker-push` tag the Docker image from it by default, and `make helm-sync-version` (also run automatically by `make docker-push`) writes it into `charts/my-world-cup-app/Chart.yaml`'s `appVersion`. Also add a matching entry to [`CHANGELOG.md`](CHANGELOG.md). Skip this for changes that don't warrant a new release (docs-only tweaks, CI config, etc.).
- Check that all required tools are installed:

```bash
make check-deps
```

- Format, vet, and test your changes before committing:

```bash
make check
```

- Commit your changes on the newly created branch, preferably making one commit per logical change.
- Push the commits to the remote repository with the command:

```bash
git push --set-upstream origin BRANCH_NAME
```

- Create a Pull Request (PR) against the `main` branch of the original repository. See this [tutorial](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/creating-a-pull-request-from-a-fork).
- Update the content with the reviewer's suggestions (if needed).
- After your PR is approved and merged, update your local repository with the commands below.

```bash
git checkout main
git pull upstream main
```

- Remove the local branch after your PR is approved and merged, using the command:

```bash
git branch -d BRANCH_NAME
```

- Update the `main` branch of your remote fork.

```bash
git push origin main
```

- Push the deletion of the local branch to your repository on GitHub with the command:

```bash
git push --delete origin BRANCH_NAME
```

- To keep your fork in sync with the original repository, run these commands:

```bash
git pull upstream main
git push origin main
```

Reference:
- https://blog.scottlowe.org/2015/01/27/using-fork-branch-git-workflow/

# Tip

**You can use the text editor of your choice, whichever you feel most comfortable with.**

But VSCode (https://code.visualstudio.com), combined with the extensions below, helps the editing/review process, mainly by allowing content preview before commit, checking Markdown syntax, and generating an automatic summary as section titles are created/changed.

- Go: https://marketplace.visualstudio.com/items?itemName=golang.Go
- Markdown-lint: https://marketplace.visualstudio.com/items?itemName=DavidAnson.vscode-markdownlint
- Markdown-toc: https://marketplace.visualstudio.com/items?itemName=AlanWalk.markdown-toc
- Markdown-all-in-one: https://marketplace.visualstudio.com/items?itemName=yzhang.markdown-all-in-one
- YAML: https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml
- Helm-intellisense: https://marketplace.visualstudio.com/items?itemName=Tim-Koehler.helm-intellisense
- Docker: https://marketplace.visualstudio.com/items?itemName=ms-azuretools.vscode-docker
- GitLens: https://marketplace.visualstudio.com/items?itemName=eamodio.gitlens
- Themes for VSCode:
    - https://code.visualstudio.com/docs/getstarted/themes
    - https://vscodethemes.com/
