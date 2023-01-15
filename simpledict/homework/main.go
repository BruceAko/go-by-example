package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

var wg sync.WaitGroup

type DictRequest struct {
	TransType string `json:"trans_type"`
	Source    string `json:"source"`
	UserID    string `json:"user_id"`
}

type DictRequest2 struct {
	Q     string
	From  string
	To    string
	Appid int
	Salt  int
	Sign  string
}

type DictResponse struct {
	Rc   int `json:"rc"`
	Wiki struct {
		KnownInLaguages int `json:"known_in_laguages"`
		Description     struct {
			Source string      `json:"source"`
			Target interface{} `json:"target"`
		} `json:"description"`
		ID   string `json:"id"`
		Item struct {
			Source string `json:"source"`
			Target string `json:"target"`
		} `json:"item"`
		ImageURL  string `json:"image_url"`
		IsSubject string `json:"is_subject"`
		Sitelink  string `json:"sitelink"`
	} `json:"wiki"`
	Dictionary struct {
		Prons struct {
			EnUs string `json:"en-us"`
			En   string `json:"en"`
		} `json:"prons"`
		Explanations []string      `json:"explanations"`
		Synonym      []string      `json:"synonym"`
		Antonym      []string      `json:"antonym"`
		WqxExample   [][]string    `json:"wqx_example"`
		Entry        string        `json:"entry"`
		Type         string        `json:"type"`
		Related      []interface{} `json:"related"`
		Source       string        `json:"source"`
	} `json:"dictionary"`
}

type DictResponse2 struct {
	From        string              `json:"from"`
	To          string              `json:"to"`
	TransResult []map[string]string `json:"trans_result"`
}

func query(word string) {
	client := &http.Client{}
	request := DictRequest{TransType: "en2zh", Source: word}
	buf, err := json.Marshal(request)
	if err != nil {
		log.Fatal(err)
	}
	var data = bytes.NewReader(buf)
	req, err := http.NewRequest("POST", "https://api.interpreter.caiyunai.com/v1/dict", data)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("DNT", "1")
	req.Header.Set("os-version", "")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.51 Safari/537.36")
	req.Header.Set("app-name", "xy")
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("device-id", "")
	req.Header.Set("os-type", "web")
	req.Header.Set("X-Authorization", "token:qgemv4jr1y38jyq6vhvi")
	req.Header.Set("Origin", "https://fanyi.caiyunapp.com")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Referer", "https://fanyi.caiyunapp.com/")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cookie", "_ym_uid=16456948721020430059; _ym_d=1645694872")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal("bad StatusCode:", resp.StatusCode, "body", string(bodyText))
	}
	var dictResponse DictResponse
	err = json.Unmarshal(bodyText, &dictResponse)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("彩云翻译结果：")
	fmt.Println(word, "UK:", dictResponse.Dictionary.Prons.En, "US:", dictResponse.Dictionary.Prons.EnUs)
	for _, item := range dictResponse.Dictionary.Explanations {
		fmt.Println(item)
	}
	wg.Done()
}

func query2(word string) {
	var appID = 20210225000707349
	var password = "rXOf9TWDrZSTrqdmUH4r"
	tran := DictRequest2{
		Q:     word,
		From:  "en",
		To:    "cht",
		Appid: appID,
	}
	tran.Salt = time.Now().Second()
	content := strconv.Itoa(appID) + word + strconv.Itoa(tran.Salt) + password
	sign := SumString(content) //计算sign值
	tran.Sign = sign
	values := tran.ToValues()
	resp, err := http.PostForm("https://api.fanyi.baidu.com/api/trans/vip/translate", values)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal("bad StatusCode:", resp.StatusCode, "body", string(body))
	}
	var dictResponse DictResponse2
	err = json.Unmarshal(body, &dictResponse)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("百度翻译结果：")
	for _, m := range dictResponse.TransResult {
		fmt.Println(m["dst"])
	}
	wg.Done()
}

func SumString(content string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(content))
	bys := md5Ctx.Sum(nil)
	//bys := md5.Sum([]byte(content))//这个md5.Sum返回的是数组,不是切片哦
	value := hex.EncodeToString(bys)
	return value
}

func (tran DictRequest2) ToValues() url.Values {
	values := url.Values{
		"q":     {tran.Q},
		"from":  {tran.From},
		"to":    {tran.To},
		"appid": {strconv.Itoa(tran.Appid)},
		"salt":  {strconv.Itoa(tran.Salt)},
		"sign":  {tran.Sign},
	}
	return values
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: simpleDict WORD\nexample: simpleDict hello\n")
		os.Exit(1)
	}
	word := os.Args[1]
	wg.Add(2)
	go query(word)
	go query2(word)
	wg.Wait()
}
