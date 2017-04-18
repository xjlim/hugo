// Copyright 2017-present The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package commands defines and implements command-line commands and flags
// used by Hugo. Commands and flags are implemented using Cobra.

package release

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/hugo/helpers"
)

type ReleaseHandler struct {
	patch int
}

func New(patch int) *ReleaseHandler {
	return &ReleaseHandler{patch: patch}
}

func (r *ReleaseHandler) Run() error {
	var version helpers.HugoVersion

	if r.patch > 0 {
		version = helpers.CurrentHugoVersion.NextPatchLevel(r.patch)
	} else {
		version = helpers.CurrentHugoVersion.Next()
	}

	if confirm(fmt.Sprint("Start release of ", version, "?")) {
		fmt.Println("Start")

		// Plan:
		// Release notes?
		// Adapt version numbers
		// Commit
		// Tag
		tag := "v" + version.String()
		// Run goreleaser
		// Prepare version numbers for next release.
		// Commit.

	}
	return nil
}

func confirm(s string) bool {
	r := bufio.NewReader(os.Stdin)
	tries := 10

	for ; tries > 0; tries-- {
		fmt.Printf("%s [y/n]: ", s)

		res, err := r.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if len(res) < 2 {
			continue
		}

		return strings.ToLower(strings.TrimSpace(res))[0] == 'y'
	}

	return false
}

func git(args ...string) (output string, err error) {
	var cmd = exec.Command("git", args...)
	outputs, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git failed: %q", err)
	}
	return string(outputs), nil
}
