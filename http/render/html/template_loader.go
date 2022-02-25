// Copyright 2022 xgfone
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

package html

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// File represents a template file.
type File interface {
	Name() string
	Data() string
}

type file struct {
	name string
	data string
}

// NewFile returns a File.
func NewFile(name, data string) File { return file{name, data} }
func (f file) Name() string          { return f.name }
func (f file) Data() string          { return f.data }

// Loader is used to load the template file from the disk.
type Loader interface {
	// LoadAll reloads the information and content of all the files.
	LoadAll() ([]File, error)
}

// FileFilter is used to filter the template files if it returns true.
type FileFilter func(filepath string) bool

// NewDirLoader is the same as NewDirLoaderWithFilter, not filter any files.
func NewDirLoader(dirs ...string) Loader {
	return NewDirLoaderWithFilter(func(s string) bool { return false }, dirs...)
}

// NewDirLoaderWithFilter returns a new Loader to load the files below the dirs.
//
// Notice: the name of the template file is stripped with the prefix dir.
func NewDirLoaderWithFilter(filter FileFilter, dirs ...string) Loader {
	if filter == nil {
		panic("NewDirLoaderWithFilter: filter must not be nil")
	} else if len(dirs) == 0 {
		panic("NewDirLoaderWithFilter: no dirs")
	}

	_dirs := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		if _len := len(dir); _len > 0 {
			if dir[_len-1] == os.PathSeparator {
				_dirs = append(_dirs, dir)
			} else {
				_dirs = append(_dirs, dir+string(os.PathSeparator))
			}
		}
	}

	return loader{dirs: _dirs, filter: filter}
}

type loader struct {
	dirs   []string
	filter FileFilter
}

func (l loader) LoadAll() (files []File, err error) {
	for _, dir := range l.dirs {
		err = filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			} else if fi.IsDir() {
				return nil
			} else if file, err := l.loadFile(dir, path); err != nil {
				return err
			} else if file != nil {
				files = append(files, file)
			}
			return nil
		})

		if err != nil {
			return
		}
	}
	return
}

func (l loader) loadFile(prefix, filename string) (File, error) {
	if l.filter(filename) {
		return nil, nil
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	name := strings.TrimPrefix(filename, prefix)
	name = strings.Replace(name, "\\", "/", -1)
	return NewFile(name, string(data)), nil
}
