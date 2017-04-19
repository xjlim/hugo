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

package release

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var issueRe = regexp.MustCompile(`(?i)[Updates?|Closes?|Fix.*|See] #(\d+)`)

type gitInfo struct {
	Hash    string
	Author  string
	Subject string
	Body    string

	GitHubCommit *gitHubCommit
}

func (g gitInfo) Issues() []int {
	return extractIssues(g.Body)
}

func extractIssues(body string) []int {
	var i []int
	m := issueRe.FindAllStringSubmatch(body, -1)
	for _, mm := range m {
		issueID, err := strconv.Atoi(mm[1])
		if err != nil {
			continue
		}
		i = append(i, issueID)
	}
	return i
}

type gitInfos []gitInfo

func git(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git failed: %q: %q", err, out)
	}
	return string(out), nil
}

func getGitInfos() (gitInfos, error) {
	var g gitInfos

	log, err := gitLog()
	if err != nil {
		return g, err
	}

	log = strings.Trim(log, "\n\x1e'")
	entries := strings.Split(log, "\x1e")

	for _, entry := range entries {
		items := strings.Split(entry, "\x1f")
		gi := gitInfo{
			Hash:    items[0],
			Author:  items[1],
			Subject: items[2],
			Body:    items[3],
		}
		gc, err := fetchCommit(gi.Hash)
		if err == nil {
			gi.GitHubCommit = &gc
		}
		g = append(g, gi)
	}

	return g, nil
}

func gitLog() (string, error) {
	prevTag, err := gitShort("describe", "--tags", "--abbrev=0", "--always", "HEAD^")
	if err != nil {
		return "", err
	}
	log, err := git("log", "--pretty=format:%x1e%h%x1f%aE%x1f%s%x1f%b", "--abbrev-commit", prevTag+"..HEAD")
	if err != nil {
		return ",", err
	}

	return log, err
}

func gitShort(args ...string) (output string, err error) {
	output, err = git(args...)
	return strings.Replace(strings.Split(output, "\n")[0], "'", "", -1), err
}
