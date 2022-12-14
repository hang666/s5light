package server

import (
	"log"
	"sync"

	"github.com/txthinking/socks5"
)

var wg sync.WaitGroup

func Server() {
	listenAndServe()
	wg.Wait()
}

func listenAndServe() {
	for _, acc := range Accounts {
		wg.Add(1)
		go func(acc *AccountStruct) {
			log.Printf("Server listening on %s", acc.BindAddress)
			socks5.Debug = true
			server, err := socks5.NewClassicServer(acc.BindAddress, acc.UDPBindIP, acc.Username, acc.Password, acc.TCPTimeout, acc.UDPTimeout)
			if err != nil {
				log.Panicln(err)
			}
			server.ListenAndServe(&DefaultHandle{acc.WhitelistMap, acc.OutAddress})
			wg.Done()
		}(acc)
	}
}
