package user

import (
	"regexp"
)

// FindEmailContent finds a content from an email body with given regexp
func FindEmailContent(body []byte, reg *regexp.Regexp) (matchResult string, err error) {
	return string(reg.Find(body)), nil
}

//func GetEffectiveTimeLineOfIssue(assignee,issueID string)string{
//
//}