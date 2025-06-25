package util

import (
	"fmt"
	"time"
)

func PrintElapsedTime(tag string, beginTime time.Time) {
	elapsed := time.Since(beginTime).Milliseconds() // 得到毫秒
	fmt.Printf("call [%s] time consumed: %d ms\n", tag, elapsed)
}
