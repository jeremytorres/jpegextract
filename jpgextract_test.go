/*
 Copyright (c) 2013 Jeremy Torres, https://github.com/jeremytorres/jpegextract

 Permission is hereby granted, free of charge, to any person obtaining
 a copy of this software and associated documentation files (the
 "Software"), to deal in the Software without restriction, including
 without limitation the rights to use, copy, modify, merge, publish,
 distribute, sublicense, and/or sell copies of the Software, and to
 permit persons to whom the Software is furnished to do so, subject to
 the following conditions:

 The above copyright notice and this permission notice shall be
 included in all copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
 NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
 LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
 OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package main

import (
	"os"
	"strings"
	"testing"
)

func TestDirectoryValid(t *testing.T) {
	// 1. Test directory existence

	const pathSep = string(os.PathSeparator)

	tmpDir := os.TempDir()

	if !strings.HasSuffix(tmpDir, pathSep) {
		tmpDir += pathSep
	}

	t.Logf("TmpDir: %s\n", tmpDir)

	// create tmp dir and delete it
	dirName := tmpDir + "ParserTest"
	err := os.MkdirAll(dirName, 0755)
	if err != nil {
		t.Fatalf("Unable to create test tmp dir: '%s'.  Error %v\n", dirName, err)
	}
	t.Logf("Created tmp directory '%s'", dirName)

	// delete directory
	err = os.Remove(dirName)
	if err != nil {
		t.Fatalf("Unable to delete tmp dir: '%s'", dirName)
	}
	t.Logf("Deleted tmp directory '%s'", dirName)

	isValid, errStr := validateUserDir(dirName)
	t.Logf("isValid: %v errStr: %v", isValid, errStr)
	if isValid || errStr == "" {
		t.Fail()
	}

	// 2. Test path points to a directory
	err = os.MkdirAll(dirName, 0755)
	if err != nil {
		t.Fatalf("Unable to create test tmp dir: '%s'", dirName)
	}
	t.Logf("Created tmp directory '%s'", dirName)

	fileName := dirName + pathSep + "ParserTestFile.txt"
	_, err = os.Create(fileName)
	defer os.Remove(fileName)
	if err != nil {
		t.Fatalf("Unable to create test tmp file: '%s'.  Error %v\n", fileName, err)
	}
	t.Logf("Created test tmp file: '%s'", fileName)

	isValid, errStr = validateUserDir(fileName)
	t.Logf("isValid: %v errStr: %v", isValid, errStr)
	if isValid || errStr == "" {
		t.Fail()
	}
}

func TestIsImageMagicConvertInPath(t *testing.T) {
	// Test the command recoginition in lieu of an actual 'convert' image
	exists := isImagicConvertInPath("echo")
	if !exists {
		t.Fail()
	}
	exists = isImagicConvertInPath("blah")
	if exists {
		t.Fail()
	}
}
