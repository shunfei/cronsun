package cronsun

import (
	"fmt"
	"runtime"
)

const VersionNumber = "0.3.1"

var (
	Version = fmt.Sprintf("v%s (build %s)", VersionNumber, runtime.Version())
)
