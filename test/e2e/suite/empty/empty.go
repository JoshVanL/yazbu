package empty

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/joshvanl/yazbu/e2e/framework"
)

var _ = framework.CasesDescribe("Empty database should list", func() {
	f := framework.NewDefaultFramework("foo")

	It("should list empty database", func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		By("First list should write database files")
		out, err := f.Helper().YazbuList(ctx, f.Helper().YazbuDefaultConfig())
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(MatchRegexp(`9090/joshvanl-test-1/client/yazbu/testing-2: "level"=0 "msg"="db file does not exist, writing" "db_file"="joshvanl-test-1/yazbu/testing-2/backup.db`))
		Expect(out).To(MatchRegexp(`9090/joshvanl-test-2/client/yazbu/testing-2: "level"=0 "msg"="db file does not exist, writing" "db_file"="joshvanl-test-2/yazbu/testing-2/backup.db`))
		Expect(out).To(MatchRegexp(`9090/joshvanl-test-2/client/yazbu/testing-1: "level"=0 "msg"="db file does not exist, writing" "db_file"="joshvanl-test-2/yazbu/testing-1/backup.db`))
		Expect(out).To(MatchRegexp(`9090/joshvanl-test-1/client/yazbu/testing-1: "level"=0 "msg"="db file does not exist, writing" "db_file"="joshvanl-test-1/yazbu/testing-1/backup.db`))
		Expect(out).To(MatchRegexp(`9091/joshvanl-test-2/client/yazbu/testing-2: "level"=0 "msg"="db file does not exist, writing" "db_file"="joshvanl-test-2/yazbu/testing-2/backup.db`))
		Expect(out).To(MatchRegexp(`9091/joshvanl-test-2/client/yazbu/testing-1: "level"=0 "msg"="db file does not exist, writing" "db_file"="joshvanl-test-2/yazbu/testing-1/backup.db`))
		Expect(out).To(MatchRegexp(`9091/joshvanl-test-1/client/yazbu/testing-2: "level"=0 "msg"="db file does not exist, writing" "db_file"="joshvanl-test-1/yazbu/testing-2/backup.db`))
		Expect(out).To(MatchRegexp(`9091/joshvanl-test-1/client/yazbu/testing-1: "level"=0 "msg"="db file does not exist, writing" "db_file"="joshvanl-test-1/yazbu/testing-1/backup.db`))

		By("Second list should only list empty databases")
		out, err = f.Helper().YazbuList(ctx, f.Helper().YazbuDefaultConfig())
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(Equal("DATASET    ENDPOINT    BUCKET    ID    PARENT    TYPE    PATH    SIZE    TIMESTAMP\n"))
	})
})
