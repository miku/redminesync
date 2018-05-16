package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/miku/redminesync"
	log "github.com/sirupsen/logrus"
)

var usageMessage = `redminesyncfiles [-f ID] [-t ID] -d DIRECTORY

Downloads all reachable attachements from redmine into a local folder. The
target folder structure will look like:

    rsf/123/download/456/file.txt

Where 123 is the issue number and 456 the download id.

  -f INT    start with this issue number, might shorten the process
  -t INT    end with this issue number, might shorten the process
`

var (
	startIssueNumber = flag.Int("f", 1, "start issue number")
	endIssueNumber   = flag.Int("t", -1, "end issue number, -1 means automatically find the max issue number")
	syncDir          = flag.String("d", filepath.Join(UserHomeDir(), ".redminesync"), "sync directory")
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

// UserHomeDir returns the home directory of the user.
func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

// downloadFile saves the contents of a URL to a file. The directory the files
// is in must exist.
func downloadFile(link, filepath string) (err error) {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	log.Printf("downloaded [%d]: %s", n, link)
	return nil
}

func downloadAttachment(link, rootDirectory string) error {
	u, err := url.Parse(link)
	if err != nil {
		return err
	}
	dst := filepath.Join(rootDirectory, u.Path)
	dstDir := filepath.Dir(dst)
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			log.Fatal(err)
		}
		log.Printf("created directory: %s", dstDir)
	}
	return downloadFile(link, dst)
}

func main() {
	flag.Parse()

	log.Printf("syncing redmine attachements to %s", *syncDir)

	if *endIssueNumber == -1 {
		maxIssue, err := redminesync.FindMaxIssue()
		if err != nil {
			log.Fatal(err)
		}
		*endIssueNumber = maxIssue
		log.Printf("found max issue number: %d", maxIssue)
	}

	for i := *startIssueNumber; i <= *endIssueNumber; i++ {
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
			// fmt.Printf("% 5d\t%6d\t% 10d\t%s\n", i, attachment.Id, attachment.Filesize, attachment.ContentUrl)
			if err := downloadAttachment(attachment.ContentUrl, *syncDir); err != nil {
				log.Fatal(err)
			}
		}
	}
}
