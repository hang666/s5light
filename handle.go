package s5light

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/txthinking/socks5"
)

type DefaultHandle struct {
	whitelistMap WhitelistMapType
}

func checkIsWhitelisted(address string, whitelistMap WhitelistMapType) bool {
	//log.Printf("client come in: %s", address)
	w_map := whitelistMap
	if len(w_map) == 0 {
		return true
	}
	var ok bool
	if strings.Contains(address, ":") {
		_, ok = w_map[strings.Split(address, ":")[0]]
	} else {
		_, ok = w_map[address]
	}
	return ok
}

func (h *DefaultHandle) TCPHandle(s *socks5.Server, c *net.TCPConn, r *socks5.Request) error {
	if !checkIsWhitelisted(c.RemoteAddr().String(), h.whitelistMap) {
		return fmt.Errorf("%s is not whitelisted", c.RemoteAddr().String())
	}
	if r.Cmd == socks5.CmdConnect {
		rc, err := r.Connect(c)
		if err != nil {
			return err
		}
		defer rc.Close()
		go func() {
			var bf [1024 * 2]byte
			for {
				if s.TCPTimeout != 0 {
					if err := rc.SetDeadline(time.Now().Add(time.Duration(s.TCPTimeout) * time.Second)); err != nil {
						return
					}
				}
				i, err := rc.Read(bf[:])
				if err != nil {
					return
				}
				if _, err := c.Write(bf[0:i]); err != nil {
					return
				}
			}
		}()
		var bf [1024 * 2]byte
		for {
			if s.TCPTimeout != 0 {
				if err := c.SetDeadline(time.Now().Add(time.Duration(s.TCPTimeout) * time.Second)); err != nil {
					return nil
				}
			}
			i, err := c.Read(bf[:])
			if err != nil {
				return nil
			}
			if _, err := rc.Write(bf[0:i]); err != nil {
				return nil
			}
		}
		return nil
	}
	if r.Cmd == socks5.CmdUDP {
		caddr, err := r.UDP(c, s.ServerAddr)
		if err != nil {
			return err
		}
		ch := make(chan byte)
		defer close(ch)
		s.AssociatedUDP.Set(caddr.String(), ch, -1)
		defer s.AssociatedUDP.Delete(caddr.String())
		io.Copy(io.Discard, c)
		if socks5.Debug {
			log.Printf("A tcp connection that udp %#v associated closed\n", caddr.String())
		}
		return nil
	}
	return socks5.ErrUnsupportCmd
}

func (h *DefaultHandle) UDPHandle(s *socks5.Server, addr *net.UDPAddr, d *socks5.Datagram) error {
	if !checkIsWhitelisted(string(addr.IP), h.whitelistMap) {
		return fmt.Errorf("%s is not whitelisted", string(addr.IP))
	}
	src := addr.String()
	var ch chan byte
	if s.LimitUDP {
		any, ok := s.AssociatedUDP.Get(src)
		if !ok {
			return fmt.Errorf("This udp address %s is not associated with tcp", src)
		}
		ch = any.(chan byte)
	}
	send := func(ue *socks5.UDPExchange, data []byte) error {
		select {
		case <-ch:
			return fmt.Errorf("This udp address %s is not associated with tcp", src)
		default:
			_, err := ue.RemoteConn.Write(data)
			if err != nil {
				return err
			}
			if socks5.Debug {
				log.Printf("Sent UDP data to remote. client: %#v server: %#v remote: %#v data: %#v\n", ue.ClientAddr.String(), ue.RemoteConn.LocalAddr().String(), ue.RemoteConn.RemoteAddr().String(), data)
			}
		}
		return nil
	}

	dst := d.Address()
	var ue *socks5.UDPExchange
	iue, ok := s.UDPExchanges.Get(src + dst)
	if ok {
		ue = iue.(*socks5.UDPExchange)
		return send(ue, d.Data)
	}

	if socks5.Debug {
		log.Printf("Call udp: %#v\n", dst)
	}
	var laddr *net.UDPAddr
	any, ok := s.UDPSrc.Get(src + dst)
	if ok {
		laddr = any.(*net.UDPAddr)
	}
	raddr, err := net.ResolveUDPAddr("udp", dst)
	if err != nil {
		return err
	}
	rc, err := socks5.Dial.DialUDP("udp", laddr, raddr)
	if err != nil {
		if !strings.Contains(err.Error(), "address already in use") {
			return err
		}
		rc, err = socks5.Dial.DialUDP("udp", nil, raddr)
		if err != nil {
			return err
		}
		laddr = nil
	}
	if laddr == nil {
		s.UDPSrc.Set(src+dst, rc.LocalAddr().(*net.UDPAddr), -1)
	}
	ue = &socks5.UDPExchange{
		ClientAddr: addr,
		RemoteConn: rc,
	}
	if socks5.Debug {
		log.Printf("Created remote UDP conn for client. client: %#v server: %#v remote: %#v\n", addr.String(), ue.RemoteConn.LocalAddr().String(), d.Address())
	}
	if err := send(ue, d.Data); err != nil {
		ue.RemoteConn.Close()
		return err
	}
	s.UDPExchanges.Set(src+dst, ue, -1)
	go func(ue *socks5.UDPExchange, dst string) {
		defer func() {
			ue.RemoteConn.Close()
			s.UDPExchanges.Delete(ue.ClientAddr.String() + dst)
		}()
		var b [65507]byte
		for {
			select {
			case <-ch:
				if socks5.Debug {
					log.Printf("The tcp that udp address %s associated closed\n", ue.ClientAddr.String())
				}
				return
			default:
				if s.UDPTimeout != 0 {
					if err := ue.RemoteConn.SetDeadline(time.Now().Add(time.Duration(s.UDPTimeout) * time.Second)); err != nil {
						log.Println(err)
						return
					}
				}
				n, err := ue.RemoteConn.Read(b[:])
				if err != nil {
					return
				}
				if socks5.Debug {
					log.Printf("Got UDP data from remote. client: %#v server: %#v remote: %#v data: %#v\n", ue.ClientAddr.String(), ue.RemoteConn.LocalAddr().String(), ue.RemoteConn.RemoteAddr().String(), b[0:n])
				}
				a, addr, port, err := socks5.ParseAddress(dst)
				if err != nil {
					log.Println(err)
					return
				}
				d1 := socks5.NewDatagram(a, addr, port, b[0:n])
				if _, err := s.UDPConn.WriteToUDP(d1.Bytes(), ue.ClientAddr); err != nil {
					return
				}
				if socks5.Debug {
					log.Printf("Sent Datagram. client: %#v server: %#v remote: %#v data: %#v %#v %#v %#v %#v %#v datagram address: %#v\n", ue.ClientAddr.String(), ue.RemoteConn.LocalAddr().String(), ue.RemoteConn.RemoteAddr().String(), d1.Rsv, d1.Frag, d1.Atyp, d1.DstAddr, d1.DstPort, d1.Data, d1.Address())
				}
			}
		}
	}(ue, dst)
	return nil
}
