package cache

import (
	"context"
	"errors"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/powerman/narada-go/narada/staging"
	"github.com/qarea/ctxtg"
	"github.com/qarea/planningms/entities"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().Unix())
	os.Exit(staging.TearDown(m.Run()))
}

var ctx = context.Background()

func TestLoadFromDump(t *testing.T) {
	dir, clean := tempDir(t)
	defer clean()
	s1 := randSpentTime()
	s2 := randSpentTime()
	storage1 := SpentTimeInMemory{}
	storage1.backupFolder = dir
	storage1.spentTime = make(map[ctxtg.UserID]entities.SpentTime)
	storage1.spentTime[s1.UserID] = s1
	storage1.spentTime[s2.UserID] = s2
	err := storage1.backup()
	if err != nil {
		t.Fatal(err)
	}

	storage2 := NewSpentTimeInMemory(10*time.Hour, dir, 1*time.Second)
	storage2.backupFolder = dir
	storage2.spentTime = make(map[ctxtg.UserID]entities.SpentTime)
	err = storage2.restore()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(storage1.spentTime, storage2.spentTime) {
		t.Error("Invalid backup loaded")
	}
}

func TestSuccessfullyLoadWithOutDump(t *testing.T) {
	dir, clean := tempDir(t)
	defer clean()
	storage := NewSpentTimeInMemory(10*time.Hour, dir, 1*time.Second)
	if storage == nil {
		t.Error("Should be ok")
	}
	if len(storage.spentTime) > 0 {
		t.Error("Should be empty")
	}
}

func TestLoadFromDumpLoadFromTmpIfNotExist(t *testing.T) {
	dir, clean := tempDir(t)
	defer clean()
	s1 := randSpentTime()
	s2 := randSpentTime()
	storage1 := SpentTimeInMemory{}
	storage1.backupFolder = dir
	storage1.spentTime = make(map[ctxtg.UserID]entities.SpentTime)
	storage1.spentTime[s1.UserID] = s1
	storage1.spentTime[s2.UserID] = s2
	err := storage1.backup()
	if err != nil {
		t.Fatal(err)
	}

	err = os.Rename(storage1.backupFilePath(), storage1.tempBackupFilePath())
	if err != nil {
		t.Fatal(err)
	}

	storage2 := NewSpentTimeInMemory(10*time.Hour, dir, 1*time.Second)
	storage2.backupFolder = dir
	storage2.spentTime = make(map[ctxtg.UserID]entities.SpentTime)
	err = storage2.restore()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(storage1.spentTime, storage2.spentTime) {
		t.Error("Invalid backup loaded")
	}
}

func TestNewSpentTime(t *testing.T) {
	s := SpentTimeInMemory{}
	s.spentTime = make(map[ctxtg.UserID]entities.SpentTime)
	st := randSpentTime()
	err := s.NewSpentTime(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	if s.spentTime[st.UserID] != st {
		t.Error("Invalid spent time")
	}
}

func TestNewSpentTimeRewrite(t *testing.T) {
	s := SpentTimeInMemory{}
	s.spentTime = make(map[ctxtg.UserID]entities.SpentTime)
	st := entities.SpentTime{
		UserID:     ctxtg.UserID(rand.Int63()),
		PlanningID: entities.PlanningID(rand.Int63()),
	}
	err := s.NewSpentTime(ctx, randSpentTime())
	if err != nil {
		t.Fatal(err)
	}
	err = s.NewSpentTime(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	if s.spentTime[st.UserID] != st {
		t.Error("Invalid spent time")
	}
}

func TestModifyInvalidUserID(t *testing.T) {
	s := SpentTimeInMemory{}
	s.spentTime = make(map[ctxtg.UserID]entities.SpentTime)
	st := randSpentTime()
	err := s.NewSpentTime(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Modify(ctx, st.UserID/2, func(st1 *entities.SpentTime) (*entities.SpentTime, error) {
		if st1 != nil {
			t.Error("Should be empty")
		}
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestModifyDeleteFromStorage(t *testing.T) {
	s := SpentTimeInMemory{}
	s.spentTime = make(map[ctxtg.UserID]entities.SpentTime)
	st := randSpentTime()
	err := s.NewSpentTime(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Modify(ctx, st.UserID, func(st1 *entities.SpentTime) (*entities.SpentTime, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.spentTime[st.UserID]; ok {
		t.Error("Should be empty")
	}
}

func TestModifyInvalidUserIDWithErr(t *testing.T) {
	s := SpentTimeInMemory{}
	s.spentTime = make(map[ctxtg.UserID]entities.SpentTime)
	st := randSpentTime()
	err := s.NewSpentTime(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	testErr := errors.New("test err")
	err = s.Modify(ctx, st.UserID/2, func(st1 *entities.SpentTime) (*entities.SpentTime, error) {
		if st1 != nil {
			t.Error("Should be empty")
		}
		return nil, testErr
	})
	if err != testErr {
		t.Fatal(err)
	}
}

func TestModifyDeleteFromStorageWithErr(t *testing.T) {
	s := SpentTimeInMemory{}
	s.spentTime = make(map[ctxtg.UserID]entities.SpentTime)
	st := randSpentTime()
	err := s.NewSpentTime(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	testErr := errors.New("test err")
	err = s.Modify(ctx, st.UserID, func(st1 *entities.SpentTime) (*entities.SpentTime, error) {
		return nil, testErr
	})
	if err != testErr {
		t.Fatal(err)
	}
	if _, ok := s.spentTime[st.UserID]; ok {
		t.Error("Should be empty")
	}
}

func TestModifyAddToStorageWithErr(t *testing.T) {
	s := SpentTimeInMemory{}
	s.spentTime = make(map[ctxtg.UserID]entities.SpentTime)
	st := randSpentTime()
	testErr := errors.New("test err")
	err := s.Modify(ctx, st.UserID, func(st1 *entities.SpentTime) (*entities.SpentTime, error) {
		return &st, testErr
	})
	if err != testErr {
		t.Fatal(err)
	}
	if s.spentTime[st.UserID] != st {
		t.Error("Should be empty")
	}
}

func TestModifyStorageWithErr(t *testing.T) {
	s := SpentTimeInMemory{}
	s.spentTime = make(map[ctxtg.UserID]entities.SpentTime)
	st := randSpentTime()
	newLast := st.Last / 2
	err := s.NewSpentTime(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	testErr := errors.New("test err")
	err = s.Modify(ctx, st.UserID, func(st1 *entities.SpentTime) (*entities.SpentTime, error) {
		st1.Last = newLast
		return st1, testErr
	})
	if err != testErr {
		t.Fatal(err)
	}
	if s.spentTime[st.UserID].Last != newLast {
		t.Error("Should be empty")
	}
}

func TestModifyStorage(t *testing.T) {
	s := SpentTimeInMemory{}
	s.spentTime = make(map[ctxtg.UserID]entities.SpentTime)
	st := randSpentTime()
	newLast := st.Last / 2
	err := s.NewSpentTime(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Modify(ctx, st.UserID, func(st1 *entities.SpentTime) (*entities.SpentTime, error) {
		st1.Last = newLast
		return st1, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if s.spentTime[st.UserID].Last != newLast {
		t.Error("Should be empty")
	}
}

func randSpentTime() entities.SpentTime {
	return entities.SpentTime{
		UserID:     ctxtg.UserID(rand.Int63()),
		PlanningID: entities.PlanningID(rand.Int63()),
		Started:    rand.Int63(),
		Last:       rand.Int63(),
	}
}

func tempDir(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir("", "planning-cache-test")
	if err != nil {
		t.Fatal(err)
	}
	return dir, func() { os.RemoveAll(dir) }
}
