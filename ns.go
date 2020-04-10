package ns

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

type Entry struct {
	Number     int
	LocalAddr  net.IP
	LocalPort  int
	RemoteAddr net.IP
	RemotePort int
	State      int
	UID        int
	INode      int
}

type Entries []Entry

func (es Entries) FindByLocalPort(port int) int {
	for i, e := range es {
		if e.LocalPort == port {
			return i
		}
	}
	return -1
}

func (es Entries) FindByRemotePort(port int) int {
	for i, e := range es {
		if e.RemotePort == port {
			return i
		}
	}
	return -1
}

func (es Entries) Filter(f func(e Entry) bool) Entries {
	r := make([]Entry, 0, len(es))
	for _, e := range es {
		if f(e) {
			r = append(r, e)
		}
	}
	return r
}

func parseAddr(addr string) (ip net.IP, port int, err error) {
	if len(addr) == 13 {
		ip = make([]byte, 4, 4)
		_, err = fmt.Sscanf(
			addr,
			"%02x%02x%02x%02x:%04x",
			&ip[0], &ip[1], &ip[2], &ip[3],
			&port,
		)
		if err != nil {
			return net.IP{}, 0, err
		}
		return ip, port, nil
	}
	ip = make([]byte, 16, 16)
	_, err = fmt.Sscanf(
		addr,
		"%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x:%04x",
		&ip[0], &ip[1], &ip[2], &ip[3], &ip[4], &ip[5], &ip[6], &ip[7],
		&ip[8], &ip[9], &ip[10], &ip[11], &ip[12], &ip[13], &ip[14], &ip[15],
		&port,
	)
	if err != nil {
		return net.IP{}, 0, err
	}
	return ip, port, nil
}

func parseState(st string) (int, error) {
	var state int
	_, err := fmt.Sscanf(st, "%02x", &state)
	if err != nil {
		return 0, err
	}
	return state, nil
}

func Parse(r io.Reader) (Entries, error) {
	var es []Entry
	sc := bufio.NewScanner(r)
	sc.Scan() // ignore header
	for sc.Scan() {
		var e Entry
		var local, remote, st, skip string
		_, err := fmt.Sscanf(
			sc.Text(),
			"%d: %s %s %s %s %s %s %d %s %d",
			&e.Number,
			&local,
			&remote,
			&st,
			&skip, // tx_queue:rx_queue
			&skip, // tr tm->when
			&skip, // retrnsmt
			&e.UID,
			&skip, // timeout
			&e.INode,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to parse first step: %w", err)
		}
		e.LocalAddr, e.LocalPort, err = parseAddr(local)
		if err != nil {
			return nil, fmt.Errorf("failed to parse local address: %w", err)
		}
		e.RemoteAddr, e.RemotePort, err = parseAddr(remote)
		if err != nil {
			return nil, fmt.Errorf("failed to parse remote address: %w", err)
		}
		e.State, err = parseState(st)
		if err != nil {
			return nil, fmt.Errorf("failed to parse state: %w", err)
		}
		es = append(es, e)
	}
	if sc.Err() != nil {
		return nil, fmt.Errorf("failed to read text: %w", sc.Err())
	}
	return Entries(es), nil
}
