package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Result struct {
	URL     string
	Details []interface{}
	Error   error
}

func proxy() http.Client {
	userName := "xxxxxx" //账号
	PassWord := "xxxxxx" //密码
	ProxyUrls := "xxxxx.cn:6442"
	proxyURL := "http://" + userName + ":" + PassWord + "@" + ProxyUrls // 替换为你的代理服务器地址
	proxyURLParsed, _ := url.Parse(proxyURL)
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURLParsed),
	}
	return http.Client{Transport: transport}
}
func urlList(url string) (string, error) {
	//client := proxy()
	//resp, err := client.Get("https://www.xxx.com/search/" + url)
	resp, err := http.Get("https://www.dasenic.com/product/chipFindSearch?keyword=" + url)
	if err != nil {
		return "", fmt.Errorf("请求失败: %s", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var urls string
	doc.Find("a.mx-auto").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && strings.Contains(href, url) {
			urls = href
			return // 提取到匹配的链接后直接返回
		}
	})
	return urls, err
}
func getDetails(url string, mpn string) ([]interface{}, error) {
	//client := proxy()
	//resp, err := client.Get("https://www.xxxx.com/" + url)
	resp, err := http.Get("https://www.xxxx.com" + url)
	if err != nil {
		fmt.Println("转跳失败")
		return nil, fmt.Errorf("请求失败: %s", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	//属性
	extractedData := make(map[string]string)
	doc.Find("ul.list-unstyled").Each(func(i int, ul *goquery.Selection) {
		ul.Find("li").Each(func(j int, li *goquery.Selection) {
			key := li.Find("span.text-muted").Text()
			value := li.Find("span").Last().Text()
			extractedData[key] = value
		})
	})
	extractedDataJson, err := json.Marshal(extractedData)
	if err != nil {
		//extractedDataJson = []byte("")
		return nil, fmt.Errorf("转换为JSON失败: %s", err)
	}
	newProduct := []interface{}{mpn,
		extractedDataJson,
	}
	return newProduct, nil
}

func findContentDiv(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.Data == "div" {
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == "content" {
				return node
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if result := findContentDiv(child); result != nil {
			return result
		}
	}

	return nil
}

func extractText(node *html.Node, extractedContent *string) {
	if node.Type == html.TextNode {
		*extractedContent += node.Data
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		extractText(child, extractedContent)
	}
}
func QueryDetails(urls []string) [][]interface{} {
	startTime := time.Now()
	var wg sync.WaitGroup
	resultCh := make(chan Result, len(urls))
	for _, urlArray := range urls {
		wg.Add(1)
		go func(urlArray string) {
			defer wg.Done()
			href, err := urlList(urlArray)
			if err != nil {
				//resultCh <- Result{URL: url, Error: err}
				return
			}
			details, err := getDetails(href, urlArray)
			resultCh <- Result{URL: urlArray, Details: details, Error: err}
		}(urlArray)
	}
	go func() {
		wg.Wait()
		close(resultCh)
	}()
	var results []Result
	for result := range resultCh {
		results = append(results, result)
	}
	elapsed := time.Since(startTime)
	fmt.Printf("Elapsed Time: %s", elapsed)

	// 提取Details字段并组成新的二维数组
	var details [][]interface{}
	for _, result := range results {
		details = append(details, result.Details)
	}
	return details
}
