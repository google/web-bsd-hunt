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
package gamerpc

import(
	"fmt"
	"net/http"
	"strings"
)

const(
	GR_NETRPC = iota+1
	GR_JSONRPC
)

func StringToRpcType(str string) (int, error) {
	switch(str) {
	case "netrpc":
		return GR_NETRPC, nil
	case "jsonrpc":
		return GR_JSONRPC, nil
	}

	return 0, fmt.Errorf("unsupported rpc type '%s'", str)
}

//
// TODO(tadhunt): find a better home for this
//
func ContentTypeIsJSON(header http.Header) error {
	h := header["Content-Type"]
	if h == nil {
		return fmt.Errorf("missing Content-Type")
	}

	for _, ctype := range h {
		if strings.Contains(ctype, "application/json") {
			return nil
		}
	}

	return fmt.Errorf("expected Content-Type application/json got '%v'", h)
}

