package main
import (
	"lealife"
	"time"
	"fmt"
	"os"
)

func main() {
	start := time.Now()
	fmt.Println("start...")

	lea := lealife.NewLeaSpider()
	
    url := "http://www.keenthemes.com/preview/metronic_admin"
    path := "/Users/life/Desktop/LeaSpider";
	lea.Fetch(url, path)
	
	fmt.Printf("time cost %v\n", time.Now().Sub(start))
}