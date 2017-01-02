// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package sunrpc

import (
	"errors"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

/*
From RFC 5531:
    The RPC call message has three unsigned-integer fields -- remote
    program number, remote program version number, and remote procedure
    number -- that uniquely identify the procedure to be called.
*/
type procedureKey struct {
	programNumber   uint32
	programVersion  uint32
	procedureNumber uint32
}

var procedureRegistry = struct {
	sync.RWMutex
	pMap map[procedureKey]string
}{
	pMap: make(map[procedureKey]string),
}

func isExported(name string) bool {
	firstRune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(firstRune)
}

func isValidProcedureName(procedureName string) bool {
	// procedureName must be of the format 'T.MethodName' to satisfy
	// criteria set by 'net/rpc' package for remote functions.

	procedureTypeName := strings.Split(procedureName, ".")
	if len(procedureTypeName) != 2 {
		return false
	}

	for _, name := range procedureTypeName {
		if !isExported(name) {
			return false
		}
	}

	return true
}

// RegisterProcedure will register the procedure name which will be uniquely
// indentified by (programNumber, programVersion, procedureNumber) pair.
func RegisterProcedure(programNumber uint32, programVersion uint32, procedureNumber uint32, procedureName string) error {

	if !isValidProcedureName(procedureName) {
		return errors.New("Invalid procedure name")
	}

	procedureRegistry.Lock()
	defer procedureRegistry.Unlock()

	key := procedureKey{programNumber, programVersion, procedureNumber}
	procedureRegistry.pMap[key] = procedureName
	return nil
}

// GetProcedureName will return a string containing procedure name and a bool
// value which is set to true only if the procedure is found in registry.
func GetProcedureName(programNumber uint32, programVersion uint32, procedureNumber uint32) (string, bool) {
	procedureRegistry.RLock()
	defer procedureRegistry.RUnlock()

	key := procedureKey{programNumber, programVersion, procedureNumber}
	procedureName, ok := procedureRegistry.pMap[key]
	return procedureName, ok
}
