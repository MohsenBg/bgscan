// Package migration upgrades on-disk layouts to the current format.
//
// The main entry point is MigrationService, created with New.
// MigrateTargets moves old IP lists into the target directory and
// rewrites them in the current layout. A format marker file records
// the layout version of a directory so a finished upgrade is skipped
// on later runs.
package migration
