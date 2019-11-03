package util

import (
	"net"
	"net/http"
	"strings"
	"io/ioutil"
)

// InternalIP return internal ip.
func InternalIP() string {
	inters, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, inter := range inters {
		if inter.Flags&net.FlagUp != 0 && !strings.HasPrefix(inter.Name, "lo") {
			addrs, err := inter.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String()
					}
				}
			}
		}
	}
	return ""
}

func OutsiteIP() string{
	oIP, err := http.Get("http://ip.cip.cc/")
	if err !=nil{
		return ""
	}
	body, _ := ioutil.ReadAll(oIP.Body)
	if body != nil{
		return strings.TrimSpace(string(body))
	}
	return ""
}
