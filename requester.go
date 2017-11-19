package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"
)

//持续迭代数组中的IP地址
func ip_generate(ips []string) chan string {
	out := make(chan string)

	go func() {
		i := 0
		length := len(ips)
		for {
			if i >= length {
				i = 0
			}
			out <- ips[i]
			i++
		}
	}()

	return out
}

//获取随机端口
func rand_generate() chan int {
	out := make(chan int, 1024)
	go func() {
		rand.Seed(time.Now().UnixNano())
		out <- rand.Intn(65534-1024) + 1024
	}()
	return out
}

type Work struct {
	//设置请求任务数
	n int64

	//停止信号
	stop chan os.Signal

	//结束信号
	done chan struct{}

	//源地址数组
	ipaddrs []string

	//当前源地址
	ip chan string

	//请求的源端口
	port chan int

	//多久时间打印一次结果
	time time.Duration

	//请求的url
	url string

	//失败次数累加和
	failedNumber int
}

func (w *Work) Run() {
	tick := time.Tick(w.time)

	w.ip = ip_generate(w.ipaddrs)
	w.port = rand_generate()

	go func() {
		w.runWorker()
	}()

loop:
	for {
		select {
		case <-tick:
			fmt.Println("当前失败数:", w.failedNumber)
		case <-w.stop:
			fmt.Println("Good bye\n")
			close(w.done)
			break loop
		}
	}
}

func (w *Work) runWorker() {
	var throttle <-chan time.Time

	//计算多久发出一次请求
	throttle = time.Tick(time.Duration(1e6/w.n) * time.Microsecond)
loop:
	for {
		select {
		case <-throttle:
			//发起一次请求
			go w.makeRequest()
		case <-w.done:
			fmt.Println("Stop request\n")
			break loop
		}
	}
}

func (w *Work) makeRequest() {
	c := w.getHttpClient()
	resp, err := c.Get(w.url)
	if err != nil {
		w.failedNumber++
		return
	}

	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
}

func (w *Work) getHttpClient() *http.Client {
	go func() {
		fmt.Println("卡在这了")
	}()
	//获取源地址
	localAddr, err := net.ResolveIPAddr("ip", <-w.ip)
	if err != nil {
		panic(err)
	}

	localTCPAddr := net.TCPAddr{
		IP:   localAddr.IP,
		Port: <-w.port,
	}

	tr := &http.Transport{
		Dial:              (&net.Dialer{LocalAddr: &localTCPAddr, Timeout: 2 * time.Second}).Dial,
		DisableKeepAlives: false,
	}

	c := &http.Client{
		Transport: tr,
	}
	return c
}
