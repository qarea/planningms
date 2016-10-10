//  +build integration

package storage

import (
	"log"
	"math/rand"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/powerman/narada-go/narada/staging"

	_ "github.com/go-sql-driver/mysql"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().Unix())
	os.Exit(staging.TearDown(m.Run()))
}

type cleanupFunc func()

func prepareDB() cleanupFunc {
	err := exec.Command("narada-setup-mysql").Run()
	if err != nil {
		log.Fatalln("narada-setup-mysql failed: ", err)
	}
	return cleanup
}

func cleanup() {
	err := exec.Command("narada-setup-mysql", "--clean").Run()
	if err != nil {
		log.Fatalln("narada-setup-mysql --clean failed ", err)
	}
}
