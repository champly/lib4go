package ping

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

var deadline = 10

func Send(addr string) (microSecond int64, err error) {

	icmp := GetDefaultICMP(1)
	return SendWithICMP(icmp, addr)
}

// func SendWithTimeOut(addr string, timeout time)

func SendWithICMP(icmp ICMP, addr string) (microSecond int64, err error) {

	conn, err := net.DialIP("ip4:icmp", nil, &net.IPAddr{IP: net.ParseIP(addr)})
	if err != nil {
		fmt.Printf("Fail to connect to remote host: %s\n", err)
		return 0, err
	}
	defer conn.Close()
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp)
	if _, err := conn.Write(buffer.Bytes()); err != nil {
		return 0, err
	}
	tStart := time.Now()
	conn.SetReadDeadline((time.Now().Add(time.Second * time.Duration(deadline))))
	recv := make([]byte, 1024)
	receiveCnt, err := conn.Read(recv)
	if err != nil {
		return 0, fmt.Errorf("%s, deadline:%ds", err.Error(), deadline)
	}
	tEnd := time.Now()
	duration := tEnd.Sub(tStart).Microseconds()
	fmt.Printf("%d bytes from %s: seq=%d time=%sms\n", receiveCnt, addr, icmp.SequenceNum, getMs(duration))
	return duration, nil
}

func getMs(t int64) string {
	ms := t / 1000
	ws := t % 1000
	if ws > 99 {
		return fmt.Sprintf("%d.%d", ms, ws)
	}

	if ws > 9 {
		return fmt.Sprintf("%d.0%d", ms, ws)
	}

	if ws > 0 {
		return fmt.Sprintf("%d.00%d", ms, ws)
	}

	return fmt.Sprintf("%d", ms)
}
