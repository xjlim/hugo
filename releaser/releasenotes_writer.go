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

// Package release implements a set of utilities and a wrapper around Goreleaser
// to help automate the Hugo release process.
package releaser

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

// TODO(bep)
// Work flow:
// Create draft release notes in docs/temp/<tag>-release-notes.md (folder)
// Commit
//

const (
	issueLinkTemplate            = "[#%d](https://github.com/spf13/hugo/issues/%d)"
	linkTemplate                 = "[%s](%s)"
	releaseNotesMarkdownTemplate = `
# Enhancements
{{ template "change-headers"  .Enhancements -}}
# Fixes
{{ template "change-headers"  .Fixes -}}

{{ define "change-headers" }}
{{ $tmplChanges := index . "templateChanges" -}}
{{- $outChanges := index . "outChanges" -}}
{{- $coreChanges := index . "coreChanges" -}}
{{- $docsChanges := index . "docsChanges" -}}
{{- $otherChanges := index . "otherChanges" -}}
{{- with $tmplChanges -}}
## Templates
{{ template "change-section" . }}
{{- end -}}
{{- with $outChanges -}}
## Output
{{- template "change-section"  . }}
{{- end -}}
{{- with $coreChanges -}}
## Core
{{ template "change-section" . }}
{{- end -}}
{{- with $docsChanges -}}
## Docs
{{- template "change-section"  . }}
{{- end -}}
{{- with $otherChanges -}}
## Other
{{ template "change-section"  . }}
{{- end -}}
{{ end }}


{{ define "change-section" }}
{{ range . }}
{{- if .GitHubCommit -}}
* {{ .Subject }} {{ . | commitURL }} {{ . | authorURL }} {{ range .Issues }}{{ . | issue }} {{ end }}
{{ else -}}
* {{ .Subject }} {{ range .Issues }}{{ . | issue }} {{ end }}
{{ end -}}
{{- end }}
{{ end }}
`
)

var templateFuncs = template.FuncMap{
	"issue": func(id int) string {
		return fmt.Sprintf(issueLinkTemplate, id, id)
	},
	"commitURL": func(info gitInfo) string {
		if info.GitHubCommit.HtmlURL == "" {
			return ""
		}
		return fmt.Sprintf(linkTemplate, info.Hash, info.GitHubCommit.HtmlURL)
	},
	"authorURL": func(info gitInfo) string {
		if info.GitHubCommit.Author.Login == "" {
			return ""
		}
		return fmt.Sprintf(linkTemplate, "@"+info.GitHubCommit.Author.Login, info.GitHubCommit.Author.HtmlURL)
	},
}

func writeReleaseNotes(infos gitInfos, to io.Writer) error {
	changes := gitInfosToChangeLog(infos)

	tmpl, err := template.New("").Funcs(templateFuncs).Parse(releaseNotesMarkdownTemplate)
	if err != nil {
		return err
	}

	err = tmpl.Execute(to, changes)
	if err != nil {
		return err
	}

	return nil

}

func writeReleaseNotesToTmpFile(infos gitInfos) (string, error) {
	f, err := ioutil.TempFile("", "hugorelease")
	if err != nil {
		return "", err
	}

	defer f.Close()

	if err := writeReleaseNotes(infos, f); err != nil {
		return "", err
	}

	return f.Name(), nil
}

func getRelaseNotesDocsTempDirAndName(tag string) (string, string) {
	return hugoFilepath("docs/temp"), fmt.Sprintf("%s-relnotes.md", tag)
}

func getRelaseNotesDocsTempFilename(tag string) string {
	return filepath.Join(getRelaseNotesDocsTempDirAndName(tag))
}

func writeReleaseNotesToDocsTemp(tag string, infos gitInfos) (string, error) {
	docsTempPath, name := getRelaseNotesDocsTempDirAndName(tag)
	os.Mkdir(docsTempPath, os.ModePerm)

	f, err := os.Create(filepath.Join(docsTempPath, name))
	if err != nil {
		return "", err
	}

	defer f.Close()

	if err := writeReleaseNotes(infos, f); err != nil {
		return "", err
	}

	return f.Name(), nil

}

func writeReleaseNotesToDocs(title, sourceFilename string) (string, error) {
	targetFilename := filepath.Base(sourceFilename)
	contentDir := hugoFilepath("docs/content/release-notes")
	targetFullFilename := filepath.Join(contentDir, targetFilename)
	os.Mkdir(contentDir, os.ModePerm)

	b, err := ioutil.ReadFile(sourceFilename)
	if err != nil {
		return "", err
	}

	f, err := os.Create(targetFullFilename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.WriteString(fmt.Sprintf(`
---
date: %s
title: %s
---

	`, time.Now().Format("2006-02-06"), title)); err != nil {
		return "", err
	}

	if _, err := f.Write(b); err != nil {
		return "", err
	}

	return targetFullFilename, nil

}
