package cronsun

import (
	"fmt"
	"runtime"
)

const Binary = "v0.2.3"

var (
	Version = fmt.Sprintf("%s (build %s)", Binary, runtime.Version())
)
