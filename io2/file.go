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

package io2

import (
	"errors"
	"io"

	"github.com/xgfone/go-apiserver/internal/writer"
)

// NewFileWriter returns a new file writer that rotates the files
// based on the file size, which is used as the log writer.
//
// filesize is parsed as the file size, which maybe have a unit suffix,
// such as "123", "123M, 123G". Valid size units contain "b", "B", "k", "K",
// "m", "M", "g", "G", "t", "T", "p", "P", "e", "E". The lower units are 1000x,
// and the upper units are 1024x.
func NewFileWriter(filepath, filesize string, filenum int) (io.WriteCloser, error) {
	if filepath == "" {
		return nil, errors.New("the log filepath must not be empty")
	}

	size, err := writer.ParseSize(filesize)
	if err != nil {
		return nil, err
	}

	return writer.NewSizedRotatingFile(filepath, int(size), filenum), nil
}
