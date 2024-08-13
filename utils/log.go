package utils

import (
	"fmt"
	"time"
)

func Log(message ...any) {
	fmt.Println(time.Now().Format("2006/01/02 15:04:05"), ": INFO :: ", message)
}
func LogError(message ...any) {
	fmt.Println(time.Now().Format("2006/01/02 15:04:05"), ": ERROR :: ", message)
}
