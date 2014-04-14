package leacrawler
import (
	"time"
	"fmt"
)

func Fetch(url, path string) {
	start := time.Now()
	fmt.Println("start...")
	lea := leacrawler.NewCrawler()
	lea.Fetch(url, path)
	fmt.Printf("time cost %v\n", time.Now().Sub(start))
}