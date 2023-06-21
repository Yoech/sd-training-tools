package CCCommon

import (
	"time"
)

func TimeMillSecond() int64 {
	now := time.Now()
	unix := now.UnixNano() / 1e6
	return unix
}
