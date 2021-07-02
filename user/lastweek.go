package user

import (
	"bytes"
	"regexp"
	"time"

	"github.com/bo-er/mail-it/mail"
)

// emailBodyFilter filters an email's body
type emailBodyFilter func([][]byte) bool

func FilterByNameOfAssignee(body [][]byte, username string) bool {
	return bytes.ContainsAny(body[0], "经办人: "+username) || bytes.ContainsAny(body[0], "Assignee: "+username)
}

func FilterByAssigneeAndIssueID(body [][]byte, username, issueID string) bool {
	return (bytes.ContainsAny(body[0], "经办人: "+username) || bytes.ContainsAny(body[0], "Assignee: "+username)) &&
		bytes.ContainsAny(body[0], "键值: "+issueID)
}

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
	for _, mail := range mails {
		body, _ := mail.VisibleText()
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


