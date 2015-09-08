package main
import (
	"github.com/lealife/leacrawler"
	"time"
	"fmt"
	"os"
)

func main() {
	argNum := len(os.Args)
	
	if argNum != 3 {
		fmt.Println("-------------")
		fmt.Println("error!!!")
		fmt.Println("Usage: bin url target_path")
		fmt.Println("eg: bin http://baidu.com /Users/web")
		fmt.Println("-------------")
		fmt.Println("author by life. http://leanote.com/blog/life")
		fmt.Println("-------------")
		return;
	}
	
	url := os.Args[1]
	path := os.Args[2]
    
	start := time.Now()
	fmt.Println("start...")

	lea := leacrawler.NewCrawler()
	
//    url := "http://bucketadmin.themebucket.net/index.html"
//    path := "/Users/life/Desktop/LeaSpider";
	lea.Fetch(url, path)
	
	fmt.Printf("time cost %v\n", time.Now().Sub(start))
}