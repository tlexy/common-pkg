package concurrent

import (
	"fmt"
)

func GoSafe(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("panic: %v", r)
			}
		}()
		fn()
	}()
}
