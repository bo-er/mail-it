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
	BriefBody string
}

// FindEmailContent finds a content from an email body with given regexp
func FindEmailContent(body []byte, reg *regexp.Regexp) (matchResult string, err error) {
	return string(reg.Find(body)), nil
}

func ParseEmail(m mail.Email) (mailBrief MailBrief, err error) {
	contents, err := m.VisibleText()
	if err != nil {
		return mailBrief, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(contents[0]))
	var briefBody string
	var startReadingBriefBody bool
	for scanner.Scan() {
		newline := scanner.Text()
		if newline == "" {
			continue
		}
		stateChanged := setReadBriefBodyState(newline, briefBody, &startReadingBriefBody)
		if stateChanged {
			continue
		}
		if startReadingBriefBody {
			briefBody += newline
			continue
		}
		trimedLine := strings.TrimSpace(newline)
		if strings.Trim
	}
}

func setReadBriefBodyState(newline, briefBody string, state *bool) bool {
	if !*state && briefBody == "" &&
		strings.HasSuffix(newline, "--") && strings.HasPrefix(newline, "--") {
		*state = true
		return true
	}
	if *state && briefBody != "" && strings.HasPrefix(newline, ">") {
		*state = false
		return true
	}
	return false
}

//func GetEffectiveTimeLineOfIssue(assignee,issueID string)string{
//
//}
