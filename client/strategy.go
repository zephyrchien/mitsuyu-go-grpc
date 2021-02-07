package client

import (
	"github.com/ZephyrChien/Mitsuyu/common"
	"net"
	"strconv"
	"strings"
)

func splitRules(str, sep string) []string {
	ss := make([]string, 0, 4)
	for _, s := range strings.Split(str, sep) {
		ss = append(ss, strings.TrimSpace(s))
	}
	return ss
}

func blockReservedAddr(addr *common.Addr) (block bool) {
	if addr.Isdn && addr.Host != "localhost" {
		return false
	}
	if addr.Isdn && addr.Host == "localhost" {
		return true
	}
	x := net.ParseIP(addr.Host)
	if x == nil {
		return true
	}
	if x.IsInterfaceLocalMulticast() ||
		x.IsLinkLocalMulticast() || x.IsLinkLocalUnicast() ||
		x.IsLoopback() || x.IsMulticast() || x.IsUnspecified() {
		return true
	}
	return false
}

func matchIPRange(ip, ipRange string) bool {
	ipBinary := net.ParseIP(ip)
	if ipBinary == nil {
		return false
	}
	for _, ips := range splitRules(ipRange, ",") {
		_, ipnet, err := net.ParseCIDR(ips)
		if err != nil {
			continue
		}

		if ipnet.Contains(ipBinary) {
			return true
		}
	}
	return false
}

func matchPortRange(port, portRange string) bool {
	portInt, _ := strconv.Atoi(port)
	for _, ps := range splitRules(portRange, ",") {
		if strings.Contains(ps, "-") {
			pps := splitRules(ps, "-")
			ppsIntMin, _ := strconv.Atoi(pps[0])
			ppsIntMax, _ := strconv.Atoi(pps[1])
			if portInt >= ppsIntMin && portInt <= ppsIntMax {
				return true
			}
		} else if port == ps {
			return true
		}
	}
	return false
}

func matchDomainPrefix(dname, prefix string) bool {
	for _, pre := range splitRules(prefix, ",") {
		if strings.HasPrefix(dname, pre) {
			return true
		}
	}
	return false
}

func matchDomainSuffix(dname, suffix string) bool {
	for _, end := range splitRules(suffix, ",") {
		if strings.HasSuffix(dname, end) {
			return true
		}
	}
	return false
}

func matchDomainContain(dname, subdname string) bool {
	for _, sub := range splitRules(subdname, ",") {
		if strings.Contains(dname, sub) {
			return true
		}
	}
	return false
}

func matchRules(addr *common.Addr, rules *common.Strategy) bool {
	if !addr.Isdn && rules.IPRange != "" && matchIPRange(addr.Host, rules.IPRange) {
		return true
	}
	if rules.PortRange != "" && matchPortRange(addr.Port, rules.PortRange) {
		return true
	}
	if rules.DomainContain != "" && matchDomainContain(addr.Host, rules.DomainContain) {
		return true
	}
	if rules.DomainSuffix != "" && matchDomainSuffix(addr.Host, rules.DomainSuffix) {
		return true
	}
	if rules.DomainPrefix != "" && matchDomainPrefix(addr.Host, rules.DomainPrefix) {
		return true
	}
	return false
}
