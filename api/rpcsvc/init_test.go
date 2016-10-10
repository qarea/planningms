package rpcsvc

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/powerman/narada-go/narada/staging"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().Unix())
	os.Exit(staging.TearDown(m.Run()))
}

//func Test(t *testing.T) { TestingT(t) }
