package autoload

import (
	"fmt"
	"os"

	"github.com/stackus/dotenv"
)

func init() {
	if err := dotenv.Load(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "dotenv failed to autoload: %s", err)
		os.Exit(1)
	}
}
