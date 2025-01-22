package test

import (
	"fmt"
	"testing"
	"time"
)

func TestDriver(t *testing.T) {
	for i := 0; i < 1000; i++ {
		fmt.Println(i)
		time.Sleep(1 * time.Second)
	}

	fmt.Println("done")
}
