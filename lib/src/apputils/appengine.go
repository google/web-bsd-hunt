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
	"errors"
	"fmt"
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"net/http"
	"net/url"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"google.golang.org/appengine/log"

	"golang.org/x/net/context"
)

func DefaultVersionHostname(r *http.Request) string {
	ctx := appengine.NewContext(r)
	hostname := appengine.DefaultVersionHostname(ctx)
	return hostname
}


func Log(r *http.Request, msg string) {
	ctx := appengine.NewContext(r)
	log.Errorf(ctx, "%s", msg)
}

func LogHeaders(ctx context.Context, msg string, headers http.Header) {
	for header, value := range headers {
		log.Errorf(ctx, "%s Header: %v value %v\n", msg, header, value)
	}
}

func LogRequestHeaders(r *http.Request, msg string) {
	ctx := appengine.NewContext(r)
	LogHeaders(ctx, msg, r.Header)
}

func CopyHeader(dst http.Header, src http.Header, key string) {
	vv, found := src[key]
	if !found {
		return
	}

	for _, v := range vv {
		dst.Add(key, v)
	}
}

var ErrRedirect = errors.New("redirect foiled")

const (
	RPROXY_LOG_REQUEST_HEADERS	= 1 << 0
	RPROXY_LOG_RESPONSE_HEADERS	= 1 << 1
	RPROXY_LOG_ERROR		= 1 << 2
	RPROXY_LOG_SUCCESS		= 1 << 3
	RPROXY_LOG_PROXY_REQUEST	= 1 << 5
	RPROXY_LOG_PROXY_HEADERS	= 1 << 6
	RPROXY_DISABLE_REDIRECT		= 1 << 7
)

var optionMap = map[string] uint64 {
	"RPROXY_LOG_REQUEST_HEADERS":	RPROXY_LOG_REQUEST_HEADERS,
	"RPROXY_LOG_RESPONSE_HEADERS":	RPROXY_LOG_RESPONSE_HEADERS,
	"RPROXY_LOG_ERROR":		RPROXY_LOG_ERROR,
	"RPROXY_LOG_SUCCESS":		RPROXY_LOG_SUCCESS,
	"RPROXY_LOG_PROXY_REQUEST":	RPROXY_LOG_PROXY_REQUEST,
	"RPROXY_LOG_PROXY_HEADERS":	RPROXY_LOG_PROXY_HEADERS,
	"RPROXY_DISABLE_REDIRECT":	RPROXY_DISABLE_REDIRECT,
}

type Header struct {
	Key	string
	Value	string
}

type RProxyConfig struct {
	Copy	[]string
	Add	[]*Header
	Options	uint64
}

func RProxyOptions(str string) uint64 {
	strs := strings.Split(str, ",")

	var options uint64
	for _, s := range strs {
		if s == "" {
			continue
		}
		v, found := optionMap[s]
		if !found {
			fmt.Printf("RProxyOptions: Ignoring unknown option %s", s)
		}
		options |= v
	}

	return options
}

func RProxy(r *http.Request, cfg *RProxyConfig, method string, u *url.URL, data []byte) ([]byte, error) {
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)
	var buf io.Reader
	var httpRsp *http.Response

	if (cfg.Options & RPROXY_LOG_REQUEST_HEADERS) != 0 {
		LogHeaders(ctx, "Get request", r.Header)
	}

	if data != nil {
		buf = bytes.NewReader(data)
	}

	httpReq, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		goto fail
	}

	for _, h := range cfg.Copy {
		CopyHeader(httpReq.Header, r.Header, h)
	}
	for _, h := range cfg.Add {
		httpReq.Header.Set(h.Key, h.Value)
	}

	if (cfg.Options & RPROXY_LOG_PROXY_REQUEST) != 0 {
		log.Errorf(ctx, "Proxy Request method %s url %s data %v\n", method, u, data)
	}

	if (cfg.Options & RPROXY_LOG_PROXY_HEADERS) != 0 {
		LogHeaders(ctx, "game server request", httpReq.Header)
	}

	if (cfg.Options & RPROXY_DISABLE_REDIRECT) != 0 {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return ErrRedirect
		}
	}

	httpRsp, err = client.Do(httpReq)
	if err != nil {
		ue, ok := err.(*url.Error)
		if !ok || ue.Err != ErrRedirect {
			goto fail
		}
	}
	defer httpRsp.Body.Close()

	data, err = ioutil.ReadAll(httpRsp.Body)
	if err != nil {
		goto fail
	}

	if (cfg.Options & RPROXY_LOG_RESPONSE_HEADERS) != 0 {
		LogHeaders(ctx, "Response Headers", httpRsp.Header)
	}

	switch httpRsp.StatusCode {
	case http.StatusOK:
		break
	case  http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect:
		if (cfg.Options & RPROXY_DISABLE_REDIRECT) != 0 {
			break
		}
		fallthrough
	default:
		err = fmt.Errorf("HTTP Error %d: %s", httpRsp.StatusCode, string(data))
		goto fail
	}

	if (cfg.Options & RPROXY_LOG_SUCCESS) != 0 {
		log.Errorf(ctx, "%s", string(data))
	}
	return data, nil

fail:
	if (cfg.Options & RPROXY_LOG_ERROR) != 0 {
		log.Errorf(ctx, "rproxy %s err '%v'", u, err)
	}
	return nil, err
}
