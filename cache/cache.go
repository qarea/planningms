// Package cache provide in-memory cache for spent time with file backup
package cache

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/powerman/narada-go/narada"

	"github.com/qarea/ctxtg"
	"github.com/qarea/planningms/entities"
)

var log = narada.NewLog("spent time cache: ")

const (
	backupFileName = "backup.json"
	tmpExtension   = ".tmp"
)

//NewSpentTimeInMemory correctly initialize SpentTimeInMemory instance with configuration
func NewSpentTimeInMemory(backupFreq time.Duration, backupFolder string, sharedLock time.Duration) *SpentTimeInMemory {
	s := &SpentTimeInMemory{
		backupFolder:       backupFolder,
		spentTime:          make(map[ctxtg.UserID]entities.SpentTime),
		sharedLockDuration: sharedLock,
	}

	if err := s.withSharedLock(s.restore); err != nil {
		log.ERR("Failed to restore %v", err)
	}
	s.backupEvery(backupFreq)

	return s
}

//SpentTimeInMemory cache with backup in-memory storage to hard drive
type SpentTimeInMemory struct {
	l         sync.Mutex
	spentTime map[ctxtg.UserID]entities.SpentTime

	backupFolder       string
	sharedLockDuration time.Duration
}

//NewSpentTime save st into in-memory storage. SpentTime in storage unique by UserID.
//Only single instance per user could be stored in in-memory storage
func (s *SpentTimeInMemory) NewSpentTime(_ context.Context, st entities.SpentTime) error {
	s.save(st)
	return nil
}

//Modify get SpentTime from in-memory storage and pass it to f.
//If f returns not nil SpentTime it saves it back to in-memory storage else it removes it from storage
func (s *SpentTimeInMemory) Modify(_ context.Context, userID ctxtg.UserID, f entities.ModifySpentTimeFunc) error {
	s.l.Lock()
	defer s.l.Unlock()
	var err error
	var spentTime *entities.SpentTime
	if st, ok := s.spentTime[userID]; ok {
		spentTime, err = f(&st)
	} else {
		spentTime, err = f(nil)
	}
	if spentTime == nil {
		delete(s.spentTime, userID)
	} else {
		s.spentTime[userID] = *spentTime
	}
	return err
}

func (s *SpentTimeInMemory) save(st entities.SpentTime) {
	s.l.Lock()
	s.spentTime[st.UserID] = st
	s.l.Unlock()
}

func (s *SpentTimeInMemory) getAndDelete(userID ctxtg.UserID) *entities.SpentTime {
	s.l.Lock()
	defer s.l.Unlock()
	if st, ok := s.spentTime[userID]; ok {
		delete(s.spentTime, userID)
		return &st
	}
	return nil
}

func (s *SpentTimeInMemory) backupEvery(t time.Duration) {
	go func() {
		for {
			<-time.After(t)
			if err := s.withSharedLock(s.backup); err != nil {
				log.ERR("Failed to backup %+v", err)
			} else {
				log.DEBUG("Backuped in-memory data")
			}
		}
	}()
}

func (s *SpentTimeInMemory) restore() error {
	var b []byte
	b, err := ioutil.ReadFile(s.backupFilePath())
	if err != nil {
		log.ERR("Failed to read %s", s.backupFilePath())
		b, err = ioutil.ReadFile(s.tempBackupFilePath())
		if err != nil {
			log.ERR("Failed to read %s", s.tempBackupFilePath())
			return err
		}

	}
	var sts []entities.SpentTime
	err = json.Unmarshal(b, &sts)
	if err != nil {
		return err
	}
	s.fromSlice(sts)
	return nil
}

func (s *SpentTimeInMemory) backup() error {
	b, err := json.Marshal(s.toSlice())
	if err != nil {
		return errors.Wrap(err, "failed to marshal data")
	}
	tempFilePath := s.tempBackupFilePath()
	err = ioutil.WriteFile(tempFilePath, b, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write to temp file")
	}
	err = os.Rename(tempFilePath, s.backupFilePath())
	if err != nil {
		return errors.Wrap(err, "failed to rename temp file")
	}
	return nil
}

func (s *SpentTimeInMemory) backupFilePath() string {
	return filepath.Join(s.backupFolder, backupFileName)
}

func (s *SpentTimeInMemory) tempBackupFilePath() string {
	return filepath.Join(s.backupFolder, backupFileName+tmpExtension)
}

func (s *SpentTimeInMemory) fromSlice(sts []entities.SpentTime) {
	s.l.Lock()
	for _, st := range sts {
		s.spentTime[st.UserID] = st
	}
	s.l.Unlock()
}

func (s *SpentTimeInMemory) toSlice() []entities.SpentTime {
	var sts []entities.SpentTime
	s.l.Lock()
	for id, st := range s.spentTime {
		st.UserID = id
		sts = append(sts, st)
	}
	s.l.Unlock()
	return sts
}

func (s *SpentTimeInMemory) withSharedLock(f func() error) error {
	l, err := narada.SharedLock(s.sharedLockDuration)
	if err != nil {
		return errors.Wrapf(entities.ErrMaintenance, "can't obtain lock %v", err)
	}
	defer l.UnLock()
	return f()
}
