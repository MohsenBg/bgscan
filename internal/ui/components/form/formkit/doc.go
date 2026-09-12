// Package formkit provides shared building blocks for the DNS tunnel
// configuration forms. It removes the boilerplate that every protocol form
// (vaydns, dnstt, slipstream, masterdns, stormdns, thefeed) used to repeat:
// the component lifecycle, the generic save flow and the field builders with
// overflow-safe parsing and inline validation.
package formkit
