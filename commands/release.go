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

package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/hugo/release"
)

type releaser struct {
	cmd *cobra.Command

	// Will be zero for main releases.
	patchLevel int
}

func createReleaser() *releaser {
	r := &releaser{
		cmd: &cobra.Command{
			Use:    "release",
			Short:  "Release a new version of Hugo.",
			Hidden: true,
		},
	}

	r.cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return r.release()
	}

	r.cmd.PersistentFlags().IntVarP(&r.patchLevel, "patch", "p", 0, "Patch level, defaults to 0 for main releases")

	return r
}

func (r *releaser) release() error {
	return release.New(r.patchLevel).Run()
}
