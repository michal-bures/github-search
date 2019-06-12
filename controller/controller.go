package controller

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"githubsearch/pages"
	"githubsearch/refiners"
	"githubsearch/searcher"
	"log"
	"net/http"
	"time"
)

type Controller struct {
	searcher *searcher.GithubSearcher
}

func (c Controller) IndexPageHandler(w http.ResponseWriter, r *http.Request) {
	keyword := getQueryParam(r, "search")
	language := getQueryParam(r, "language")
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	errorHandler:= func(e error) {
		handleError(w, e)
	}

	if keyword == "" {
		pages.SearchPage(w, errorHandler, pages.SearchPageData{
			ShowResults: false,
		})
	} else {
		results, err := c.searcher.Search(ctx, keyword, language)
		handleError(w, err)

		pipeline := [...]refiners.SearchResultsRefiner{
			refiners.MatchPattern{Pattern: keyword},
			refiners.SortByRepositoryScore{MaxRequests: 20, Client: (*c.searcher).Client},
		}

		for _, refinement := range pipeline {
			results = refinement.Apply(ctx, results)
		}

		pages.SearchPage(w, errorHandler, pages.SearchPageData{
			ShowResults:    true,
			SearchLanguage: language,
			SearchString:   r.URL.Query()["search"][0],
			Results:        convertResults(results),
		})
	}
}

func handleError(w http.ResponseWriter, e error) {
	if e != nil {
		log.Print(errors.WithStack(e))
		w.WriteHeader(500)
		fmt.Fprint(w, "Oops, something went wrong")
	}
}

func convertResults(codeResults *[]github.CodeResult) []pages.SearchResult {
	searchResults := make([]pages.SearchResult, len(*codeResults))

	fmt.Printf("Total results: %d\n", len(*codeResults))

	for i, codeResult := range *codeResults {
		searchResults[i] = pages.SearchResult{
			Name:       codeResult.Name,
			Path:       codeResult.Path,
			FileUrl:    codeResult.HTMLURL,
			Repository: codeResult.Repository.Name,
			Fragments:  getFragments(codeResult.TextMatches),
		}
	}
	return searchResults
}

func getFragments(matches []github.TextMatch) []*string {
	fragments := make([]*string, len(matches))
	for i, match := range matches {
		fragments[i] = match.Fragment
	}
	return fragments
}

func NewController(searcher *searcher.GithubSearcher) *Controller {
	return &Controller{
		searcher: searcher,
	}
}

func getQueryParam(r *http.Request, name string) string {
	values := r.URL.Query()[name]
	if values == nil {
		return ""
	}
	return values[0]
}
