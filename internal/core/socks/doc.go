// Package socks is a small SOCKS5 client. Service.Connect runs the
// SOCKS5 greeting, optional username and password authentication,
// and the CONNECT request over a connection the caller already has,
// such as a tunnel connection. On success the same connection is
// ready to carry traffic to the target.
package socks
