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
package main

/*
 * These constants relect the state of the huntd code from
 * https://github.com/ctdk/bsdgames-osx.git 
 * as of commit 993b79dcf751ce491060d0a16c4e68607371d704
 */
const(
	DEBIAN_ULONG_BUG	= false		// this bug doesn't exist
	VERSION_AFTER_JOIN	= true		// but the protocol is different in this respect
)

func ShowBugs() {
	logger.Log(LOG_STARTUP, "OS: Darwin")
	logger.Log(LOG_STARTUP, "DEBIAN_ULONG_BUG   = %v", DEBIAN_ULONG_BUG)
	logger.Log(LOG_STARTUP, "VERSION_AFTER_JOIN = %v", VERSION_AFTER_JOIN)
}
