package redis

import (
	"fmt"
	"net"
	"strings"
)

func Uint32Kaxap(k uint32) string {
	return fmt.Sprintf("4%x", k)
}

func IPKaxap(k net.IP) string {
	k = k.To4()
	return fmt.Sprintf("4%x%x%x%x", k[0], k[1], k[2], k[3])
}

func KaxapMSISDN(v string) string {
	return strings.SplitN(v, "|", 2)[0]
}

func Msisdn(m *Mirror, input interface{}) (string, error) {
	var key string
	if ip, ok := input.(net.IP); ok {
		key = IPKaxap(ip)
	} else if u32, ok := input.(uint32); ok {
		key = Uint32Kaxap(u32)
	} else {
		return "", ErrNotFound
	}

	s, err := m.Get(key)
	if err != nil {
		return "", err
	}
	return KaxapMSISDN(s.(string)), err
}
