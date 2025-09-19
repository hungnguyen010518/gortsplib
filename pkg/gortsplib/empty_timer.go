package gortsplib

import (
	"time"
)

func emptyTimer() *time.Timer {
	t := time.NewTimer(0)
	<-t.C
	return t
}

func EmptyTimer() *time.Timer {
	return emptyTimer()
}
