package user

import (
	"fmt"
	"github.com/bo-er/mail-it/mail"
	"regexp"
)

// emailBodyFilter filters an email's body
type emailBodyFilter func([][]byte) bool

// GetLastWeekWork gets your last week work on jira.
func GetLastWeekWork(info mail.MailboxInfo, filter emailBodyFilter, keyMap map[string]interface{}, regexes []string) ([]string, error) {
	var maskAsRead, delete bool
	var projectCounterMap = map[string]struct{}{}
	var lastweekProjects []string
	mails, err := mail.GetWithKeyMap(info, keyMap, maskAsRead, delete)
	if err != nil {
		return nil, err
	}
	regexps := make([]*regexp.Regexp, len(regexes))
	for index, r := range regexes {
		regexps[index] = regexp.MustCompile(r)
	}
	fmt.Println("邮件数量", len(mails))
	for _, mail := range mails {
		body,_ := mail.VisibleText()
		ok := filter(body)
		if !ok {
			continue
		}
		for _, r := range regexps {
			result, _ := FindEmailContent(body[0], r)
			if result == "" {
				continue
			}
			projectCounterMap[result] = struct{}{}
		}
	}

	for project, _ := range projectCounterMap {
		lastweekProjects = append(lastweekProjects, project)
	}
	return lastweekProjects, nil
}
