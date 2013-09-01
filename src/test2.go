package main

import (
	"sync"
)
var a string
var c = make(chan int)

var w sync.WaitGroup

func f() {
	w.Add(1)
	
	println("??");
	
	w.Done()
}

func main() {
	w.Add(1)
	go func() {
		go f()
		println("--------")
	    w.Done()
	}()
	w.Wait()
}
