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

package helper

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func ExampleContextCancelWithReason() {
	internal := time.Millisecond * 100

	ctx, cancel := ContextCancelWithReason(context.TODO())
	go func() {
		time.Sleep(internal)
		cancel(errors.New("cancel reason1"))
		cancel(errors.New("cancel reason2"))
	}()

	start := time.Now()
	<-ctx.Done()

	if time.Since(start) < internal {
		fmt.Println("the sleep internal failed")
	}
	fmt.Println(ctx.Err())

	// Output:
	// cancel reason1
}
