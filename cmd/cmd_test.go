package cmd_test

import (
	"context"
	"fmt"
	"os"

	"github.com/axiomhq/axiom-go/axiom"
	"go.uber.org/zap"

	"github.com/axiomhq/pkg/cmd"
)

func Example() {
	os.Clearenv()
	os.Setenv("DEBUG", "1")

	mainFunc := func(_ context.Context, _ *zap.Logger, _ *axiom.Client) error {
		// All your actual application code goes here! See doc.go for more info.

		fmt.Print("Hello World!")

		return nil
	}

	cmd.Run("example", mainFunc,
		cmd.WithAxiomOptions(
			axiom.SetNoEnv(),
			axiom.SetURL("http://axiom.local"),
			axiom.SetAccessToken("xapt-1234"),
		),
	)

	// Output:
	// Hello World!
}
