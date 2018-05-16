package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// IssueResponse represents an issue, including various optional items, such as
// children, attachments, relations, changesets, journals and watchers
// (http://www.redmine.org/projects/redmine/wiki/Rest_Issues#Showing-an-issue).
type IssueResponse struct {
	Issue struct {
		AssignedTo struct {
			Id   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"assigned_to"`
		Attachments []struct {
			Author struct {
				Id   int64  `json:"id"`
				Name string `json:"name"`
			} `json:"author"`
			ContentType string `json:"content_type"`
			ContentUrl  string `json:"content_url"`
			CreatedOn   string `json:"created_on"`
			Description string `json:"description"`
			Filename    string `json:"filename"`
			Filesize    int64  `json:"filesize"`
			Id          int64  `json:"id"`
		} `json:"attachments"`
		Author struct {
			Id   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"author"`
		Changesets []struct {
			Comments    string `json:"comments"`
			CommittedOn string `json:"committed_on"`
			Revision    string `json:"revision"`
			User        struct {
				Id   int64  `json:"id"`
				Name string `json:"name"`
			} `json:"user"`
		} `json:"changesets"`
		CreatedOn    string `json:"created_on"`
		CustomFields []struct {
			Id    int64  `json:"id"`
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"custom_fields"`
		Description  string `json:"description"`
		DoneRatio    int64  `json:"done_ratio"`
		FixedVersion struct {
			Id   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"fixed_version"`
		Id       int64 `json:"id"`
		Journals []struct {
			CreatedOn string        `json:"created_on"`
			Details   []interface{} `json:"details"`
			Id        int64         `json:"id"`
			Notes     string        `json:"notes"`
			User      struct {
				Id   int64  `json:"id"`
				Name string `json:"name"`
			} `json:"user"`
		} `json:"journals"`
		Priority struct {
			Id   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"priority"`
		Project struct {
			Id   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"project"`
		StartDate string `json:"start_date"`
		Status    struct {
			Id   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"status"`
		Subject string `json:"subject"`
		Tracker struct {
			Id   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"tracker"`
		UpdatedOn string `json:"updated_on"`
	} `json:"issue"`
}

func main() {
	for i := 1; i < 14000; i++ {
		issueNo, baseURL := fmt.Sprintf("%d", i), "https://projekte.ub.uni-leipzig.de"
		link := fmt.Sprintf("%s/issues/%s.json?include=attachments", baseURL, issueNo)

		req, err := http.NewRequest("GET", link, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("X-Redmine-API-Key", "a244016e302926344f08afb18c0af1b5b581ea67")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode == 404 || resp.StatusCode == 403 {
			continue
		}
		if resp.StatusCode >= 400 {
			log.Fatalf("%s: %s", resp.Status, link)
		}

		var issue IssueResponse
		if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
			log.Fatalf("decode: %s", err)
		}
		for _, attachment := range issue.Issue.Attachments {
			fmt.Printf("% 5d\t%6d\t% 10d\t%s\n", i, attachment.Id, attachment.Filesize, attachment.ContentUrl)
		}
	}
}
