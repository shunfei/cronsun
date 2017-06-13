package cronsun

import (
	"fmt"
	"runtime"
)

const Binary = "v0.1.2"

var (
	Version = fmt.Sprintf("%s (build %s)", Binary, runtime.Version())
)
