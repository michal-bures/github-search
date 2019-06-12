package searcher

import (
	"context"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"os"
)

var accessTokenEnvVariable = "GITHUB_API_ACCESS_TOKEN"

type GithubSearcher struct {
	Client *github.Client
}

func (engine GithubSearcher) Search(keyword string, language string) (*[]github.CodeResult, error) {
	c := engine.Client
	options := &github.SearchOptions{
		TextMatch: true,
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	result, _, err := c.Search.Code(context.Background(), keyword+" language:"+language, options)

	if err != nil {
		return nil, err
	}

	return &result.CodeResults, nil
}

func NewSearcher() *GithubSearcher {
	return &GithubSearcher{
		Client: initGithubClient(),
	}
}

func initGithubClient() *github.Client {
	githubAccessToken, foundToken := os.LookupEnv(accessTokenEnvVariable)
	if !foundToken {
		panic("missing env variable " + accessTokenEnvVariable)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubAccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}
