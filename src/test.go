package main

import (
	"regexp"
	"path/filepath"
	"os"
	"lealife/util"
	"strings"
	"fmt"
	"sort"
)

func trimQueryParams(url string) string {
	pos := strings.Index(url, "?");
	if pos != -1 {
		url = util.Substr(url, 0, pos);
	}
	
	pos = strings.Index(url, "#");
	if pos != -1 {
		url = util.Substr(url, 0, pos);
	}
	return url;
}

func main() {
	regular := "http:\\/|https:\\/|javascript:|mailto:|&quot; class=|@.*?\\..+"
	reg := regexp.MustCompile(regular)
	url := "javascript:"
	
	println(reg.MatchString(url))
	filename := filepath.Dir("a/b/c/aaabceeeeeeee.php") // a.php
	println(filename)
	filename = util.Substr(filename, 0, len(filename) - len(".php")) // a
	println(filename)
	
	// _, err2 := os.Open("D:\\a.jpg")
	_, err2 := os.Stat("D:/a.jpg")
	if err2 != nil && os.IsNotExist(err2) {
        println("file not exist!\n")
    } else {
    	println("file exists")
    }
    
	println(trimQueryParams("a.jgp33333##"))
	
	regular = "(?i)(src=|href=)[\"']([^#].*?)[\"']"
	reg = regexp.MustCompile(regular)
	println(reg.MatchString("url(a)"))
	println()
	re := reg.FindAllStringSubmatch("src='xaaxx3333333333331'  href=\"aaaxx\"", -1)
	for _, each := range re {
		fmt.Println(each[2]);
	}
	
	url = "http://www.a.comddd/b/c"
	url = strings.Replace(url, "http://", "", 1)
	index := strings.Index(url, "/")
	if(index == -1) {
		println(url)
	} else {
		println(util.Substr(url, 0, index))
	}
	
	println("--------------")
	arr := []string{"life", "ax", "cdj"}
	sort.Strings(arr)
	fmt.Println(arr)
	
	cUrl := "//lifedddddddddddddddddddd"
	println(util.Substring(cUrl, 2))
	
	queryParam, fragment := "", ""
	url = "http://a.com?id=12#ddd"
	pos := strings.Index(url, "?");
	if pos != -1 {
		queryParam = util.Substring(url, pos)
		url = util.Substr(url, 0, pos);
	} else {
		pos = strings.Index(url, "#");
		if pos != -1 {
			fragment = util.Substring(url, pos)
			url = util.Substr(url, 0, pos);
		}
	}
	
	println(queryParam, fragment)
	
	urlArr := strings.Split("a/b/c/d.html", "/")
	urlArr = urlArr[:len(urlArr)-1]
	println(strings.Join(urlArr, "/"))
}