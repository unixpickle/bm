package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/flock"
)

const (
	DataFileName     = ".bm_bookmark_data"
	DataFileLockName = ".bm_bookmark_data.lock"
)

type CommandRecord struct {
	ID      string
	Command string
	Date    time.Time
}

type DataFile struct {
	lock *flock.Flock
	file *os.File
}

func NewDataFile() (*DataFile, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	lock := flock.New(filepath.Join(homedir, DataFileLockName))
	if err := lock.Lock(); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(
		filepath.Join(homedir, DataFileName),
		os.O_RDWR|os.O_CREATE|os.O_SYNC,
		0600,
	)
	if err != nil {
		lock.Unlock()
		return nil, err
	}
	return &DataFile{
		lock: lock,
		file: file,
	}, nil
}

// Reset seeks to the beginning of the data file.
func (d *DataFile) Reset() error {
	_, err := d.file.Seek(0, io.SeekStart)
	return err
}

// Read the next record.
//
// If the record is nil, an error is always returned.
// An io.EOF means we have finished reading records.
func (d *DataFile) Read() (*CommandRecord, error) {
	var line []byte
	for {
		var ch [1]byte
		n, err := d.file.Read(ch[:])
		if err == io.EOF {
			if len(strings.TrimSpace(string(line))) == 0 {
				return nil, io.EOF
			} else {
				return nil, io.ErrUnexpectedEOF
			}
		} else if n != 1 {
			return nil, err
		}
		if ch[0] == '\n' {
			break
		} else {
			line = append(line, ch[0])
		}
	}
	var obj CommandRecord
	if err := json.Unmarshal(line, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

// Write a record to the end of the file.
//
// This will change the seek location in the file.
func (d *DataFile) Write(record *CommandRecord) error {
	origOffset, err := d.file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	if _, err := d.file.Write(append(data, '\n')); err != nil {
		// Attempt to revert the partial write, although
		// there's no guarantee this will work, and the
		// file may become corrupted.
		d.file.Truncate(origOffset)

		return err
	}
	return nil
}

// Delete deletes the record with the given ID.
//
// This will change the seek location in the file.
//
// If the delete operation fails, the file may be
// inadvertently closed, in which case future operations
// on d will fail.
func (d *DataFile) Delete(id string) error {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	tmpPath := filepath.Join(tmpDir, "history.tmp")
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	d.Reset()
	for {
		record, err := d.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if record.ID == id {
			continue
		}
		encoded, err := json.Marshal(record)
		if err != nil {
			return err
		}
		if _, err := tmpFile.Write(append(encoded, '\n')); err != nil {
			return err
		}
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	oldFileName := d.file.Name()
	d.file.Close()
	if err := os.Rename(tmpPath, oldFileName); err != nil {
		return err
	}
	file, err := os.OpenFile(oldFileName, os.O_RDWR|os.O_SYNC, 0)
	if err != nil {
		return err
	}
	d.file = file
	return nil
}

// GenerateUniqueID generates a random ID that is not
// already used by a record.
//
// This will change the seek location in the file.
func (d *DataFile) GenerateUniqueID() (string, error) {
	ids := map[string]bool{}
	d.Reset()
	for {
		entry, err := d.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}
		ids[entry.ID] = true
	}
	for i := 0; i < 1e6; i++ {
		entry := strconv.FormatInt(int64(i), 36)
		if !ids[entry] {
			return entry, nil
		}
	}
	return "", errors.New("exhausted space of possible UIDs")
}

// CanUseID returns true if the ID is not already in use
// by a record.
//
// This will change the seek location in the file.
func (d *DataFile) CanUseID(id string) (bool, error) {
	d.Reset()
	for {
		entry, err := d.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return false, err
		}
		if entry.ID == id {
			return false, nil
		}
	}
	return true, nil
}

func (d *DataFile) Close() error {
	err := d.file.Close()
	err1 := d.lock.Unlock()
	if err != nil {
		return err
	} else {
		return err1
	}
}
