package websvc

import (
	_ "log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/powerman/narada-go/narada/staging"
	. "gopkg.in/check.v1"

	"../../cfg"
)

type TestIP struct {
	origRealIPHeader string
}

var _ = Suite(&TestIP{})

func (s *TestIP) SetUpTest(c *C) {
	s.origRealIPHeader = cfg.HTTP.RealIPHeader
}

func (s *TestIP) TearDownTest(c *C) {
	cfg.HTTP.RealIPHeader = s.origRealIPHeader
}

func (s *TestIP) Test(c *C) {
	r, err := http.NewRequest("GET", "http://websvc.test/", nil)
	c.Assert(err, IsNil)
	r.RemoteAddr = "1.2.3.4:0"
	r.Header.Set("X-Real-IP", "4.3.2.1")

	cfg.HTTP.RealIPHeader = ""
	c.Check(remoteIP(r), Equals, "1.2.3.4")
	cfg.HTTP.RealIPHeader = "X-Real-IP"
	c.Check(remoteIP(r), Equals, "4.3.2.1")
}

func (s *TestIP) TestCrasher(c *C) {
	crasher(c, 1, ".*failed to detect remote IP.*\n", func() {
		r, err := http.NewRequest("GET", "http://websvc.test/", nil)
		if err != nil {
			log.Fatal(err)
		}
		r.RemoteAddr = "1.2.3.4:0"
		r.Header.Set("X-Real-IP", "4.3.2.1")
		cfg.HTTP.RealIPHeader = "X-Real-REMOTE_ADDR"
		remoteIP(r)
	})
}

func crasher(c *C, exitCode int, stderrRegex string, test func()) {
	if os.Getenv("BE_CRASHER") == "1" {
		test()
		return
	}
	pc, _, _, _ := runtime.Caller(1)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	caller := parts[len(parts)-1]
	cmd := exec.Command(os.Args[0], "-check.f=^"+caller+"$")
	cmd.Dir = staging.BaseDir
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	_, err := cmd.Output()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		c.Check(e.Sys().(syscall.WaitStatus).ExitStatus(), Equals, exitCode)
		c.Check(string(e.Stderr), Matches, stderrRegex)
	} else {
		c.Errorf("process ran with err %v, want exit status 1", err)
	}
}
