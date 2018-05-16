package redminesync

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// MaxIssueNumber for probing the real maximum.
const MaxIssueNumber = 200000

// FindMaxIssue returns the maximum issue id by probing the API.
func FindMaxIssue() (int, error) {
	return findMax(0, MaxIssueNumber)
}

func findMax(a, b int) (result int, err error) {
	log.Printf("searching max issue number: %d %d", a, b)
	mid := a + (b-a)/2
	if a == b || a == mid {
		return a, nil
	}
	issueNo, baseURL := fmt.Sprintf("%d", mid), "https://projekte.ub.uni-leipzig.de"
	link := fmt.Sprintf("%s/issues/%s.json", baseURL, issueNo)
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Add("X-Redmine-API-Key", "a244016e302926344f08afb18c0af1b5b581ea67")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		result, err = findMax(a, mid)
	} else {
		result, err = findMax(mid, b)
	}
	return
}
