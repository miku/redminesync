# redminesync

Download attachments from Redmine.

By default, download all available downloads.

```
$ redminesync -h
redminesync [-k apikey] [-b URL] [-f ID] [-t ID] [-d DIRECTORY] [-verbose] [-P]

Downloads all reachable attachments from redmine into a local folder. The
target folder structure will look like:

    $HOME/.redminesync/123/456/file.txt

Where 123 is the issue number and 456 the download id.

  -b URL          redmine base url (default: https://projects.localhost)
  -k KEY          redmine api key [b345678931234567899111111111234567894367]
  -d DIRECTORY    target directory (default: $HOME/.redminesync)
  -f INT          start with this issue number, might shorten the process
  -t INT          end with this issue number, might shorten the process
  -verbose        be verbose
  -P              show progressbar

Limitation: Currently all ticket ids are rechecked on every invocation, since any tickets might have a new upload.
```
