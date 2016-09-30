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
	"net/http"

	"github.com/tadhunt/httputils"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func logit(r *http.Request, herr *httputils.HttpError) {
	ctx := appengine.NewContext(r)
	if herr.Errs != nil {
		log.Errorf(ctx, "Binding Error: %s", herr.Errs.Error())
	}

	if herr.LogMsg != "" {
		log.Errorf(ctx, "%s", herr.LogMsg)
	}
}

/*
 * TODO(tadhunt): replace hacky wrapper around open source http logging library to correctly
 * log appengine errors with something better
 */
func InternalServerError(w http.ResponseWriter, r *http.Request, msg string, logErr error) {
	herr := httputils.NewHttpError()
	herr.Skip++
	herr.InternalServerError(msg, logErr)

	logit(r, herr)
	herr.Write(w)
}

func Error(w http.ResponseWriter, r *http.Request, code int, msg string, logErr error) {
	herr := httputils.NewHttpError()
	herr.Skip++
	herr.Error(code, msg, logErr)

	logit(r, herr)
	herr.Write(w)
}
