package socks

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

type Config struct {
	User     string
	Password string
}

type Service interface {
	Connect(ctx context.Context, conn net.Conn, target string, config Config) (net.Conn, error)
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) Connect(
	ctx context.Context,
	conn net.Conn,
	target string,
	config Config,
) (net.Conn, error) {
	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return nil, err
		}

		defer func() {
			_ = conn.SetDeadline(time.Time{})
		}()
	}

	if err := authenticate(conn, config); err != nil {
		return nil, fmt.Errorf("SOCKS5 authentication: %w", err)
	}

	if err := connect(conn, target); err != nil {
		return nil, fmt.Errorf("SOCKS5 connect: %w", err)
	}

	return conn, nil
}

func authenticate(conn net.Conn, config Config) error {
	var method byte

	if config.User != "" {
		method = 0x02 // username/password
	} else {
		method = 0x00 // no authentication
	}

	// Greeting:
	//
	// +----+----------+----------+
	// |VER | NMETHODS | METHODS  |
	// +----+----------+----------+
	//
	// VER = 5
	// NMETHODS = 1
	// METHODS = selected method
	_, err := conn.Write([]byte{
		0x05,
		0x01,
		method,
	})
	if err != nil {
		return err
	}

	response := make([]byte, 2)
	if _, err := io.ReadFull(conn, response); err != nil {
		return err
	}

	if response[0] != 0x05 {
		return fmt.Errorf("invalid SOCKS5 version: %d", response[0])
	}

	if response[1] == 0xff {
		return fmt.Errorf("no acceptable authentication method")
	}

	if response[1] != method {
		return fmt.Errorf(
			"unexpected authentication method: 0x%02x",
			response[1],
		)
	}

	if method == 0x02 {
		return usernamePasswordAuth(
			conn,
			config.User,
			config.Password,
		)
	}

	return nil
}

func usernamePasswordAuth(
	conn net.Conn,
	username string,
	password string,
) error {
	if len(username) > 255 {
		return fmt.Errorf("username too long")
	}

	if len(password) > 255 {
		return fmt.Errorf("password too long")
	}

	// RFC 1929:
	//
	// +----+------+----------+------+----------+
	// |VER | ULEN |  UNAME   | PLEN |  PASSWD  |
	// +----+------+----------+------+----------+

	packet := make([]byte, 0, 3+len(username)+len(password))

	packet = append(packet, 0x01)
	packet = append(packet, byte(len(username)))
	packet = append(packet, username...)
	packet = append(packet, byte(len(password)))
	packet = append(packet, password...)

	if _, err := conn.Write(packet); err != nil {
		return err
	}

	response := make([]byte, 2)
	if _, err := io.ReadFull(conn, response); err != nil {
		return err
	}

	if response[0] != 0x01 {
		return fmt.Errorf(
			"invalid authentication version: %d",
			response[0],
		)
	}

	if response[1] != 0x00 {
		return fmt.Errorf("authentication rejected")
	}

	return nil
}

func connect(conn net.Conn, target string) error {
	host, portString, err := net.SplitHostPort(target)
	if err != nil {
		return fmt.Errorf("invalid target %q: %w", target, err)
	}

	port, err := strconv.ParseUint(portString, 10, 16)
	if err != nil {
		return fmt.Errorf("invalid target port: %w", err)
	}

	atyp, address, err := encodeAddress(host)
	if err != nil {
		return err
	}

	request := make([]byte, 0, 7+len(address))

	// SOCKS5 CONNECT request.
	request = append(
		request,
		0x05, // VER
		0x01, // CMD = CONNECT
		0x00, // RSV
		atyp, // ATYP
	)

	request = append(request, address...)

	var portBuf [2]byte
	binary.BigEndian.PutUint16(portBuf[:], uint16(port))
	request = append(request, portBuf[:]...)

	if _, err := conn.Write(request); err != nil {
		return err
	}

	return readConnectResponse(conn)
}

func encodeAddress(host string) (byte, []byte, error) {
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			return 0x01, ip4, nil
		}

		if ip16 := ip.To16(); ip16 != nil {
			return 0x04, ip16, nil
		}
	}

	if len(host) > 255 {
		return 0, nil, fmt.Errorf("hostname too long")
	}

	return 0x03, append([]byte{byte(len(host))}, host...), nil
}

func readConnectResponse(conn net.Conn) error {
	header := make([]byte, 4)

	if _, err := io.ReadFull(conn, header); err != nil {
		return err
	}

	if header[0] != 0x05 {
		return fmt.Errorf("invalid SOCKS5 version: %d", header[0])
	}

	if header[1] != 0x00 {
		return fmt.Errorf(
			"SOCKS5 CONNECT failed: %s",
			replyError(header[1]),
		)
	}

	switch header[3] {
	case 0x01:
		buf := make([]byte, 4)
		if _, err := io.ReadFull(conn, buf); err != nil {
			return err
		}

	case 0x03:
		buf := make([]byte, 1)
		if _, err := io.ReadFull(conn, buf); err != nil {
			return err
		}

		address := make([]byte, buf[0])
		if _, err := io.ReadFull(conn, address); err != nil {
			return err
		}

	case 0x04:
		buf := make([]byte, 16)
		if _, err := io.ReadFull(conn, buf); err != nil {
			return err
		}

	default:
		return fmt.Errorf(
			"invalid SOCKS5 address type: 0x%02x",
			header[3],
		)
	}

	// Destination port.
	var port [2]byte
	_, err := io.ReadFull(conn, port[:])
	return err
}

func replyError(code byte) string {
	switch code {
	case 0x01:
		return "general SOCKS server failure"
	case 0x02:
		return "connection not allowed"
	case 0x03:
		return "network unreachable"
	case 0x04:
		return "host unreachable"
	case 0x05:
		return "connection refused"
	case 0x06:
		return "TTL expired"
	case 0x07:
		return "command not supported"
	case 0x08:
		return "address type not supported"
	default:
		return "unknown error"
	}
}
