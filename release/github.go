package release

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// TODO(bep) => spf13
const gitHubCommitsApi = "https://api.github.com/repos/bep/hugo/commits/%s"

type gitHubCommit struct {
	Author  gitHubAuthor `json:"author"`
	HtmlURL string       `json:"html_url"`
}

type gitHubAuthor struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	HtmlURL   string `json:"html_url"`
	AvatarURL string `json:"avatar_url"`
}

func fetchCommit(ref string) (gitHubCommit, error) {
	var commit gitHubCommit

	u := fmt.Sprintf(gitHubCommitsApi, ref)

	resp, err := http.Get(u)
	if err != nil {
		return commit, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&commit)
	if err != nil {
		return commit, err
	}

	return commit, nil
}
