package git

import (
	"fmt"
)

// TargetDirectoryNotExistsErr is returned when the target directory of the git commands does not exist or is not a
// directory.
type TargetDirectoryNotExistsErr struct {
	dirPath string
}

func (err TargetDirectoryNotExistsErr) Error() string {
	return fmt.Sprintf("%s does not exist or is not a directory", err.dirPath)
}
