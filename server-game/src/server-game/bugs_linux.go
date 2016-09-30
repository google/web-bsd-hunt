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

const(
	DEBIAN_ULONG_BUG	= true
	VERSION_AFTER_JOIN	= false
)

func ShowBugs() {
	logger.Log(LOG_STARTUP, "OS: Linux")
	logger.Log(LOG_STARTUP, "DEBIAN_ULONG_BUG   = %v", DEBIAN_ULONG_BUG)
	logger.Log(LOG_STARTUP, "VERSION_AFTER_JOIN = %v", VERSION_AFTER_JOIN)
}
