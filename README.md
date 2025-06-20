# GVM (Global Version Manager)
A programming language version manager, like `nvm`, but extensible to support all programming languages.

![Workflow ci](https://github.com/toodofun/gvm/actions/workflows/go.yml/badge.svg)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/toodofun/gvm/blob/master/LICENSE)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/toodofun/gvm?logo=go)
[![Go Report Card](https://goreportcard.com/badge/github.com/toodofun/gvm)](https://goreportcard.com/report/github.com/toodofun/gvm)
[![Test Coverage](https://codecov.io/gh/toodofun/gvm/branch/master/graph/badge.svg)](https://codecov.io/gh/toodofun/gvm)

* 🚀 Supported Interaction Methods
    * Command Line Interface (CLI)
        * `ls-remote <lang>`: List remote versions of a language
        * `ls <lang>`: List installed versions of a language
        * `install <lang> <version>`: Install a specific version of a language
        * `uninstall <lang> <version>`: Uninstall a specific version of a language
        * `use <lang> <version>`: Set the default version of a language
        * `current <lang>`: Show the current version of a language
    * Terminal User Interface (TUI)
        * `ui`: Run in terminal interface
* 🚀 Supported Programming Languages
    * [x] Golang
    * [ ] Node
    * [ ] Java
    * [ ] Python

## Screenshots
### Languages Page
![languages](assets/languages.png)

### Version Management Page
![language-versions](assets/language-versions.png)

## Command Line
```shell
MacBook-Pro :: ~ » gvm -h
A tool to manage multiple versions of programming languages.

Usage:
  gvm [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  current     Show Current version of a language
  help        Help about any command
  install     Install a specific version of a language
  ls          List installed versions of language
  ls-remote   List remote versions of language
  ui          Run in the terminal UI
  uninstall   Uninstall a specific version of a language
  use         Set default versions of language

Flags:
  -d, --debug   debug mode
  -h, --help    help for gvm

Use "gvm [command] --help" for more information about a command.
```

## Acknowledgements
Grateful acknowledgement to [JetBrains](https://www.jetbrains.com/) for supporting this project through their Open Source License Program and providing exceptional development tools.

## Issues

If you have an issue: report it on the [issue tracker](https://github.com/toodofun/gvm/issues)

## Contributing

Contributions are always welcome. For more information, check out the [contributing guide](CONTRIBUTING.md)

## License

Licensed under the MIT License. See [LICENSE](LICENSE) for details.
