package e2e

import (
	"flag"

	. "github.com/onsi/ginkgo/v2"

	"github.com/joshvanl/yazbu/e2e/framework/config"
)

var (
	cfg = config.GetConfig()
)

func init() {
	cfg.AddFlags(flag.CommandLine)
}

var _ = SynchronizedBeforeSuite(func() []byte {
	return nil
}, func([]byte) {
})

//var _ = SynchronizedAfterSuite(func() {})
