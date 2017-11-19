package main

import (
	"fmt"
	"strconv"
	"strings"
)

func ip2Num(ip string) int {
	ips := strings.Split(ip, ".")
	ip1, _ := strconv.Atoi(ips[0])
	ip2, _ := strconv.Atoi(ips[1])
	ip3, _ := strconv.Atoi(ips[2])
	ip4, _ := strconv.Atoi(ips[3])
	return ip1<<24 | ip2<<16 | ip3<<8 | ip4
}

func num2Ip(num int) string {
	return fmt.Sprintf("%d.%d.%d.%d", num>>24&0xff, num>>16&0xff, num>>8&0xff, num&0xff)
}

func Gen_ip(ipRange string) []string {
	var ips []string
	ipArr := strings.Split(ipRange, "-")
	start, end := ip2Num(ipArr[0]), ip2Num(ipArr[1])
	for ; start <= end; start++ {
		ips = append(ips, num2Ip(start))
	}
	return ips
}
