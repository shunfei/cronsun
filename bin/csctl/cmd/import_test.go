package cmd

import (
	"strings"
	"testing"
)

func TestCheckCrons(t *testing.T) {
	crontab := `
*/1 * * * * /usr/bine/echo hello 
* * * * * /usr/bin/ls 
* & * * * /usr/bin/php -v 
* * * * /usr/bin/go run main.go
`
	invalidCrons := checkCrons(strings.Split(crontab, "\n"))
	if len(invalidCrons) != 2 {
		t.Error("should have 2 cron expression,but get none.")
	}
	if invalidCrons[0] != "* & * * * /usr/bin/php -v" ||
		invalidCrons[1] != "* * * * /usr/bin/go run main.go" {
		t.Error("invalid cron expression should * & * * * /usr/bin/php -v and * * * * /usr/bin/go run main.go.")
	}
}
