// Package fileutil handles the file side of the program: reading, writing,
// atomic replacement, temp files, sorting, and directory listings.
//
// It stays ignorant of target lists and result schemas. Column order and
// directory layout belong to the packages that define those formats.
package fileutil
