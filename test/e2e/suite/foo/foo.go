package foo

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/joshvanl/yazbu/test/e2e/framework"
)

var _ = framework.CasesDescribe("foo", func() {
	_ = framework.NewDefaultFramework("foo")
	It("barr", func() {
	})
})
