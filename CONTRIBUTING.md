# How to Contribute

We want to make contributing to this project as easy as possible.

## Reporting Issues

If you have an issue, please report it on the [issue tracker](https://github.com/toodofun/gvm/issues).

When you are up for writing a PR to solve the issue you encountered, it's not
needed to first open a separate issue. In that case only opening a PR with a
description of the issue you are trying to solve is just fine.

## Contributing Code

Pull requests are always welcome. When in doubt if your contribution fits within
the rest of the project, feel free to first open an issue to discuss your idea.

This is not needed when fixing a bug or adding an enhancement, as long as the
enhancement you are trying to add can be found in the public NuGet API docs as
this project only supports what is in the public API docs.

## Coding style

We try to follow the Go best practices, where it makes sense, and use
[`goimports`](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) to format code in this project.
As a general rule of thumb we prefer to keep line width for comments below 80
chars and for code (where possible and sensible) below 100 chars.

Before making a PR, please look at the rest this package and try to make sure
your contribution is consistent with the rest of the coding style.

New struct field or methods should be placed (as much as possible) in the same
order as the ordering used in the public API docs. The idea is that this makes it
easier to find things.

### Setting up your local development environment to Contribute to `gvm`

1. [Fork](https://github.com/toodofun/gvm/fork), then clone the repository.
   ```sh
   git clone https://github.com/<your-username>/gvm.git
   # or via ssh
   git clone git@github.com:<your-username>/gvm.git
   ```
1. Install dependencies:
   ```sh
   make tools
   ```
1. Make your changes on your feature branch
1. Run the tests and `goimports`
   ```sh
   make test && make format
   ```
1. Open up your pull request