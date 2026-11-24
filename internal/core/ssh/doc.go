// Package ssh connects to SSH servers over connections the caller
// already has, such as tunnel connections. SSHService performs the
// handshake and authentication, and returns dial functions that
// open connections through the SSH client. Host keys are checked
// against a known_hosts file unless the caller opts out in
// SSHConfig.
package ssh
