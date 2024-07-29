package main

import (
	"fmt"
	"time"
	"waterpool/pool"
)

var a int = 0

func main() {
	p := pool.NewPool(5)
	for i := 0; i < 10; i++ {
		p.Schedule(send)
	}

}

func send() {
	fmt.Printf("hello, %d \n", a)
	a++
	time.Sleep(2 * time.Second)
}
