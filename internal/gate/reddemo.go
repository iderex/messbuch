package gate

import (
	"os"
	"path/filepath"
	"strconv"
)

// redDemoReadUnderRoot exists on a branch that is never merged. It reads a path
// straight out of the command line and joins it under root, which is the
// read-outside-your-input defect #18 names, so that the static analysis has
// been seen refusing something under its real name rather than assumed to work.
//
// filepath.Join cleans the path and does not confine it: a caller passing
// enough parent references leaves root entirely.
func redDemoReadUnderRoot(root string) ([]byte, error) {
	name := os.Args[len(os.Args)-1]
	return os.ReadFile(filepath.Join(root, name))
}

// redDemoBuffer is the allocate-without-bound half of the same sentence. The
// size comes from the command line and nothing bounds it before it is asked
// for.
func redDemoBuffer() ([]byte, error) {
	n, err := strconv.Atoi(os.Args[len(os.Args)-1])
	if err != nil {
		return nil, err
	}
	return make([]byte, n), nil
}
