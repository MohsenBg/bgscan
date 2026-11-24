// Package thefeedprobe checks whether a target IP works as a DNS
// resolver for thefeed content. TheFeedProbe sends one encrypted
// metadata TXT query through the target and checks that the reply
// decrypts with the configured passphrase. TheFeedResult is the
// result type, and Schema describes its stored record layout.
package thefeedprobe
