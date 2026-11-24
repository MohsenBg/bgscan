// Package target reads, writes, and streams bgscan target lists.
//
// A list is a CSV file of Entry rows in the form value,port,path,enable.
// The package splits mixed lists into separate IP and domain files,
// deduplicates entries on import, and streams enabled entries to a
// scan in file order or in shuffled order.
package target
