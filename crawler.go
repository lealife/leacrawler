package leacrawler  
  
import (
    "io/ioutil"
    "net/http"
    "strings"
    "regexp"
    "log"
    "os"
    "path/filepath"
    "github.com/lealife/leacrawler/util"
    "sync"
)

type Crawler struct {
	indexUrl string
	scheme string // http:// 或 https://
	host string // www.lealife.com lealife.com
	schemeAndHost string // http://lealife.com
	targetPath string
	noChildrenFileExts []string
	hadDoneUrl map[string] bool
	exceptionUrl map[string] bool
	
	defaultFilename string // 生成的文件名
	t int
	goroutineNum int // 正在运行的goroutine数目
	lock *sync.Mutex
	// 并发
	w sync.WaitGroup
	ch chan bool
}

// 实例化Crawler
func NewCrawler() *Crawler {
	lea := &Crawler{
		targetPath: "D:",
		defaultFilename: "index.html",
		t: 1,
		goroutineNum: 0,
		lock: &sync.Mutex{},
		noChildrenFileExts: []string{".js", ".ico", ".png", ".jpg", ".gif"}}
	lea.ch = make(chan bool, 1000) // 仅limit个goroutine
	lea.hadDoneUrl = make(map[string]bool, 1000)
	lea.exceptionUrl = make(map[string]bool, 1000)
	
	lea.setLogOutputWriter()
	return lea
}

// 入口
func (this *Crawler) Fetch(url, targetPath string) {
	url = strings.TrimSpace(url)
	
	this.parseUrl(url)
	
	// 保存路径
	this.doTargetPath(targetPath)
	
	// 去掉scheme
	// a.com, a.com/index.html
	url = util.Substring(url, len(this.scheme))
	
//	url2, ok := this.getRalativeUrl("a.com/b/c/d/kk/eee.html", "http://a.com/e/c/d/kk")
//	println(url2)
//	println(ok)
//	return
	
	this.goDo(url, false)
	this.w.Wait()
	
	// 处理异常
	this.doExceptionUrl()
}

// go routine do it
func (this *Crawler) goDo(url string, needException bool) {
	// this.do(url, false)
	this.w.Add(1)
	
	// println(">>>>>>>>>>>>申请资源" + url)
	this.ch <- true // 使用资源
	// println(">>>>>>>>>>>>申请资源成功" + url)
	this.lock.Lock()
		this.goroutineNum++
		log.Println("当前共运行", this.goroutineNum, "goroutine")
	this.lock.Unlock()
	go func() {
		defer func() {
			this.w.Done()
		}()
		children := this.do(url, needException)
		
		this.lock.Lock()
			this.goroutineNum--
			log.Println("当前共运行", this.goroutineNum, " goroutine")
		this.lock.Unlock()
		
		// println("<<<<<<<<<<<<<释放资源" + url)
		<-this.ch // 释放资源
		
		for _, cUrl := range children {
			this.goDo(cUrl, false)
		}
	}()
}

// needException 需要处理异常?
// 这里的url可能是: a.com/b/c/d(没有schema), 不是以.html, .css, .js为后缀的
// 那么断定是一个页面, 此时自动生成一个文件名 => a.com/b/c/d/d_leaui_index.html
// 生成的文件名都按一个规则即可, 不必事先mapping
// 返回一个[]string 待处理的子
func (this *Crawler) do(url string, needException bool) (children []string) {
	children = nil
	url = this.trimUrl(url)
	if this.isNotNeedUrl(url, needException) {
		return;
	}
	
	// 文件是否已存在
	// url = a.com/a/?id=12&id=1221, 那么genUrl=a.com/a/index.html?id=121
	genUrl := this.genUrl(url)
	if this.isExists(genUrl)  {
		return;
	}
	
	// 得到内容
	fullUrl := this.scheme + url
	if needException {
		log.Println("正在处理 `异常` " + fullUrl)
	} else {
		log.Println("正在处理 " + fullUrl)
	}
	
	content, err := this.getContent(fullUrl)
	if !needException && (err != nil || content == "") { // !needException防止处理异常时无限循环
		this.exceptionUrl[url] = true
		return;
	}
	
	this.hadDoneUrl[url] = true
	
	ext := strings.ToLower(filepath.Ext(this.trimQueryParams(url))) // 很可能是a.css?v=1.3
	// css文件中 url(../../img/search-icon.png)
	if(ext == ".css") {
		children = this.doCSS(url, content)
		return;
	}
	
	// 如果是js, image文件就不往下执行了
	if(util.InArray(this.noChildrenFileExts, ext)) {
		// 保存该文件
		if !this.writeFile(url, content) {
			return;
		}
		return;
	}
	
	if(this.t == 1) {
		// 解析html里的href, src
		children = this.doHTML(url, genUrl, content)
	}
	
	return
}

// 处理css
func (this *Crawler) doCSS(url, content string) (children []string) {
	children = nil
	// 保存该文件
	if !this.writeFile(url, content) {
		return;
	}
		
	regular := "(?i)url\\((.+?)\\)"
	reg := regexp.MustCompile(regular)
	re := reg.FindAllStringSubmatch(content, -1)
	
	log.Println(url + " 含有: ");
	log.Println(re);
	baseDir := filepath.Dir(url)
	
	for _, each := range re {
		cUrl := this.trimUrl(each[1])
		// 这里, goDo会申请资源, 导致doCSS一直不能释放资源
		children = append(children, this.cleanUrl(baseDir + "/" + cUrl))
	}
	
	return
}

// url : a.com/a/b/d.html
// a.com/a/b/c genFilename: c_leaui_index.html
// 生成子的相对目录有用
func (this *Crawler) doHTML(pUrl, realPUrl, content string) (children []string) {
	regular := "(?i)(src=|href=)[\"']([^#].*?)[\"']"
	reg := regexp.MustCompile(regular)
	re := reg.FindAllStringSubmatch(content, -1)
	
	log.Println(pUrl + " => " + realPUrl);
	log.Println(pUrl + " 含有: ");
	//log.Println(re);
	
	baseDir := filepath.Dir(realPUrl)
	for _, each := range re {
		// 为了完整替换
		// 只替换src=""里的会有子串的问题, 一个url是另一个url子串
		rawFullUrl := each[0] // src='http://www.uiueux.com/wp/webzine/wp-content/themes/webzine/js/googlefont.js.php?ver=1.6.4'
		rawFullUrlPrefix := each[1]; // src=
		
		// http://a.com/, /a/b/c/d.html, /a/b.jgp
		// 如果是/a/b.jpg, 那么是相对host的, 而不是本文件的路径
		rawCUrl := each[2]
		cUrl := rawCUrl; // strings.TrimRight(rawCUrl, "/") // 万一就是/呢?
		
		// 如果一个链接以//开头, 那么省略了http:, 如果以/开头, 则相对于host
		prefixNotHttp := false
		if strings.HasPrefix(cUrl, "//") {
			cUrl = this.scheme + util.Substring(cUrl, 2)
			prefixNotHttp = true
		} else if strings.HasPrefix(cUrl, "/") {
			cUrl = this.schemeAndHost + cUrl
		}
		
		// 如果这个url是一个目录, 新建一个文件
		// 如果这个url是以http://a.com开头的, host是一样的, 
		// 那么content的url是相对于该url
		// 生成的url, 如果是目录, 会生成一个文件
		cRealUrl, ok := this.getRalativeUrl(realPUrl, cUrl)
		
		// 错误, 不是本页面, 本host的页面
		if ok == -1 {
			// 如果之前//替换成了http://
			if prefixNotHttp {
				content = strings.Replace(content, rawFullUrl, rawFullUrlPrefix + "\"" + cRealUrl + "\"", -1)
			}
			continue
		}
		// 表示已处理过, 是相对目录了, 必须把内容的替换掉
		// 但要处理的还是之前的链接http://
		if ok == 1 {
			cRealUrl = strings.Trim(cRealUrl, "/")
			// 把//变成/
			for strings.Index(cRealUrl, "//") != -1 {
				cRealUrl = strings.Replace(cRealUrl, "//", "/", -1)
			}
			log.Println(rawCUrl + " >>>>>> "  + cRealUrl)
			content = strings.Replace(content, rawFullUrl, rawFullUrlPrefix + "\"" + cRealUrl + "\"", -1)
			cUrl = strings.Replace(cUrl, this.scheme, "", 1) // 把sheme去掉, do
			children = append(children, cUrl) // 不需要clean
		} else {
			children = append(children, this.cleanUrl(baseDir + "/" + cRealUrl))
		}
	}
	
	// 把content保存起来
	if !this.writeFile(realPUrl, content) {
		return;
	}
	
	// this.t++
	// return
	
	return
}

// 得到相对目录
// realPUrl: a.com/b/c/index.html 不是a.com/b/c
// cUrl如果是以this.scheme + this.host开头, 则需要转换成相对目录
// cUrl a.com/c/d/e/g
// 在realPUrl页面到cUrl跳转
func (this *Crawler) getRalativeUrl(realPUrl, cUrl string) (url string, ok int) {
	ok = 0
	url = cUrl
	
	if strings.HasPrefix(cUrl, this.scheme + this.host) {
		url = ""
		ok = 1
		realCUrl := this.genUrl(cUrl) // 如果是目录, 生成一个
		// 如果realPUrl == realCurl 那么返回"#"
		realPUrl = strings.Replace(realPUrl, this.host, "", 1) // 去掉a.com
		realCUrl = strings.Replace(realCUrl, this.scheme + this.host, "", 1) // 去掉http://a.com
		
		realPUrl = this.trimUrl(realPUrl)
		realCUrl = this.trimUrl(realCUrl)
		
		if realPUrl == realCUrl {
			url = "#"
			return
		}
		
		// 去掉两个url相同的部分
		realPUrlArr := strings.Split(realPUrl, "/")
		realCUrlArr := strings.Split(realCUrl, "/")
		log.Println(realPUrlArr)
		log.Println(realCUrlArr)
		i, j := 0, 0
		for ; i < len(realCUrlArr) && j < len(realPUrlArr) && realCUrlArr[i] == realPUrlArr[j]; {
			realCUrlArr[i] = ""
			i++
			j++
		}
		
		// 有多个少../?
		n := len(realPUrlArr) - i - 1
		for k := 0; k < n; k++ {
			url += "../"
		}
		url += strings.Join(realCUrlArr, "/")
		
		return;
	}
	
	// 如果是以http://, https://开头的, 返回false
	if strings.HasPrefix(cUrl, "http://") || strings.HasPrefix(cUrl, "https://") {
		ok = -1
		return
	}
	
	return
}

// trimSpace, /, \, ", '
func (this *Crawler) trimUrl(url string) string {
	if(url != "") {
		url = strings.TrimSpace(url)
		url = strings.Trim(url, "\"")
		url = strings.Trim(url, "'")
		url = strings.Trim(url, "/")
		url = strings.Trim(url, "\\")
	}
	
	return url
}

// 处理异常
func (this *Crawler) doExceptionUrl() {
	if(len(this.exceptionUrl) > 0) {
		log.Println("正在处理异常Url....");
		for url, _ := range this.exceptionUrl {
			this.do(url, true)
		}
	}
}

// 如果url是 a.com/b/c/d 
// 生成一个文件a.com/b/c/d/d_leaui_index.html
// 返回 d_leaui_index.html
// 如果不是一个目录, 返回""
func (this *Crawler) genFilename(url string) (string, bool) {
	urlArr := strings.Split(url, "/")
	if urlArr != nil  {
		last := urlArr[len(urlArr) - 1]
		ext := strings.ToLower(filepath.Ext(last))
		if ext == "" {
			return this.defaultFilename, true // 需要append到url后面
		} else if util.InArray([]string{".php", ".jsp", ".asp", ".aspx"}, ext) {
			filename := filepath.Base(last) // a.php
			filename = util.Substr(filename, 0, len(filename) - len(ext)) // a
			return filename + ".html", false
		}
	}
	return "", true;
}

// 生成真实的url
// 传来的url可能是http://a.com, 也可能是a.com
// getRelativeUrl传来的可以是http://a.com
// url = a.com/a/?id=12&id=1221, 那么genUrl=a.com/a/index.html?id=121
func (this *Crawler) genUrl(url string) string {
	// 去掉?后面的
	queryParam, fragment := "", "" // 包含?,#
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

	// 如果url == host
	if url == this.host || url == this.schemeAndHost {
		return url + "/" + this.defaultFilename + queryParam + fragment
	}

	genFilename, needApend := this.genFilename(url)
	if genFilename != "" {
		if needApend {
			url += "/" + genFilename + queryParam + fragment
		} else {
			// 是a.php => a.html
			urlArr := strings.Split(url, "/")
			urlArr = urlArr[:len(urlArr)-1]
			url = strings.Join(urlArr, "/") + "/" + genFilename
		}
	}
	
	return url
}

func (this *Crawler) writeFile(url, content string) bool {
	// $path = a.html?a=a11
	url = this.trimQueryParams(url)
	
	fullPath := this.targetPath + "/" + url
	dir := filepath.Dir(fullPath)
	log.Println("写目录", dir);
	if err := os.MkdirAll(dir, 0777); err != nil {
		log.Println("写目录" + dir + " 失败")
		return false
	}
	
	// 写到文件中
	file, err := os.Create(fullPath)
    defer file.Close()
    if err != nil {
    	log.Println("写文件" + fullPath + " 失败")
		return false
	}
	file.WriteString(content)
	return true;
}

func (this *Crawler) cleanUrl(url string) string {
	url = filepath.Clean(url)
	return strings.Replace(url, "\\", "/", -1)
}


// 将url ?, #后面的字符串去掉
func (this *Crawler) trimQueryParams(url string) string {
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

// 判断是否已存在
// url = a/b/c/d.html
func (this *Crawler) isExists(url string) bool {
	return util.IsExists(this.targetPath + "/" + url)
}

// 不需要处理的url
// needException false 表示不要处理, 那么就要判断是否在其中
func (this *Crawler) isNotNeedUrl(url string, needException bool) bool {
	if  _, ok := this.hadDoneUrl[url]; ok {
		return true
	}
	_, ok := this.exceptionUrl[url];
	if !needException && ok {
		return true
	}
	
	// http:\\/|https:\\/|
	regular := "#|javascript:|mailto:|&quot; class=|@.*?\\..+"
	reg := regexp.MustCompile(regular)
	if reg.MatchString(url) {
		return true
	}
	
	if (strings.HasPrefix(url, "http:/") || strings.HasPrefix(url, "https:/")) && 
		!strings.HasPrefix(url, this.scheme + this.host) {
		return true
	}
	
	return false
}

// 处理url, 得到scheme, host
func (this *Crawler) parseUrl(url string) {
	if(strings.HasPrefix(url, "http://")) {
		this.scheme = "http://";
	} else {
		this.scheme = "https://";
	}
	
	// http://lealife.com/b/c
	url = strings.Replace(url, this.scheme, "", 1)
	index := strings.Index(url, "/")
	if(index == -1) {
		this.host = url
	} else {
		this.host = util.Substr(url, 0, index)
	}
	
	this.schemeAndHost = this.scheme + this.host
}

func (this *Crawler) getNoChildrenFileExts() []string {
	return this.noChildrenFileExts;
}

// 得到内容
func (this *Crawler) getContent(url string) (content string, err error) {
	var resp *http.Response
	resp, err = http.Get(url)
	if(resp != nil && resp.Body != nil) {
		defer resp.Body.Close()
	} else {
		log.Println("ERROR " + url + " 返回为空 ")
	}
    if resp == nil || resp.Body == nil || err != nil || resp.StatusCode != http.StatusOK {
    	log.Println("ERROR " + url)
		log.Println(err)
		return
    }
    
    var buf []byte
   	buf, err = ioutil.ReadAll(resp.Body)
   	if(err != nil) {
		return
	}
   	content = string(buf);
    return
}

// 生成存储位置
func (this *Crawler) doTargetPath(path string) {
	path = strings.TrimRight(path, "/"); // 不能TrimLeft, 万一是linux呢?
	path = strings.Trim(path, "\\");
	if path != "" {
		this.targetPath = path;
	}
	
	// 生成目录
	if this.targetPath != "" {
		os.MkdirAll(this.targetPath, 0777)
	} else {
		panic("存储位置异常")
	}
}

func (this *Crawler) setLogOutputWriter() {
	/*
	logfile, err := os.OpenFile("C:/Users/Administrator/workspace/lea/log.txt", os.O_RDWR|os.O_CREATE, 0);
	if err != nil {
        log.Printf("%s\r\n", err.Error());
        os.Exit(-1);
	}
	log.SetOutput(logfile)
	*/
}