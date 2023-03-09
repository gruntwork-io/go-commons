package git

import (
	"github.com/gruntwork-io/go-commons/files"
	"github.com/gruntwork-io/go-commons/shell"
	"github.com/sirupsen/logrus"
)

// Clone runs git clone to clone the specified repository into the given target directory.
func Clone(logger *logrus.Entry, repo string, targetDir string) error {
	if !files.IsDir(targetDir) {
		return TargetDirectoryNotExistsErr{dirPath: targetDir}
	}

	opts := shell.NewShellOptions()
	if logger != nil {
		opts.Logger = logger
	}
	return shell.RunShellCommand(opts, "git", "clone", repo, targetDir)
}

// Checkout checks out the given ref for the repo cloned in the target directory.
func Checkout(logger *logrus.Entry, ref string, targetDir string) error {
	if !files.IsDir(targetDir) {
		return TargetDirectoryNotExistsErr{dirPath: targetDir}
	}

	opts := shell.NewShellOptions()
	if logger != nil {
		opts.Logger = logger
	}
	opts.WorkingDir = targetDir
	return shell.RunShellCommand(opts, "git", "checkout", ref)
}
