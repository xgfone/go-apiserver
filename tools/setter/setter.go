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

package setter

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/xgfone/go-apiserver/helper"
)

// Setter is an interface to set itself to a value.
type Setter interface {
	Set(interface{}) error
}

// Set does the best to set the value of dst to src.
//
// Notice: dst must be a pointer to a variable of the built-in types,
// time.Time, time.Duration, Setter or sql.Scanner.
func Set(dst, src interface{}) (err error) {
	src = helper.Indirect(src)
	switch d := dst.(type) {
	case nil:
		return

	case *bool:
		var v bool
		if v, err = convertToBool(src); err == nil {
			*d = v
		}

	case *string:
		var v string
		if v, err = convertToString(src); err == nil {
			*d = v
		}

	case *float32:
		var v float64
		if v, err = convertToFloat64(src); err == nil {
			*d = float32(v)
		}

	case *float64:
		var v float64
		if v, err = convertToFloat64(src); err == nil {
			*d = v
		}

	case *int:
		var v int64
		if v, err = convertToInt64(src); err == nil {
			*d = int(v)
		}

	case *int8:
		var v int64
		if v, err = convertToInt64(src); err == nil {
			*d = int8(v)
		}

	case *int16:
		var v int64
		if v, err = convertToInt64(src); err == nil {
			*d = int16(v)
		}

	case *int32:
		var v int64
		if v, err = convertToInt64(src); err == nil {
			*d = int32(v)
		}

	case *int64:
		var v int64
		if v, err = convertToInt64(src); err == nil {
			*d = v
		}

	case *uint:
		var v uint64
		if v, err = convertToUint64(src); err == nil {
			*d = uint(v)
		}

	case *uint8:
		var v uint64
		if v, err = convertToUint64(src); err == nil {
			*d = uint8(v)
		}

	case *uint16:
		var v uint64
		if v, err = convertToUint64(src); err == nil {
			*d = uint16(v)
		}

	case *uint32:
		var v uint64
		if v, err = convertToUint64(src); err == nil {
			*d = uint32(v)
		}

	case *uint64:
		var v uint64
		if v, err = convertToUint64(src); err == nil {
			*d = v
		}

	case *time.Duration:
		var v time.Duration
		if v, err = convertToDuration(src); err == nil {
			*d = v
		}

	case *time.Time:
		var v time.Time
		if v, err = convertToTime(src); err == nil {
			*d = v
		}

	case Setter:
		err = d.Set(src)

	case sql.Scanner:
		err = d.Scan(src)

	default:
		err = fmt.Errorf("unsupported the dst type %T", dst)
	}

	return
}

func convertToBool(i interface{}) (dst bool, err error) {
	switch src := i.(type) {
	case nil:
	case bool:
		dst = src
	case string:
		dst, err = strconv.ParseBool(src)
	case []byte:
		switch len(src) {
		case 0:
		case 1:
			switch src[0] {
			case '\x00':
			case '\x01':
				dst = true
			default:
				dst, err = strconv.ParseBool(string(src))
			}
		default:
			dst, err = strconv.ParseBool(string(src))
		}
	case float32:
		dst = src != 0
	case float64:
		dst = src != 0
	case int:
		dst = src != 0
	case int8:
		dst = src != 0
	case int16:
		dst = src != 0
	case int32:
		dst = src != 0
	case int64:
		dst = src != 0
	case uint:
		dst = src != 0
	case uint8:
		dst = src != 0
	case uint16:
		dst = src != 0
	case uint32:
		dst = src != 0
	case uint64:
		dst = src != 0
	case fmt.Stringer:
		dst, err = strconv.ParseBool(src.String())
	default:
		err = fmt.Errorf("unsupported the dst type %T", dst)
	}
	return
}

func convertToString(i interface{}) (dst string, err error) {
	switch src := i.(type) {
	case nil:
	case bool:
		dst = strconv.FormatBool(src)
	case string:
		dst = src
	case []byte:
		dst = string(src)
	case float32:
		dst = strconv.FormatFloat(float64(src), 'f', -1, 32)
	case float64:
		dst = strconv.FormatFloat(src, 'f', -1, 64)
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		dst = fmt.Sprint(src)
	case fmt.Stringer:
		dst = src.String()
	default:
		err = fmt.Errorf("unsupported the dst type %T", dst)
	}
	return
}

func convertToInt64(i interface{}) (dst int64, err error) {
	switch src := i.(type) {
	case nil:
	case bool:
		if src {
			dst = 1
		}
	case string:
		dst, err = strconv.ParseInt(src, 0, 64)
	case []byte:
		dst, err = strconv.ParseInt(string(src), 0, 64)
	case float32:
		dst = int64(src)
	case float64:
		dst = int64(src)
	case int:
		dst = int64(src)
	case int8:
		dst = int64(src)
	case int16:
		dst = int64(src)
	case int32:
		dst = int64(src)
	case int64:
		dst = src
	case uint:
		dst = int64(src)
	case uint8:
		dst = int64(src)
	case uint16:
		dst = int64(src)
	case uint32:
		dst = int64(src)
	case uint64:
		dst = int64(src)
	case fmt.Stringer:
		dst, err = strconv.ParseInt(src.String(), 0, 64)
	default:
		err = fmt.Errorf("unsupported the dst type %T", dst)
	}
	return
}

func convertToUint64(i interface{}) (dst uint64, err error) {
	switch src := i.(type) {
	case nil:
	case bool:
		if src {
			dst = 1
		}
	case string:
		dst, err = strconv.ParseUint(src, 0, 64)
	case []byte:
		dst, err = strconv.ParseUint(string(src), 0, 64)
	case float32:
		dst = uint64(src)
	case float64:
		dst = uint64(src)
	case int:
		dst = uint64(src)
	case int8:
		dst = uint64(src)
	case int16:
		dst = uint64(src)
	case int32:
		dst = uint64(src)
	case int64:
		dst = uint64(src)
	case uint:
		dst = uint64(src)
	case uint8:
		dst = uint64(src)
	case uint16:
		dst = uint64(src)
	case uint32:
		dst = uint64(src)
	case uint64:
		dst = src
	case fmt.Stringer:
		dst, err = strconv.ParseUint(src.String(), 0, 64)
	default:
		err = fmt.Errorf("unsupported the dst type %T", dst)
	}
	return
}

func convertToFloat64(i interface{}) (dst float64, err error) {
	switch src := i.(type) {
	case nil:
	case bool:
		if src {
			dst = 1
		}
	case string:
		dst, err = strconv.ParseFloat(src, 64)
	case []byte:
		dst, err = strconv.ParseFloat(string(src), 64)
	case float32:
		dst = float64(src)
	case float64:
		dst = src
	case int:
		dst = float64(src)
	case int8:
		dst = float64(src)
	case int16:
		dst = float64(src)
	case int32:
		dst = float64(src)
	case int64:
		dst = float64(src)
	case uint:
		dst = float64(src)
	case uint8:
		dst = float64(src)
	case uint16:
		dst = float64(src)
	case uint32:
		dst = float64(src)
	case uint64:
		dst = float64(src)
	case fmt.Stringer:
		dst, err = strconv.ParseFloat(src.String(), 64)
	default:
		err = fmt.Errorf("unsupported the dst type %T", dst)
	}
	return
}

func convertToDuration(i interface{}) (dst time.Duration, err error) {
	switch src := i.(type) {
	case nil:
	case string:
		if helper.IsIntegerString(src) {
			var i int64
			i, err = strconv.ParseInt(src, 10, 64)
			dst = time.Duration(i) * time.Millisecond
		} else {
			dst, err = time.ParseDuration(src)
		}
	case []byte:
		dst, err = time.ParseDuration(string(src))
	case float32:
		dst = time.Duration(src)
	case float64:
		dst = time.Duration(src)
	case int:
		dst = time.Duration(src)
	case int8:
		dst = time.Duration(src)
	case int16:
		dst = time.Duration(src)
	case int32:
		dst = time.Duration(src)
	case int64:
		dst = time.Duration(src)
	case uint:
		dst = time.Duration(src)
	case uint8:
		dst = time.Duration(src)
	case uint16:
		dst = time.Duration(src)
	case uint32:
		dst = time.Duration(src)
	case uint64:
		dst = time.Duration(src)
	case fmt.Stringer:
		dst, err = time.ParseDuration(src.String())
	default:
		err = fmt.Errorf("unsupported the dst type %T", dst)
	}
	return
}

func convertToTime(i interface{}) (dst time.Time, err error) {
	switch src := i.(type) {
	case nil:
	case string:
		dst, err = convertStringToTime(src)
	case []byte:
		dst, err = convertStringToTime(string(src))
	case float32:
		dst = time.Unix(int64(src), 0)
	case float64:
		dst = time.Unix(int64(src), 0)
	case int:
		dst = time.Unix(int64(src), 0)
	case int64:
		dst = time.Unix(int64(src), 0)
	case uint:
		dst = time.Unix(int64(src), 0)
	case uint64:
		dst = time.Unix(int64(src), 0)
	case fmt.Stringer:
		dst, err = convertStringToTime(src.String())
	default:
		err = fmt.Errorf("unsupported the dst type %T", dst)
	}
	return
}

func convertStringToTime(s string) (time.Time, error) {
	if s == "" || s == "0000-00-00 00:00:00" {
		return time.Time{}, nil
	} else if helper.IsIntegerString(s) {
		i, err := strconv.ParseInt(s, 10, 64)
		return time.Unix(i, 0), err
	}
	return time.Parse(time.RFC3339, s)
}
