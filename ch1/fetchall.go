// Fetchall fetches URLs in parallel and reports their times and sizes.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const TIMEOUT = time.Second * 3

func main() {
	ch := make(chan map[string]string)
	count := 0
	for _, url := range os.Args[1:] {
		if !strings.HasPrefix(url, "http://") {
			url = strings.Join([]string{"http://", url}, "")
		}
		fmt.Printf("start fetch %s\n", url)
		for i := 0; i < 3; i++ {
			go fetch(url, ch)
		}
		count += 3
	}
	result := make(map[string][]string)
	for i := 0; i < count; i++ {
		select {
		case res := <-ch:
			for k, v := range res {
				if _, ok := result[k]; ok {
					result[k] = append(result[k], v)
				} else {
					result[k] = []string{v}
				}
			}
		case <-time.After(TIMEOUT):
			fmt.Println("TIME OUT")
		}
	}
	for k, v := range result {
		fmt.Printf("%s : %v\n", k, v) //输出按照URL排序后的结果
	}
}

func fetch(url string, ch chan<- map[string]string) {
	start := time.Now()
	result := make(map[string]string)
	resp, err := http.Get(url)
	if err != nil {
		result[url] = fmt.Sprintf("http-get: %v", err)
		ch <- result
		return
	}
	nbytes, err := io.Copy(io.Discard, resp.Body)
	resp.Body.Close() // don't leak resources
	if err != nil {
		result[url] = fmt.Sprintf("while reading:%v %v", url, err)
		ch <- result
		return
	}
	secs := time.Since(start).Seconds()
	//fmt.Printf("%v %v %v Bytes %v's\n",url,resp.Status,nbytes,secs)
	result[url] = fmt.Sprintf("%v %v %v Bytes %v's", url, resp.Status, nbytes, secs)
	ch <- result
}
