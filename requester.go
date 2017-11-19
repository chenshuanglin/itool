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
	out := make(chan string, 1024)

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
	rand.Seed(time.Now().UnixNano())
	go func() {
		for {
			out <- rand.Intn(65534-1024) + 1024
		}
	}()
	return out
}

//获取客户端连接服务
func (w *Work) http_client() chan *http.Client {

	out := make(chan *http.Client, 1024)
	go func() {
		for {
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
				DisableKeepAlives: true,
			}

			c := &http.Client{
				Transport: tr,
			}

			out <- c
		}
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

	//多久时间打印一次结果
	time time.Duration

	//当前源地址
	ip chan string

	//请求的源端口
	port chan int

	//当前请求客户端
	c chan *http.Client

	//请求的url
	url string

	//失败次数累加和
	failedNumber int64

	//请求次数累加和
	requestNum int64
}

func (w *Work) Run() {
	tick := time.Tick(w.time)

	w.ip = ip_generate(w.ipaddrs)
	w.port = rand_generate()
	w.c = w.http_client()

	go func() {
		w.runWorker()
	}()

loop:
	for {
		select {
		case <-tick:
			fmt.Println("当前请求总数:", w.requestNum)
			fmt.Println("当前失败数:", w.failedNumber)
		case <-w.stop:
			fmt.Println("Good bye\n")
			close(w.done)
			time.Sleep(1 * time.Second)
			break loop
		}
	}
}

func (w *Work) runWorker() {
	//计算多久发出一次请求
	tick := time.NewTicker(time.Duration(1e6/w.n) * time.Microsecond)
loop:
	for {
		select {
		case <-tick.C:
			w.requestNum++
			//发起一次请求
			go w.makeRequest()
		case <-w.done:
			tick.Stop()
			fmt.Println("Stop request\n")
			break loop
		}
	}
}

func (w *Work) makeRequest() {

	c := <-w.c
	resp, err := c.Get(w.url)
	if err != nil {
		w.failedNumber++
		go func() {
			fmt.Println(err)
		}()
		return
	}

	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
}

func (w *Work) getHttpClient() *http.Client {
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
		DisableKeepAlives: true,
	}

	c := &http.Client{
		Transport: tr,
	}
	return c
}
