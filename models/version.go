package models

import (
	"fmt"
	"runtime"
)

const Binary = "v0.1.0"

var (
	Version = fmt.Sprintf("%s (build %s)", Binary, runtime.Version())
)
