package websvc

import (
	"os"
	"testing"

	"github.com/powerman/narada-go/narada/staging"
	. "gopkg.in/check.v1"
)

func TestMain(m *testing.M) { os.Exit(staging.TearDown(m.Run())) }

func Test(t *testing.T) { TestingT(t) }
