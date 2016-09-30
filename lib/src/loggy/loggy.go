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
// lightweight logging library with the ability to independently configure
// the output function and to dynamically enable and disable and filter
// up to 64 log levels.
package loggy

import(
	"log"
	"fmt"
	"os"
	"strings"
)

type Logger struct {
	Levels	map[string]uint64
	Enabled	uint64
	Output	func(format string, v ...interface{})
}

func NewLogger(levels []string) (*Logger, error) {
	logger := &Logger{
		Levels: make(map[string]uint64),
		Output:	log.Printf,
	}

	for i, name := range levels {
		logger.Levels[name] = 1 << uint64(i)
	}

	return logger, nil
}

// like NewLogger but panic instead of returning error
func MustNewLogger(levels []string, enabled []string) *Logger {
	logger, err := NewLogger(levels)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	for _, l := range enabled {
		err := logger.Enable(l)
		if err != nil {
			panic(fmt.Sprintf("%v", err))
		}
	}

	return logger
}

// splits the enabled string at "," characters, and tries to enable each level so found,
// calls panic instead of returning error if any enabled levels have not been configured.
func MustNewLoggerFromString(levels []string, enabled string) *Logger {
	var e []string

	if enabled != "" {
		e = strings.Split(enabled, ",")
	}

	return MustNewLogger(levels, e)
}

// Enable the log level with the given name
func (logger *Logger) Enable(name string) error {
	l, found := logger.Levels[name]
	if !found {
		return fmt.Errorf("unknown log level '%s'", name)
	}

	logger.Enabled |= l

	return nil
}

// Disable the log level with the given name
func (logger *Logger) Disable(name string) error {
	l, found := logger.Levels[name]
	if !found {
		return fmt.Errorf("unknown log level '%s'", name)
	}

	logger.Enabled = logger.Enabled &^ l

	return nil
}

// Convert a name into a log level
// Can be used to cache level bits for later use
func (logger *Logger) Level(name string) (uint64, error) {
	l, found := logger.Levels[name]
	if !found {
		return 0, fmt.Errorf("unknown log level '%s'", name)
	}

	return l, nil
}

func (logger *Logger) MustLevel(name string) uint64 {
	l, err := logger.Level(name)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	return l
}

// API match with log.Printf() -- avoids level check
func (logger *Logger) Printf(format string, v ...interface{}) {
	logger.Output(format, v...)
}

// API match with log.Fatalf() -- avoids level check
func (logger *Logger) Fatalf(format string, v ...interface{}) {
	logger.Output(format, v...)
	os.Exit(1)
}

// Loggy specific API, only logs if this level is enabled
func (logger *Logger) Log(level uint64, format string, v ...interface{}) {
	if (logger.Enabled & level) != level {		// testing like this allows for multi-bit levels with overlapping bits
		return
	}

	logger.Output(format, v...)
}
