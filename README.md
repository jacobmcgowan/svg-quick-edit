# svg-quick-edit

[![Build and Test Status](https://github.com/jacobmcgowan/svg-quick-edit/actions/workflows/go-build-test.yml/badge.svg)](https://github.com/jacobmcgowan/svg-quick-edit/actions/workflows/go-build-test.yml)
[![License: MIT](https://cdn.prod.website-files.com/5e0f1144930a8bc8aace526c/65dd9eb5aaca434fac4f1c34_License-MIT-blue.svg)](/LICENSE)

svg-quick-edit is a CLI tool that allows you to quickly edit
attributes of paths in SVG files. It is useful for batch processing SVG files
to change attributes like 'fill', 'stroke', etc. The tool takes a path to an SVG
file or directory containing SVG files, and modifies the specified attributes
for all paths in the SVG file(s).

## Running

After cloning the repository, run
```bash
go build .
./svg-quick-edit --help
```

See the [command documentation](docs/svg-quick-edit.md) for more details with
examples.
