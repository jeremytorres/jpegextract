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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"unsafe"
)

func isHostLittleEndian() bool {
	// From https://groups.google.com/forum/#!topic/golang-nuts/3GEzwKfRRQw
	var i int32 = 0x01020304
	u := unsafe.Pointer(&i)
	pb := (*byte)(u)
	b := *pb
	return (b == 0x04)
}

func exitWithErr() {
	os.Exit(1)
}

func isImagicConvertInPath(bin string) bool {
	cmd := exec.Command(bin)
	err := cmd.Start()
	if err != nil {
		return false
	}
	return true
}

func validateUserDir(dirStr string) (isValid bool, errStr string) {
	errStr = ""
	isValid = true
	dir, err := os.Open(dirStr)
	defer dir.Close()

	if err != nil {
		errStr = "Error: '%s': unable to open directory!\n"
		errStr = fmt.Sprintf(errStr, dirStr)
		isValid = false
	} else if i, e := dir.Stat(); e == nil && !i.IsDir() {
		errStr = "Error: '%s': is not a directory!\n"
		errStr = fmt.Sprintf(errStr, dirStr)
		isValid = false
	}

	return isValid, errStr
}

func getFilesForExt(ext string) ([]string, error) {
	return filepath.Glob(ext)
}
