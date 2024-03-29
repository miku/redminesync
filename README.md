# redminesync

Download attachments from Redmine, you will need your [API key](https://redmine.example.com/my/api_key).

```
$ go install github.com/miku/redminesync/cmd/redminesync@latest
$ redminesync -verbose -k 123412341234123412341234 -b https://redmine.example.com
```

By default, it will fetch all available downloads.

```
$ redminesync -h
redminesync [-k apikey] [-b URL] [-f ID] [-t ID] [-d DIRECTORY] [-verbose] [-P]

Downloads all reachable attachments from redmine into a local folder. The
target folder structure will look like:

    $HOME/.cache/redminesync/123/456/file.txt

Where 123 is the issue number and 456 the download id.

  -b URL          redmine base url (default: https://projects.localhost)
  -k KEY          redmine api key [b345678931234567899111111111234567894367]
  -d DIRECTORY    target directory (default: $HOME/.redminesync)
  -f INT          start with this issue number, might shorten the process
  -t INT          end with this issue number, might shorten the process
  -verbose        be verbose
  -P              show progressbar

Limitation: Currently all ticket ids are rechecked on every invocation, since
any tickets might have a new upload.
```
