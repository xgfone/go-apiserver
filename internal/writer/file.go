// Copyright 2023 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package writer

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"sync/atomic"
)

// ParseSize parses the size string. The size maybe have a unit suffix,
// such as "123", "123M, 123G". Valid size units are "b", "B", "k", "K",
// "m", "M", "g", "G", "t", "T", "p", "P", "e", "E". The lower units are 1000x,
// and the upper units are 1024x.
//
// Notice: "" will be considered as 0.
func ParseSize(s string) (size int64, err error) {
	if s == "" {
		return
	}

	var base int64
	switch _len := len(s) - 1; s[_len] {
	case 'b', 'B':
		s = s[:_len]
	case 'k':
		base = 1000
		s = s[:_len]
	case 'K':
		base = 1024
		s = s[:_len]
	case 'm':
		base = 1000000 // 1000**2
		s = s[:_len]
	case 'M':
		base = 1048576 // 1024**2
		s = s[:_len]
	case 'g':
		base = 1000000000 // 1000**3
		s = s[:_len]
	case 'G':
		base = 1073741824 // 1024**3
		s = s[:_len]
	case 't':
		base = 1000000000000 // 1000**4
		s = s[:_len]
	case 'T':
		base = 1099511627776 // 1024**4
		s = s[:_len]
	case 'p':
		base = 1000000000000000 // 1000**5
		s = s[:_len]
	case 'P':
		base = 1125899906842624 // 1024**5
		s = s[:_len]
	case 'e':
		base = 1000000000000000000 // 1000**6
		s = s[:_len]
	case 'E':
		base = 1152921504606846976 // 1024**6
		s = s[:_len]
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
	default:
		return 0, fmt.Errorf("unknown size string '%s'", s)
	}

	if size, err = strconv.ParseInt(s, 10, 64); err == nil && base > 1 {
		size *= base
	}

	return
}

// NewSizedRotatingFile returns a new SizedRotatingFile, which is not thread-safe.
//
// Default:
//
//	fileperm: 0644
//	filesize: 100 * 1024 * 1024
//	filenum:  0
func NewSizedRotatingFile(filename string, filesize, filenum int,
	fileperm ...os.FileMode) *SizedRotatingFile {
	var filemode os.FileMode = 0644
	if len(fileperm) > 0 && fileperm[0] > 0 {
		filemode = fileperm[0]
	}

	if filenum <= 0 {
		filesize = int(math.MaxInt32)
	} else if filesize <= 0 {
		filesize = 100 * 1024 * 1024
	}

	return &SizedRotatingFile{
		filename:    filename,
		filemode:    filemode,
		maxSize:     filesize,
		backupCount: filenum,
	}
}

// SizedRotatingFile is a file rotating logging writer based on the size.
type SizedRotatingFile struct {
	file        *os.File
	filemode    os.FileMode
	filename    string
	maxSize     int
	backupCount int
	nbytes      int
	closed      int32
}

// Close implements io.Closer.
func (f *SizedRotatingFile) Close() (err error) {
	if atomic.CompareAndSwapInt32(&f.closed, 0, 1) {
		err = f.close()
	}
	return
}

// Sync is equal to Flush to flush the data to the underlying disk.
func (f *SizedRotatingFile) Sync() (err error) {
	return f.Flush()
}

// Flush flushes the data to the underlying disk.
func (f *SizedRotatingFile) Flush() (err error) {
	if f.file != nil {
		err = f.file.Sync()
	}
	return
}

// Write implements io.Writer.
func (f *SizedRotatingFile) Write(data []byte) (n int, err error) {
	if atomic.LoadInt32(&f.closed) == 1 {
		return 0, errors.New("the file has been closed")
	}

	if f.file == nil {
		if err = f.open(); err != nil {
			return
		}
	}

	if f.nbytes+len(data) > f.maxSize {
		if err = f.doRollover(); err != nil {
			return
		}
	}

	if n, err = f.file.Write(data); err != nil {
		return
	}

	f.nbytes += n
	return
}

func (f *SizedRotatingFile) open() (err error) {
	file, err := os.OpenFile(f.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, f.filemode)
	if err != nil {
		return
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return
	}

	f.nbytes = int(info.Size())
	f.file = file
	return
}

func (f *SizedRotatingFile) close() (err error) {
	if f.file != nil {
		err = f.file.Close()
		f.file = nil
	}
	return
}

func (f *SizedRotatingFile) doRollover() (err error) {
	if f.backupCount > 0 {
		if err = f.close(); err != nil {
			return fmt.Errorf("failed to close the rotating file '%s': %s", f.filename, err)
		}

		if !fileIsExist(f.filename) {
			return nil
		} else if n, err := fileSize(f.filename); err != nil {
			return fmt.Errorf("failed to get the size of the rotating file '%s': %s",
				f.filename, err)
		} else if n == 0 {
			return nil
		}

		for _, i := range ranges(f.backupCount-1, 0, -1) {
			sfn := fmt.Sprintf("%s.%d", f.filename, i)
			dfn := fmt.Sprintf("%s.%d", f.filename, i+1)
			if fileIsExist(sfn) {
				if fileIsExist(dfn) {
					os.Remove(dfn)
				}
				if err = os.Rename(sfn, dfn); err != nil {
					return fmt.Errorf("failed to rename the rotating file '%s' to '%s': %s",
						sfn, dfn, err)
				}
			}
		}

		dfn := f.filename + ".1"
		if fileIsExist(dfn) {
			if err = os.Remove(dfn); err != nil {
				return fmt.Errorf("failed to remove the rotating file '%s': %s", dfn, err)
			}
		}
		if fileIsExist(f.filename) {
			if err = os.Rename(f.filename, dfn); err != nil {
				return fmt.Errorf("failed to rename the rotating file '%s' to '%s': %s",
					f.filename, dfn, err)
			}
		}

		err = f.open()
	}

	return
}

func fileIsExist(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// fileSize returns the size of the file as how many bytes.
func fileSize(fp string) (int64, error) {
	f, e := os.Stat(fp)
	if e != nil {
		return 0, e
	}
	return f.Size(), nil
}

func ranges(start, stop, step int) (r []int) {
	if step > 0 {
		for start < stop {
			r = append(r, start)
			start += step
		}
		return
	} else if step < 0 {
		for start > stop {
			r = append(r, start)
			start += step
		}
		return
	}

	panic(fmt.Errorf("step must not be 0"))
}
