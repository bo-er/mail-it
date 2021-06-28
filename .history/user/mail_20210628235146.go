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
		if mailBrief.isComplete(){
			
		}
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
		if index := strings.Index(trimedLine, ">键值:"); index != -1 {
			mailBrief.IssueID = trimedLine[index:]
		}
		if index := strings.Index(trimedLine, ">网址:"); index != -1 {
			mailBrief.IssueID = trimedLine[index:]
		}
		if index := strings.Index(trimedLine, ">项目:"); index != -1 {
			mailBrief.IssueID = trimedLine[index:]
		}
		if index := strings.Index(trimedLine, ">问题类型:"); index != -1 {
			mailBrief.IssueID = trimedLine[index:]
		}
		if index := strings.Index(trimedLine, ">报告人:"); index != -1 {
			mailBrief.IssueID = trimedLine[index:]
		}
		if index := strings.Index(trimedLine, ">经办人:"); index != -1 {
			mailBrief.IssueID = trimedLine[index:]
		}
		if index := strings.Index(trimedLine, ">标签:"); index != -1 {
			mailBrief.IssueID = trimedLine[index:]
		}
		if index := strings.Index(trimedLine, ">修复:"); index != -1 {
			mailBrief.IssueID = trimedLine[index:]
		}
	}
	return
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

func (mb *MailBrief) isComplete() bool {
	return mb.Commenter != "" &&
		mb.IssueID != "" &&
		mb.Link != "" &&
		mb.IssueType != "" &&
		mb.Assignee != "" &&
		mb.Reporter != "" &&
		mb.Tag != "" &&
		mb.Version != "" &&
		mb.BriefBody != ""
}

//func GetEffectiveTimeLineOfIssue(assignee,issueID string)string{
//
//}
