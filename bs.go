package redminesync

import (
	"fmt"
	"net/http"
)

// MaxIssueNumber for probing the real maximum. It will be troublesome, if a
// ticket system has actually this number of issues.
const MaxIssueNumber = 1000000

// FindMaxIssue returns the maximum issue id by probing the API.
func FindMaxIssue(baseURL, apiKey string) (int, error) {
	return findMax(0, MaxIssueNumber, baseURL, apiKey)
}

func findMax(a, b int, baseURL, apiKey string) (result int, err error) {
	mid := a + (b-a)/2
	if a == b || a == mid {
		if a == MaxIssueNumber-1 {
			return a, fmt.Errorf("could not find number of issues, probably due to insufficient access")
		}
		return a, nil
	}
	issueNo := fmt.Sprintf("%d", mid)
	link := fmt.Sprintf("%s/issues/%s.json", baseURL, issueNo)
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Add("X-Redmine-API-Key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		result, err = findMax(a, mid, baseURL, apiKey)
	} else {
		result, err = findMax(mid, b, baseURL, apiKey)
	}
	return
}
