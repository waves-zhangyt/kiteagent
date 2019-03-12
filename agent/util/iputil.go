// Created By ytzhang0828@qq.com
// Use of this source code is governed by a Apache-2.0 LICENSE

/*
   util包中的ip工具
*/
package util

import (
	"net"
	"strings"
)

// 获取本机ip地址
func LocalIPv4Addr() string {
	addrs, _ := net.InterfaceAddrs()

	var ipaddrs string
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipaddr := ipnet.String()
				if strings.Contains(ipaddr, "/") {
					ipaddr = ipaddr[0:strings.Index(ipaddr, "/")]
				}
				ipaddrs += ipaddr + ","
			}
		}
	}

	return ipaddrs[0 : len(ipaddrs)-1]
}
