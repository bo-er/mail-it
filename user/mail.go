package user

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/bo-er/mail-it/db"
	"github.com/bo-er/mail-it/mail"
	"github.com/bo-er/mail-it/models"
)

var chineseReg = regexp.MustCompile("[^\u4e00-\u9fa5]")
var redisRtore *db.RedisStore

func init() {
	redisRtore = db.NewRedisStore("", "", 0)
}

func GetRedisStore() db.EmailStore {
	return redisRtore
}

// FindEmailContent finds a content from an email body with given regexp
func FindEmailContent(body []byte, reg *regexp.Regexp) (matchResult string, err error) {
	return string(reg.Find(body)), nil
}

func GenerateBriefEmail(m mail.Email, opts ...Extract) (*models.MailBrief, error) {
	b := &models.MailBrief{}
	bsArray, err := m.VisibleText()
	if err != nil {
		return b, err
	}
	mailBody := bsArray[0]
	ExtractEmailUsefulInfo(&m, b)
	for _, opt := range opts {
		opt(b, mailBody)
	}
	return b, nil
}

// Extract a type of function that extracts information from content
// and set that piece of information into models.MailBrief
// It's used only for parsing jira emails!
type Extract func(mb *models.MailBrief, mailBody []byte) *models.MailBrief

func ExtractOperator() Extract {
	return func(mb *models.MailBrief, mailBody []byte) *models.MailBrief {
		if mb.MailType != Jira {
			return mb
		}
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
	return func(mb *models.MailBrief, mailBody []byte) *models.MailBrief {
		if mb.MailType != Jira {
			return mb
		}
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
	return func(mb *models.MailBrief, mailBody []byte) *models.MailBrief {
		if mb.MailType != Jira {
			return mb
		}
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
	return func(mb *models.MailBrief, mailBody []byte) *models.MailBrief {
		if mb.MailType != Jira {
			return mb
		}
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
	return func(mb *models.MailBrief, mailBody []byte) *models.MailBrief {
		if mb.MailType != Jira {
			return mb
		}
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
	return func(mb *models.MailBrief, mailBody []byte) *models.MailBrief {
		if mb.MailType != Jira {
			return mb
		}
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
	return func(mb *models.MailBrief, mailBody []byte) *models.MailBrief {
		if mb.MailType != Jira {
			return mb
		}
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
	return func(mb *models.MailBrief, mailBody []byte) *models.MailBrief {
		if mb.MailType != Jira {
			return mb
		}
		regexString := `>(\s+)标签:\s.*\s`
		compiledRegxp := regexp.MustCompile(regexString)
		match := compiledRegxp.Find(mailBody)
		if match == nil {
			fmt.Fprintf(os.Stderr, "we didn't find an tag\n")
			return mb
		}
		target := trimBytesPrefix(match)
		whiteSpaceIndex := bytes.Index(target, []byte{' '})
		mb.Tag = string(target[whiteSpaceIndex+1:])
		return mb
	}

}

func ExtractEffectiveBody() Extract {
	return func(mb *models.MailBrief, mailBody []byte) *models.MailBrief {
		if mb.MailType != Jira {
			return mb
		}
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

func ParseEmail(m mail.Email) (mb *models.MailBrief, err error) {
	return GenerateBriefEmail(m,
		ExtractAssignee(),
		ExtractIssueIDAndProject(),
		ExtractIssueType(),
		ExtractLink(),
		ExtractOperator(),
		ExtractTag(),
		ExtractReporter(),
		ExtractEffectiveBody())

}

const JIRA_ADDRESS = "rdreport@actionsky.com"

// filterEmailFromJira is not usable due to all 'From' is 'unknown@example.com'
func filterEmailFromJira(mails []mail.Email) []mail.Email {
	for i := 0; i < len(mails); {
		if mails[i].From.Address != JIRA_ADDRESS {
			mails[i] = mails[len(mails)-1]
			mails = mails[:len(mails)-1]
		} else {
			i++
		}
	}
	return mails
}

const (
	Jira       string = "JIRA"
	Gitlab            = "GitLab"
	Confluence        = "Confluence"
	Slack             = "Slack"
)

var emailTypeMap = map[string]string{
	"[ACTION-JIRA]":       Jira,
	"Re:":                 Gitlab,
	"[action-confluence]": Confluence,
	"[Slack]":             Slack,
}
var MailBriefTimeFormat = "2006-01-02 15:04:05"

func ExtractEmailUsefulInfo(m *mail.Email, bm *models.MailBrief) {
	for k, v := range emailTypeMap {
		if strings.Contains(m.Subject, k) {
			bm.MailType = v
		}
	}
	bm.Time = m.InternalDate.Local().Format(MailBriefTimeFormat)
	bm.UID = m.UID
}

func SaveIssueFullMailBriefs(s db.EmailStore, mbs []*models.MailBrief) {

}
