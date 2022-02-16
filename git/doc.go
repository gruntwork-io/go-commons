// Package git contains routines for interacting with git over the CLI. Unlike go-git, this is not a pure go library for
// interacting with git, and relies on the git command line being available locally. This is meant to provide high level
// interfaces used throughout various Gruntwork CLIs.
//
// NOTE: The tests for these packages are intentionally stored in a separate test folder, rather than the go
// converntional style of source_test.go files. This is done to ensure that the tests for the packages are run in a
// separate docker container, as many functions in this package pollute the global git configuration.
package git
