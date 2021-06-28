package user

import (
	"regexp"

	"github.com/bo-er/mail-it/mail"
)

type MailBreif struct {
	Commenter string
	IssueID   string
	Link      string
	IssueType string
	Assignee  string
	Report
	Version   string
	BriefBody []byte
}

// FindEmailContent finds a content from an email body with given regexp
func FindEmailContent(body []byte, reg *regexp.Regexp) (matchResult string, err error) {
	return string(reg.Find(body)), nil
}

func ParseEmail(m mail.Email) MailBreif {

}

//func GetEffectiveTimeLineOfIssue(assignee,issueID string)string{
//
//}
