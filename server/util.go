package server

import (
	"net"
)

func checkIsWhitelisted(address string, whitelistMap WhitelistMapType) bool {
	//log.Printf("client come in: %s", address)
	w_map := whitelistMap
	if len(w_map) == 0 {
		return true
	}
	var ok bool
	addr, _, err := net.SplitHostPort(address)
	if err != nil {
		addr = address
	}
	_, ok = w_map[addr]
	return ok
}
