package user

import (
	"regexp"
)

type MailBreif struct {
	Commenter string
	IssueID   string
	Link      string
	Assignee  string
	Version   string
	Body      []byte
}

// FindEmailContent finds a content from an email body with given regexp
func FindEmailContent(body []byte, reg *regexp.Regexp) (matchResult string, err error) {
	return string(reg.Find(body)), nil
}

func ParseEmail

//func GetEffectiveTimeLineOfIssue(assignee,issueID string)string{
//
//}
