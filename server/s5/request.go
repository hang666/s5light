package s5

import (
	"io"
	"log"
	"net"

	"github.com/txthinking/socks5"
)

// Connect remote conn which u want to connect with your dialer
// Error or OK both replied.
func Connect(w io.Writer, r *socks5.Request, outAddress string) (*net.TCPConn, error) {
	if socks5.Debug {
		log.Println("Call:", r.Address())
	}
	var tmp *net.TCPConn
	var err error

	tcpaddr, err := net.ResolveTCPAddr("tcp", r.Address())
	if err != nil {
		var p *socks5.Reply
		if r.Atyp == socks5.ATYPIPv4 || r.Atyp == socks5.ATYPDomain {
			p = socks5.NewReply(socks5.RepHostUnreachable, socks5.ATYPIPv4, []byte{0x00, 0x00, 0x00, 0x00}, []byte{0x00, 0x00})
		} else {
			p = socks5.NewReply(socks5.RepHostUnreachable, socks5.ATYPIPv6, []byte(net.IPv6zero), []byte{0x00, 0x00})
		}
		if _, err := p.WriteTo(w); err != nil {
			return nil, err
		}
		return nil, err
	}
	if outAddress != "" {
		localAddr, _ := net.ResolveTCPAddr("tcp", outAddress+":0")

		tmp, err = socks5.Dial.DialTCP("tcp", localAddr, tcpaddr)

	} else {
		tmp, err = socks5.Dial.DialTCP("tcp", nil, tcpaddr)
	}
	if err != nil {
		var p *socks5.Reply
		if r.Atyp == socks5.ATYPIPv4 || r.Atyp == socks5.ATYPDomain {
			p = socks5.NewReply(socks5.RepHostUnreachable, socks5.ATYPIPv4, []byte{0x00, 0x00, 0x00, 0x00}, []byte{0x00, 0x00})
		} else {
			p = socks5.NewReply(socks5.RepHostUnreachable, socks5.ATYPIPv6, []byte(net.IPv6zero), []byte{0x00, 0x00})
		}
		if _, err := p.WriteTo(w); err != nil {
			return nil, err
		}
		return nil, err
	}
	rc := tmp
	a, addr, port, err := socks5.ParseAddress(rc.LocalAddr().String())
	if err != nil {
		var p *socks5.Reply
		if r.Atyp == socks5.ATYPIPv4 || r.Atyp == socks5.ATYPDomain {
			p = socks5.NewReply(socks5.RepHostUnreachable, socks5.ATYPIPv4, []byte{0x00, 0x00, 0x00, 0x00}, []byte{0x00, 0x00})
		} else {
			p = socks5.NewReply(socks5.RepHostUnreachable, socks5.ATYPIPv6, []byte(net.IPv6zero), []byte{0x00, 0x00})
		}
		if _, err := p.WriteTo(w); err != nil {
			return nil, err
		}
		return nil, err
	}
	p := socks5.NewReply(socks5.RepSuccess, a, addr, port)
	if _, err := p.WriteTo(w); err != nil {
		return nil, err
	}

	return rc, nil
}
