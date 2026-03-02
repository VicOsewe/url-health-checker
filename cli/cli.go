// Package cli defines the interface that any CLI implementation must satisfy.

package cli

// CLI is the interface every CLI adapter must implement.
type CLI interface {
	Run(args []string) error
}
