package browser

import "time"

const (
	IDByteLength             = 8
	MaxConcurrentConnections = 10
	DefaultTimeoutSec        = 30 * time.Second
	KeepAliveTimeoutSec      = 30 * time.Second
	TargetFPS                = 60
	MillisecondsPerSecond    = 1000
	FrameTimeMs              = MillisecondsPerSecond / TargetFPS
)
