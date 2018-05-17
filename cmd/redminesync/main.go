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

var usageMessage = fmt.Sprintf(`redminesync [-k apikey] [-b URL] [-f ID] [-t ID] [-d DIRECTORY]

Downloads all reachable attachments from redmine into a local folder. The
target folder structure will look like:

    %s/123/download/456/file.txt

Where 123 is the issue number and 456 the download id.

  -b URL          redmine base url (default: %s)
  -k KEY          redmine api key [%s]
  -d DIRECTORY    target directory (default: %s)
  -f INT          start with this issue number, might shorten the process
  -t INT          end with this issue number, might shorten the process

`, *syncDir, *baseURL, *apiKey, *syncDir)

var (
	startIssueNumber = flag.Int("f", 1, "start issue number")
	endIssueNumber   = flag.Int("t", 0, "end issue number, 0 means automatically find the max issue number")
	syncDir          = flag.String("d", filepath.Join(UserHomeDir(), ".redminesync"), "sync directory")
	apiKey           = flag.String("k", os.Getenv("REDMINE_API_KEY"), "redmine API key possible from envvar REDMINE_API_KEY")
	baseURL          = flag.String("b", "https://projekte.ub.uni-leipzig.de", "base URL")
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
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		out, err := os.Create(filepath)
		if err != nil {
			return err
		}
		defer out.Close()

		req, err := http.NewRequest("GET", link, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("X-Redmine-API-Key", *apiKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err)
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
	} else {
		log.Printf("already downloaded: %s", filepath)
	}
	return nil
}

func downloadAttachment(link, rootDirectory string, issue int) error {
	u, err := url.Parse(link)
	if err != nil {
		return err
	}
	dst := filepath.Join(rootDirectory, fmt.Sprintf("%d", issue), u.Path)
	dstDir := filepath.Dir(dst)
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			log.Fatal(err)
		}
	}
	return downloadFile(link, dst)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usageMessage)
	}
	flag.Parse()
	log.Printf("syncing redmine attachments to %s", *syncDir)

	if *endIssueNumber == 0 {
		maxIssue, err := redminesync.FindMaxIssue(*baseURL, *apiKey)
		if err != nil {
			log.Fatal(err)
		}
		*endIssueNumber = maxIssue
		log.Printf("found max issue number: %d", maxIssue)
	}

	for i := *startIssueNumber; i <= *endIssueNumber; i++ {
		issueNo := fmt.Sprintf("%d", i)
		link := fmt.Sprintf("%s/issues/%s.json?include=attachments", *baseURL, issueNo)

		req, err := http.NewRequest("GET", link, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("X-Redmine-API-Key", *apiKey)

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
			if err := downloadAttachment(attachment.ContentUrl, *syncDir, i); err != nil {
				log.Fatal(err)
			}
		}
	}
}
