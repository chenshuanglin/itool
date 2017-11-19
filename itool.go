package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"
)

//参数
var (
	u = flag.String("u", "http://127.0.0.1", "访问的url,默认值为http://127.0.0.1")
	n = flag.Int64("n", 1, "最大新建连接数，默认最小值为1")
	s = flag.String("s", "", "默认值为空，必须指定源IP范围，格式为：64.233.196.0-64.233.196.25")
	t = flag.Duration("t", 1*time.Second, "间隔多久打印一次数据，默认值为1s")
)

var usage = `Usage: itool [option...]
option:
	-u    访问的url,默认值为http://127.0.0.1
	-n    最大新建连接数，默认最小值为1
	-s    默认值为空，可指定源范围，格式为：64.233.196.0-64.233.196.25
	-t    间隔多久打印一次数据，默认值为1s
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	flag.Parse()

	if *t < 1*time.Second || *s == "" {
		flag.Usage()
		os.Exit(1)
	}

	var stop = make(chan os.Signal, 1)
	//捕捉ctrl + c 信号
	signal.Notify(stop, os.Interrupt)

	(&Work{
		n:            *n,
		stop:         stop,
		done:         make(chan struct{}, 1),
		ipaddrs:      Gen_ip(*s),
		time:         *t,
		url:          *u,
		failedNumber: 0,
	}).Run()

}
