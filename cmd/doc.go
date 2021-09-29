// Package cmd provides an opiniated interface for implementing Axiom based
// tools and utilities. Those tools share the same configuration, logging and
// application lifecycle behaviour. This results in an easy to use and
// understandable set of application.
//
// In the most basic case, applications should pass their name and an
// implementation of `cmd.MainFunc` to `cmd.Run()`:
//
//   package main
//
//   import (
//       "context"
//
//       "github.com/axiomhq/pkg/cmd"
//   )
//
//   func main() {
//       cmd.Run("my-app", Main)
//   }
//
//   func Main(ctx context.Context, f *cmd.Factory) error {
//       f.Logger().Info("hello, world!")
//       return nil
//   }
//
package cmd
