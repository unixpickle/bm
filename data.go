package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
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
func (d *DataFile) Write(record *CommandRecord) error {
	origOffset, err := d.file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	if _, err := d.file.Write(data); err != nil {
		// Attempt to revert the partial write, although
		// there's no guarantee this will work, and the
		// file may become corrupted.
		d.file.Truncate(origOffset)

		return err
	}
	return nil
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
