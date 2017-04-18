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
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	if os.Getenv("GITHUB_TOKEN") == "" {
		return errors.New("GITHUB_TOKEN not set, needed by goreleaser")
	}

	var (
		newVersion   helpers.HugoVersion
		finalVersion = helpers.CurrentHugoVersion
	)

	if r.patch > 0 {
		newVersion = helpers.CurrentHugoVersion.NextPatchLevel(r.patch)
	} else {
		newVersion = helpers.CurrentHugoVersion.Next()
		finalVersion = newVersion
		finalVersion.Suffix = "-DEV"
	}

	if true || confirm(fmt.Sprint("Start release of ", newVersion, "?")) {

		tag := "v" + newVersion.String()

		// Exit early if tag already exists
		out, err := git("tag", "-l", tag)

		if err != nil {
			return err
		}

		if strings.Contains(out, tag) {
			return fmt.Errorf("Tag %q already exists", tag)
		}

		// Plan:
		// Release notes?
		// OK Adapt version numbers
		// TODO(bep) push before release?
		if err := bumpVersions(newVersion); err != nil {
			return err
		}

		if _, err := git("commit", "-a", "-m", fmt.Sprintf("release: Bump versions for release of %s", newVersion)); err != nil {
			return err
		}

		if _, err := git("tag", "-a", tag, "-m", fmt.Sprintf("release: %s", newVersion)); err != nil {
			return err
		}

		if confirm("Release to GitHub") {
			if err := release(); err != nil {
				return err
			}
		}

		if err := bumpVersions(finalVersion); err != nil {
			return err
		}

		if _, err := git("commit", "-a", "-m", fmt.Sprintf("release: Prepare repository for %s", finalVersion)); err != nil {
			return err
		}

	}

	// Commit
	// Tag
	//
	// Run goreleaser
	// Prepare version numbers for next release.
	// Commit.

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

func release() error {
	cmd := exec.Command("goreleaser")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("goreleaser failed: %q: %q", err, out)
	}
	return nil
}

func git(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git failed: %q: %q", err, out)
	}
	return string(out), nil
}

func bumpVersions(ver helpers.HugoVersion) error {
	fromDev := ""
	toDev := ""

	if ver.Suffix != "" {
		toDev = "-DEV"
	} else {
		fromDev = "-DEV"
	}

	if err := replaceInFile("helpers/hugo.go",
		`Number:(\s{4,})(.*),`, fmt.Sprintf(`Number:${1}%.2f,`, ver.Number),
		`PatchLevel:(\s*)(.*),`, fmt.Sprintf(`PatchLevel:${1}%d,`, ver.PatchLevel),
		fmt.Sprintf(`Suffix:(\s{4,})"%s",`, fromDev), fmt.Sprintf(`Suffix:${1}"%s",`, toDev)); err != nil {
		return err
	}

	snapcraftGrade := "stable"
	if ver.Suffix != "" {
		snapcraftGrade = "devel"
	}
	if err := replaceInFile("snapcraft.yaml",
		`version: "(.*)"`, fmt.Sprintf(`version: "%s"`, ver),
		`grade: (.*) #`, fmt.Sprintf(`grade: %s #`, snapcraftGrade)); err != nil {
		return err
	}

	var minVersion string
	if ver.Suffix != "" {
		// People use the DEV version in daily use, and we cannot create new themes
		// with the next version before it is released.
		minVersion = ver.Prev().String()
	} else {
		minVersion = ver.String()
	}

	if err := replaceInFile("commands/new.go",
		`min_version = "(.*)"`, fmt.Sprintf(`min_version = "%s"`, minVersion)); err != nil {
		return err
	}

	// docs/config.toml
	if err := replaceInFile("docs/config.toml",
		`release = "(.*)"`, fmt.Sprintf(`release = "%s"`, ver)); err != nil {
		return err
	}

	return nil
}

func replaceInFile(filename string, oldNew ...string) error {
	fullFilename := hugoFilepath(filename)
	fi, err := os.Stat(fullFilename)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadFile(fullFilename)
	if err != nil {
		return err
	}
	newContent := string(b)

	for i := 0; i < len(oldNew); i += 2 {
		re := regexp.MustCompile(oldNew[i])
		newContent = re.ReplaceAllString(newContent, oldNew[i+1])
	}

	return ioutil.WriteFile(fullFilename, []byte(newContent), fi.Mode())

	return nil
}

func hugoFilepath(filename string) string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(pwd, filename)
}
