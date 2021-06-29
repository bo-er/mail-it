package user

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/bo-er/mail-it/mail"
)

var chineseReg = regexp.MustCompile("[^\u4e00-\u9fa5]")

type MailBrief struct {
	Operator  string
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

		trimedLine := strings.TrimSpace(strings.TrimLeft(newline, ">"))
		fmt.Println("TRIMED", trimedLine)
		if mailBrief.IssueID == "" {
			if index := strings.Index(trimedLine, "键值:"); index != -1 {
				mailBrief.IssueID = trimedLine[index:]
				continue
			}
		}
		if mailBrief.Link == "" {
			if index := strings.Index(trimedLine, "网址:"); index != -1 {
				mailBrief.Link = trimedLine[index:]
				continue
			}
		}
		if mailBrief.Project == "" {
			if index := strings.Index(trimedLine, "项目:"); index != -1 {
				mailBrief.Project = trimedLine[index:]
				continue
			}
		}
		if mailBrief.IssueType == "" {
			if index := strings.Index(trimedLine, "问题类型:"); index != -1 {
				mailBrief.IssueType = trimedLine[index:]
				continue
			}
		}
		if mailBrief.Reporter == "" {
			if index := strings.Index(trimedLine, "报告人:"); index != -1 {
				mailBrief.Reporter = trimedLine[index:]
				continue
			}
		}
		if mailBrief.Assignee == "" {
			if index := strings.Index(trimedLine, "经办人:"); index != -1 {
				mailBrief.Assignee = trimedLine[index:]
				continue
			}
		}
		if mailBrief.Tag == "" {
			if index := strings.Index(trimedLine, "标签:"); index != -1 {
				mailBrief.Tag = trimedLine[index:]
				continue
			}
		}
		if mailBrief.Version == "" {
			if index := strings.Index(trimedLine, "修复:"); index != -1 {
				mailBrief.Version = trimedLine[index:]
				continue
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
	return mb.Operator != "" &&
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

func GenerateBriefEmail(mb *MailBrief, mailBody []byte, opts ...Extract) *MailBrief {
	for _, opt := range opts {
		opt(mb, mailBody)
	}
	return mb
}

// Extract a type of function that extracts information from content
// and set that piece of information into MailBrief
type Extract func(mb *MailBrief, mailBody []byte) *MailBrief

func ExtractOperator() Extract {
	return func(mb *MailBrief, mailBody []byte) *MailBrief {
		regexString := `]([\s\S]*)-+`
		compiledRegxp := regexp.MustCompile(regexString)
		match := compiledRegxp.Find(mailBody)
		if match == nil {
			fmt.Fprintf(os.Stderr, "we don't find a operator\n")
			return mb
		}
		trimedMatch := bytes.TrimSpace(match)

		mb.Operator = extractChienes(chineseReg, match[:bytes.Index(trimedMatch, []byte{' '})])
		return mb
	}
}

func ExtractIssueIDAndProject() Extract {
	return func(mb *MailBrief, mailBody []byte) *MailBrief {
		regexString := `>(\s+)键值:\s.*\s`
		compiledRegxp := regexp.MustCompile(regexString)
		match := compiledRegxp.Find(mailBody)
		if match == nil {
			fmt.Fprintf(os.Stderr, "we didn't find an issueID\n")
			return mb
		}
		target := trimBytesPrefix(match)
		whiteSpaceIndex := bytes.Index(target, []byte{' '})
		hyphenIndex := bytes.Index(target, []byte{'-'})
		mb.IssueID = string(target[whiteSpaceIndex+1:])
		mb.Project = string(target[whiteSpaceIndex+1 : hyphenIndex])
		return mb
	}

}

func ExtractLink() Extract {
	return func(mb *MailBrief, mailBody []byte) *MailBrief {
		regexString := `>(\s+)网址:\s.*\s`
		compiledRegxp := regexp.MustCompile(regexString)
		match := compiledRegxp.Find(mailBody)
		if match == nil {
			fmt.Fprintf(os.Stderr, "we didn't find an web address\n")
			return mb
		}
		target := trimBytesPrefix(match)
		whiteSpaceIndex := bytes.Index(target, []byte{' '})
		mb.Link = string(target[whiteSpaceIndex+1:])
		return mb
	}

}

func ExtractAssignee() Extract {
	return func(mb *MailBrief, mailBody []byte) *MailBrief {
		regexString := `>(\s+)经办人:\s.*\s`
		compiledRegxp := regexp.MustCompile(regexString)
		match := compiledRegxp.Find(mailBody)
		if match == nil {
			fmt.Fprintf(os.Stderr, "we didn't find an assignee\n")
			return mb
		}
		target := trimBytesPrefix(match)
		whiteSpaceIndex := bytes.Index(target, []byte{' '})
		mb.Assignee = string(target[whiteSpaceIndex+1:])
		return mb
	}

}

func ExtractVersion() Extract {
	return func(mb *MailBrief, mailBody []byte) *MailBrief {
		regexString := `>(\s+)修复:\s.*\s`
		compiledRegxp := regexp.MustCompile(regexString)
		match := compiledRegxp.Find(mailBody)
		if match == nil {
			fmt.Fprintf(os.Stderr, "we didn't find an version\n")
			return mb
		}
		target := trimBytesPrefix(match)
		whiteSpaceIndex := bytes.Index(target, []byte{' '})
		mb.Version = string(target[whiteSpaceIndex+1:])
		return mb
	}

}

func ExtractReporter() Extract {
	return func(mb *MailBrief, mailBody []byte) *MailBrief {
		regexString := `>(\s+)报告人:\s.*\s`
		compiledRegxp := regexp.MustCompile(regexString)
		match := compiledRegxp.Find(mailBody)
		if match == nil {
			fmt.Fprintf(os.Stderr, "we didn't find an reporter\n")
			return mb
		}
		target := trimBytesPrefix(match)
		whiteSpaceIndex := bytes.Index(target, []byte{' '})
		mb.Reporter = string(target[whiteSpaceIndex+1:])
		return mb
	}

}

func ExtractIssueType() Extract {
	return func(mb *MailBrief, mailBody []byte) *MailBrief {
		regexString := `>(\s+)问题类型:\s.*\s`
		compiledRegxp := regexp.MustCompile(regexString)
		match := compiledRegxp.Find(mailBody)
		if match == nil {
			fmt.Fprintf(os.Stderr, "we didn't find an issue type\n")
			return mb
		}
		target := trimBytesPrefix(match)
		whiteSpaceIndex := bytes.Index(target, []byte{' '})
		mb.IssueType = string(target[whiteSpaceIndex+1:])
		return mb
	}

}

func ExtractTag() Extract {
	return func(mb *MailBrief, mailBody []byte) *MailBrief {
		regexString := `>(\s+)标签:\s.*\s`
		compiledRegxp := regexp.MustCompile(regexString)
		match := compiledRegxp.Find(mailBody)
		if match == nil {
			fmt.Fprintf(os.Stderr, "we didn't find an version\n")
			return mb
		}
		target := trimBytesPrefix(match)
		whiteSpaceIndex := bytes.Index(target, []byte{' '})
		mb.Tag = string(target[whiteSpaceIndex+1:])
		return mb
	}

}

func ExtractEffectiveBody() Extract {
	return func(mb *MailBrief, mailBody []byte) *MailBrief {
		regexString := `(---)+([\s\S]*)>\s(---)+`
		compiledRegxp := regexp.MustCompile(regexString)
		match := compiledRegxp.Find(mailBody)
		if match == nil {
			fmt.Fprintf(os.Stderr, "we didn't find an effective body\n")
			return mb
		}

		mb.BriefBody = string(trimLineWithGreaterPrefix(match))
		return mb
	}
}

func extractChienes(r *regexp.Regexp, content []byte) string {
	return r.ReplaceAllString(string(content), "")
}

func trimBytesPrefix(bs []byte) []byte {
	return bytes.TrimSpace(bytes.Trim(bs, ">- "))
}

func trimLineWithGreaterPrefix(bs []byte) []byte {
	trimedBytes := trimBytesPrefix(bs)
	firstGreaterSignIndex := bytes.Index(trimedBytes, []byte{'>'})
	return bytes.TrimSpace(trimedBytes[:firstGreaterSignIndex])
}

func ParseEmailV2(m mail.Email) (mailBrief *MailBrief, err error) {
	b := &MailBrief{}
	bsArray, err := m.VisibleText()
	if err != nil {
		return b, err
	}
	emailBody := bsArray[0]

	return GenerateBriefEmail(b, emailBody,
		ExtractAssignee(),
		ExtractIssueIDAndProject(),
		ExtractIssueType(),
		ExtractLink(),
		ExtractOperator(),
		ExtractTag(),
		ExtractReporter(),
		ExtractEffectiveBody()), nil

}

//func GetEffectiveTimeLineOfIssue(assignee,issueID string)string{
//
//}
