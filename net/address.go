package net

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"

	"github.com/beaconsoftwarellc/gadget/errors"
	"github.com/beaconsoftwarellc/gadget/stringutil"
)

// Address of a remote host with port information. Does not handle Zones.
type Address struct {
	Host    string
	Port    int
	IsIPv6  bool
	HasPort bool
}

func (addr *Address) String() string {
	s := addr.Host
	if addr.HasPort {
		if addr.IsIPv6 {
			s = fmt.Sprintf("[%s]:%d", addr.Host, addr.Port)
		} else {
			s += ":" + strconv.Itoa(int(addr.Port))
		}
	}
	return s
}

// NewAddressFromConnection if the RemoteAddr is set to a valid address.
func NewAddressFromConnection(conn net.Conn) (*Address, error) {
	return ParseAddress(conn.RemoteAddr().String())
}

// ParseAddress from the passed string and return it.
func ParseAddress(address string) (*Address, errors.TracerError) {
	addr := &Address{}
	if ValidateIPv6Address(address) {
		clean, testPort := cleanIPv6(address)
		hasPort := false
		port := 0
		if testPort > 0 {
			hasPort = true
			port = testPort
		}
		return &Address{Host: clean, Port: port, IsIPv6: true, HasPort: hasPort}, nil
	}
	colons := strings.Count(address, ":")
	if colons > 1 {
		return nil, errors.New("Invalid address: too many colons '%s'", address)
	} else if colons == 0 {
		return &Address{Host: address, HasPort: false}, nil
	}
	split := strings.Split(address, ":")
	addr.Host = split[0]
	port, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, errors.New("address '%s' is invalid: could not parse port data, %s", address, err)
	}
	if port <= 0 || port > math.MaxUint16 {
		return nil, errors.New("port '%d' is not a valid port number, must be uint16", port)
	}
	addr.Port = port
	addr.HasPort = true
	return addr, nil
}

// Network implements net.Addr interface
func (addr *Address) Network() string { return "tcp" }

// MarshalString from the address.
func (addr *Address) MarshalString() (string, error) {
	data, e := json.Marshal(addr)
	return string(data), e
}

// UnmarshalString to an address
func (addr *Address) UnmarshalString(s string) error {
	err := json.Unmarshal([]byte(s), addr)
	if nil == err && stringutil.IsWhiteSpace(addr.Host) && addr.Port < 1 {
		err = errors.New("Invalid Address JSON '%s'", s)
	}
	return err
}

// ensure Addr implments net.Addr
var _ net.Addr = (*Address)(nil)
