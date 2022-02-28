package utils

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

/*
 * WARNING: This is a basic and temporary file based implementation to handle local state
 * and currently does not yet support multiple process concurrency or the ability to clean
 * up stale data from processes that did not gracefully shutdown. The local state file is
 * not used anymore. This code is still important to make sure that file is deleted on
 * uninstall.
 */

import (
	"os"
	"path/filepath"
	"time"

	"github.com/nightlyone/lockfile"
)

var (
	runDataFile     string = "app-run-data.ldj"
	runDataLockFile string = "app-run-data.lock"
)

type RunData struct {
	SvcRunID    string
	AppHTTPPort int
	AppGRPCPort int
	AppID       string
	AppPort     int
	Command     string
	Created     time.Time
	PID         int
}

// DeleteRunDataFile deletes the deprecated RunData file.
func DeleteRunDataFile() error {
	lockFile, err := tryGetRunDataLock()
	if err != nil {
		return err
	}
	defer lockFile.Unlock()

	runFilePath := filepath.Join(os.TempDir(), runDataFile)
	err = os.Remove(runFilePath)
	if err != nil {
		return err
	}

	return nil
}

func tryGetRunDataLock() (*lockfile.Lockfile, error) {
	lockFile, err := lockfile.New(filepath.Join(os.TempDir(), runDataLockFile))
	if err != nil {
		// TODO: Log once we implement logging
		return nil, err
	}

	for i := 0; i < 10; i++ {
		err = lockFile.TryLock()

		// Error handling is essential, as we only try to get the lock.
		if err == nil {
			return &lockFile, nil
		}

		time.Sleep(50 * time.Millisecond)
	}

	return nil, err
}
