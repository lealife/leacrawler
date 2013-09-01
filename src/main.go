package main
import (
	"lealife"
	"time"
	"fmt"
)

func main() {
	start := time.Now()
	fmt.Println("start...")

	lea := lealife.NewLeaSpider()
	lea.Fetch("http://www.lealife.com/", "/Users/life/Desktop/LeaSpider")
	
	fmt.Printf("time cost %v\n", time.Now().Sub(start))
}