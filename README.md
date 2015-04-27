# leacrawler

Use leacrawler you can download web template(include html, css, js, image) as you want. (such as the http://themeforest.net/ web template)

Leacrawler is written by golang. Just for fun and hope you like it.

## usage:
1. go get github.com/lealife/leacrawler
2. create a new file and put the code as follows:
```golang
package main

import (
	"github.com/lealife/leacrawler"
)

func main() {
	// url and the target path
	leacrawler.Fetch("http://lealife.com", "/Users/life/Desktop/lealife")
}
```
