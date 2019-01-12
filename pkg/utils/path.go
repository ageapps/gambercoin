package utils

import (
	"path"
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

// GetRootPath get
func GetRootPath() string {
	return path.Join(basepath, "..", "..")
}
