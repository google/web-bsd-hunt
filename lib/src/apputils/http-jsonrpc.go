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
package apputils

import(
	"bytes"
	"net/http"
	"net/url"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	gjson "github.com/gorilla/rpc/json"
)

const(
	DEBUG	= false
)

/*
 * This method and the supporting RProxy() function are purpose built for App Engine,
 * And more specifically for Appengine apps that want to either
 * talk between modules in the same app, or talk between apps.
 */
func HttpJsonRpc(r *http.Request, u *url.URL, method string, rpoptions uint64, request interface{}, reply interface{}) error {
	ctx := appengine.NewContext(r)

	cfg := &RProxyConfig{
		Add:		[]*Header{
					&Header{Key: "REDACTED", Value: "REDACTED"},
					&Header{Key: "Content-Type", Value: "application/json;charset=utf-8"},
					&Header{Key: "Accept", Value: "application/json;charset=utf-8"},
				},
		Options:	rpoptions,
	}

	buf, err := gjson.EncodeClientRequest(method, request)
	if err != nil {
		log.Errorf(ctx, "encode: %v", err)
		return err
	}

	var rbuf []byte
	rbuf, err = RProxy(r, cfg, "POST", u, buf)
	if err != nil {
		log.Errorf(ctx, "RProxy: %v", err)
		return err
	}

	err = gjson.DecodeClientResponse(bytes.NewBuffer(rbuf), reply)
	if err != nil {
		log.Errorf(ctx, "Decode Client Response: %v", err)
		return err
	}

	return nil
}
