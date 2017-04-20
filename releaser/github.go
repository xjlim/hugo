package releaser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

var gitHubCommitsApi = "https://api.github.com/repos/spf13/hugo/commits/%s"

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

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return commit, err
	}
	gitHubToken := os.Getenv("GITHUB_TOKEN")
	if gitHubToken != "" {
		req.Header.Add("Authorization", "token "+gitHubToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return commit, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := ioutil.ReadAll(resp.Body)
		return commit, fmt.Errorf("GitHub lookup failed: %s", string(b))

	}

	err = json.NewDecoder(resp.Body).Decode(&commit)
	if err != nil {
		return commit, err
	}

	return commit, nil
}
