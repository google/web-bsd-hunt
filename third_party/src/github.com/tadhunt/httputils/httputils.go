/*
Copyright © 2016, Tad Hunt <tadhunt@gmail.com>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions
are met:

1. Redistributions of source code must retain the above copyright
   notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright
   notice, this list of conditions and the following disclaimer in the
   documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS
FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT,
INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN
ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/

package httputils

import (
	"fmt"
	"log"
	"net/http"

	"github.com/tadhunt/go-dbgutils"

	"github.com/mholt/binding"
)

type HttpError struct {
	Skip int // callstack frames to skip for debug tracking

	Errs binding.Errors

	Msg  string
	Code int

	LogMsg string
}

// HttpError represents both the error to send to the user and the error to log.
func NewHttpError() *HttpError {
	herr := &HttpError{
		Skip:   1,
		Errs:   nil,
		Msg:    "OK",
		Code:   http.StatusOK,
		LogMsg: "OK",
	}

	return herr
}

func (herr *HttpError) mkerror(skip int, code int, msg string, logErr error) {
	herr.Code = code
	if msg == "" {
		herr.Msg = fmt.Sprintf("%d %v", code, http.StatusText(code))
	} else {
		herr.Msg = fmt.Sprintf("%d %v", code, msg)
	}

	herr.LogMsg = fmt.Sprintf("{%s} <%d> %s [%v]", debug.FuncName(skip, true), herr.Code, herr.Msg, logErr)
}

func (herr *HttpError) Error(code int, msg string, logErr error) {
	herr.mkerror(herr.Skip+1, code, msg, logErr)
}

func (herr *HttpError) Errors(errs binding.Errors) {
	herr.Errs = errs
}

func (herr *HttpError) InternalServerError(msg string, logErr error) {
	herr.mkerror(herr.Skip+1, http.StatusInternalServerError, msg, logErr)
}

func (herr *HttpError) OK() bool {
	if herr.Errs != nil {
		return false
	}

	if herr.Code != http.StatusOK {
		return false
	}

	return true
}

func (herr *HttpError) Write(w http.ResponseWriter) {
	if herr.Errs != nil {
		log.Printf("Binding Error: %s", herr.Errs.Error())
		herr.Errs.Handle(w)
		return
	}

	log.Printf("%s\n", herr.LogMsg)
	http.Error(w, herr.Msg, herr.Code)
}

// The following functions are shorthand for creating and emitting the error all at once
func InternalServerError(w http.ResponseWriter, msg string, logErr error) {
	herr := NewHttpError()
	herr.Skip++
	herr.InternalServerError(msg, logErr)
	herr.Write(w)
}

func Error(w http.ResponseWriter, code int, msg string, logErr error) {
	herr := NewHttpError()
	herr.Skip++
	herr.Error(code, msg, logErr)
	herr.Write(w)
}
