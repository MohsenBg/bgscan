// Package resolveprobe checks whether a DNS resolver answers queries.
// ResolverProbe sends the record types in its DNSRequest to the target
// resolver and can run a DPI check first, which queries a name under
// .invalid to spot resolvers that answer for names that cannot exist.
package resolveprobe
