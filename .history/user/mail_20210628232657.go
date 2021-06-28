package user

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"

	"github.com/bo-er/mail-it/mail"
)

type MailBrief struct {
	Commenter string
	IssueID   string
	Link      string
	IssueType string
	Assignee  string
	Reporter  string
	Tag       string
	Version   string
	BriefBody []byte
}

// FindEmailContent finds a content from an email body with given regexp
func FindEmailContent(body []byte, reg *regexp.Regexp) (matchResult string, err error) {
	return string(reg.Find(body)), nil
}

func ParseEmail(m mail.Email) (mailBrief MailBrief, err error) {
	contents, err := m.VisibleText()
	if err != nil {
		return mailBrief,err
	}
	scanner :=bufio.NewScanner(bytes.NewReader(contents[0]))
	var briefBody
	var startReadingBriefBody bool
	for scanner.Scan() { 
        if strings.HasPrefix("--") && strings.HasSuffix("--"){
			startReadingBriefBody = true
		}
    }
}

//func GetEffectiveTimeLineOfIssue(assignee,issueID string)string{
//
//}
