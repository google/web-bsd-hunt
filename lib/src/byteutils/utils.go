// Copyright 2016 The Web BSD Hunt Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
////////////////////////////////////////////////////////////////////////////////
//
// TODO: High-level file comment.
package byteutils

import(
	"fmt"
)

func StrToBytes(dst []byte, src string, n int, pad byte) error {
	if n > len(dst) {
		return fmt.Errorf("bad args dst len %d n %d", len(dst), n)
	}

	bsrc := []byte(src)
	if len(bsrc) > len(dst) {
		return fmt.Errorf("bad args len(src) %d len(bsrc) %d len(dst) %d", len(src), len(bsrc), len(dst))
	}

	for i := 0; i < n; i++  {
		if i >= len(bsrc) {
			dst[i] = pad
		} else {
			dst[i] = bsrc[i]
		}
	}

	return nil
}
