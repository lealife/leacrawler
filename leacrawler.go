package leacrawler
import (
	"time"
	"fmt"
)
// Fetch
func Fetch(url, path string) {
	start := time.Now()
	fmt.Println("start...")
	lea := NewCrawler()
	lea.Fetch(url, path)
	fmt.Printf("time cost %v\n", time.Now().Sub(start))
}
