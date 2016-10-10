package rpcsvc

import (
	"testing"

	"github.com/powerman/narada-go/narada"
)

func TestVERSION(t *testing.T) {
	api := &API{}
	var args struct{}
	var res string
	err := api.Version(&args, &res)
	if err != nil {
		t.Errorf("Version(), err = %v", err)
	}
	if ver, _ := narada.Version(); res != ver {
		t.Errorf("Version() = %v, want %v", res, ver)
	}
}
