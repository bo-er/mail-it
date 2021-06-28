package user

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/bo-er/mail-it/mail"
)

type MailBrief struct {
	Commenter string
	IssueID   string
	Link      string
	Project   string
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
	fmt.Println(scanner.Text())
	for scanner.Scan() {
		if mailBrief.isComplete() {
			return
		}
		newline := scanner.Text()
		fmt.Println("NEWLINEIS", newline)
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
		strings.Trim
		trimedLine := strings.TrimSpace(strings.TrimLeft(newline, ">"))
		fmt.Println("TRIMED", trimedLine)
		if mailBrief.IssueID == "" {
			if index := strings.Index(trimedLine, ">键值:"); index != -1 {
				mailBrief.IssueID = trimedLine[index:]
			}
		}
		if mailBrief.Link == "" {
			if index := strings.Index(trimedLine, ">网址:"); index != -1 {
				mailBrief.Link = trimedLine[index:]
			}
		}
		if mailBrief.Project == "" {
			if index := strings.Index(trimedLine, ">项目:"); index != -1 {
				mailBrief.Project = trimedLine[index:]
			}
		}
		if mailBrief.IssueType == "" {
			if index := strings.Index(trimedLine, ">问题类型:"); index != -1 {
				mailBrief.IssueType = trimedLine[index:]
			}
		}
		if mailBrief.Reporter == "" {
			if index := strings.Index(trimedLine, ">报告人:"); index != -1 {
				mailBrief.Reporter = trimedLine[index:]
			}
		}
		if mailBrief.Assignee == "" {
			if index := strings.Index(trimedLine, ">经办人:"); index != -1 {
				mailBrief.Assignee = trimedLine[index:]
			}
		}
		if mailBrief.Tag == "" {
			if index := strings.Index(trimedLine, ">标签:"); index != -1 {
				mailBrief.Tag = trimedLine[index:]
			}
		}
		if mailBrief.Version == "" {
			if index := strings.Index(trimedLine, ">修复:"); index != -1 {
				mailBrief.Version = trimedLine[index:]
			}
		}

	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
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
		mb.Project != "" &&
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
