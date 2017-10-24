package cronsun

import (
	"fmt"
	"runtime"
)

const Binary = "v0.2.2"

var (
	Version = fmt.Sprintf("%s (build %s)", Binary, runtime.Version())
)
