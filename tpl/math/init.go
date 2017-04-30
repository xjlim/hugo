// Copyright 2017 The Hugo Authors. All rights reserved.
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

package math

import (
	"github.com/spf13/hugo/deps"
	"github.com/spf13/hugo/tpl/internal"
)

const name = "math"

func init() {
	f := func(d *deps.Deps) *internal.TemplateFuncsNamespace {
		ctx := New()

		examples := [][2]string{
			{"{{add 1 2}}", "3"},
			{"{{div 6 3}}", "2"},
			{"{{mod 15 3}}", "0"},
			{"{{modBool 15 3}}", "true"},
			{"{{mul 2 3}}", "6"},
			{"{{sub 3 2}}", "1"},
		}

		return &internal.TemplateFuncsNamespace{
			Name:    name,
			Context: func() interface{} { return ctx },
			Aliases: map[string]interface{}{
				"add":     ctx.Add,
				"div":     ctx.Div,
				"mod":     ctx.Mod,
				"modBool": ctx.ModBool,
				"mul":     ctx.Mul,
				"sub":     ctx.Sub,
			},
			Examples: examples,
		}

	}

	internal.AddTemplateFuncsNamespace(f)
}
