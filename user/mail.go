package user

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/bo-er/mail-it/db"
	"github.com/bo-er/mail-it/mail"
	"github.com/bo-er/mail-it/models"
)

var redisRtore *db.RedisStore

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

var once sync.Once
var regexpMap map[string]*regexp.Regexp

func init() {
	once.Do(
		func() {
			redisRtore = db.NewRedisStore("", "", 0)
			preCompileRegex()
		})
}

func preCompileRegex() {
	var regexMap = map[string]string{
		OperatorKey:          `(]\s*)?(.*:)(\s*---)`,
		IssueIDAndProjectKey: `>(\s+)键值:\s.*\s`,
		LinkKey:              `>(\s+)网址:\s.*\s`,
		AssigneeKey:          `>(\s+)经办人:\s.*\s`,
		VersionKey:           `(>\s+修复:\s)(([0-9]\.)+[0-9]{1})`,
		ReporterKey:          `>(\s+)报告人:\s.*\s`,
		IssueTypeKey:         `>(\s+)问题类型:\s.*\s`,
		TagKey:               `>(\s+)标签:\s.*\s`,
		EffectiveBodyKey:     `(---)+([\s\S]*)>\s(---)+`,
		Gitlab:               `(gitlab 在\s.*中留言:)|(gitlab\s更新了\s.*:)`,
		ChineseRegKey:        "[^\u4e00-\u9fa5]",
	}
	regexpMap = make(map[string]*regexp.Regexp, len(regexMap))
	for field, regexString := range regexMap {
		regexpMap[field] = regexp.MustCompile(regexString)
	}
}

const (
	OperatorKey          = "Operator"
	IssueIDAndProjectKey = "IssueIDAndProject"
	LinkKey              = "Link"
	AssigneeKey          = "Asignee"
	VersionKey           = "Version"
	ReporterKey          = "Reporter"
	IssueTypeKey         = "IssueType"
	TagKey               = "Tag"
	EffectiveBodyKey     = "EffectiveBody"
	GitlabKey            = "Gitlab"
	ChineseRegKey        = "ChineseReg"
)

const (
	JIRA_ADDRESS        = "rdreport@actionsky.com"
	OPERATIONAL_KEYWORD = "于在更对"
)

// Extract a type of function that extracts information from content
// and set that piece of information into models.MailBrief
// It's used only for parsing jira emails!
type Extract func(mb *models.MailBrief, mailBody []byte) *models.MailBrief

func ExtractOperator() Extract {
	return func(mb *models.MailBrief, mailBody []byte) *models.MailBrief {
		if mb.MailType != Jira {
			return mb
		}
		regexp := regexpMap[OperatorKey]
		match := regexp.FindStringSubmatch(string(mailBody))
		if match == nil {
			fmt.Fprintf(os.Stderr, "we don't find a operator.UID is: %d\n", mb.UID)
			return mb
		}
		operationString := match[2]
		operationIndex := strings.IndexAny(operationString, OPERATIONAL_KEYWORD)
		if operationIndex == -1 {
			fmt.Println("0 is: ", match[0], "1 is: ", match[1], "2 is: ", match[2])
			return mb
		}
		mb.Operator = strings.TrimSpace(operationString[:operationIndex])
		if mb.Operator == "" {
			gitlabRegex := regexpMap[Gitlab]
			gitlabMatch := gitlabRegex.Find([]byte(operationString))
			if gitlabMatch == nil {
				log.Printf("make sure to check out who is the operator!.UID is: %d\n\n", mb.UID)
				return mb
			}
			mb.Operator = string(gitlabMatch[:6])
		}
		return mb
	}
}

func ExtractIssueIDAndProject() Extract {
	return func(mb *models.MailBrief, mailBody []byte) *models.MailBrief {
		if mb.MailType != Jira {
			return mb
		}
		regexp := regexpMap[IssueIDAndProjectKey]
		match := regexp.Find(mailBody)
		if match == nil {
			fmt.Fprintf(os.Stderr, "we didn't find an issueID.UID is: %d\n", mb.UID)
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
		regexp := regexpMap[LinkKey]
		match := regexp.Find(mailBody)
		if match == nil {
			fmt.Fprintf(os.Stderr, "we didn't find an web address.UID is: %d\n", mb.UID)
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
		regexp := regexpMap[AssigneeKey]
		match := regexp.Find(mailBody)
		if match == nil {
			fmt.Fprintf(os.Stderr, "we didn't find an assignee.UID is: %d\n", mb.UID)
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
		regexp := regexpMap[VersionKey]
		match := regexp.FindStringSubmatch(string(mailBody))
		if match == nil {
			fmt.Fprintf(os.Stderr, "we didn't find an version.UID is: %d\n", mb.UID)
			return mb
		}
		fmt.Println("version[0] is", match[0])
		fmt.Println("version[1] is", match[1])
		fmt.Println("version[2] is", match[2])
		fmt.Println("------------------------------------------")
		mb.Version = match[2]
		return mb
	}

}

func ExtractReporter() Extract {
	return func(mb *models.MailBrief, mailBody []byte) *models.MailBrief {
		if mb.MailType != Jira {
			return mb
		}
		regexp := regexpMap[ReporterKey]
		match := regexp.Find(mailBody)
		if match == nil {
			fmt.Fprintf(os.Stderr, "we didn't find an reporter.UID is: %d\n", mb.UID)
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
		regex := regexpMap[IssueTypeKey]
		match := regex.Find(mailBody)
		if match == nil {
			fmt.Fprintf(os.Stderr, "we didn't find an issue type.UID is: %d\n", mb.UID)
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
		regexp := regexpMap[TagKey]
		match := regexp.Find(mailBody)
		if match == nil {
			fmt.Fprintf(os.Stderr, "we didn't find an tag.UID is: %d\n", mb.UID)
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
		regexp := regexpMap[EffectiveBodyKey]
		match := regexp.Find(mailBody)
		if match == nil {
			fmt.Fprintf(os.Stderr, "we didn't find an effective body.UID is: %d\n", mb.UID)
			return mb
		}

		mb.BriefBody = string(trimLineWithGreaterPrefix(match))
		return mb
	}
}

func extractChinese(r *regexp.Regexp, content []byte) string {
	return r.ReplaceAllString(string(content), "")
}

func trimBytesPrefix(bs []byte) []byte {
	return bytes.TrimSpace(bytes.Trim(bs, ">- "))
}

// trimBytesPrefixV2 is used to extract Chinese operator
func trimBytesPrefixV2(bs []byte) []byte {
	return bytes.TrimSpace(bytes.Trim(bs, "]-"))
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
		ExtractEffectiveBody(),
		ExtractVersion())

}

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

func SaveEmails(s db.EmailStore, mbs []*models.MailBrief) error {
	var errs []error
	for _, mb := range mbs {
		if mb.IssueID != "" {
			fmt.Println(mb.IssueID, mb.UID)
			_, err := s.LPush(mb.IssueID, mb.UID)

			if err != nil {
				fmt.Println(err)
				errs = append(errs, err)
			}
			_, err = s.LPush(mb.Time[:10], mb.UID)
			if err != nil {
				fmt.Println(err)
				errs = append(errs, err)
			}
		}

	}
	if len(errs) > 0 {
		fmt.Println(errs)
	}
	return errors.New("err. Something unexpected happend during emails saving")
}

func GetEmailWithDescTimeline(s db.EmailStore, issueID string) ([]string, error) {
	return s.SimpleSort(issueID, "desc")
}

func GetEmailWithTimeline(s db.EmailStore, issueID string) ([]string, error) {
	return s.SimpleSort(issueID, "desc")
}

func PrintEmailWithTimeline(s db.EmailStore, uids []string) error {
	for _, uid := range uids {
		bb, err := s.Get(uid, "BriefBody")
		if err != nil {
			return err
		}
		time, err := s.Get(uid, "Time")
		if err != nil {
			return err
		}
		fmt.Printf("------------------%s---------------------\n\n", time)
		fmt.Printf("%s\n\n", bb)
	}
	return nil
}



func SendEventsLoop(exit, event chan struct{}) {

	for {
		select {
		case <-time.After(time.Second):
			event <- struct{}{}
		case <-exit:
			return
		}
	}
}

func RetriveEmailsLoop(exit, event chan struct{}) {
	for {
		select {
		case <-exit:
			return
		case <-event:

		}
	}
}

func RetrieveEmails(mailboxInfo mail.MailboxInfo, store *db.RedisStore) {
	emails, _ := mail.GetWithKeyMap(mailboxInfo, nil, true, false)
	var wg sync.WaitGroup
	var briefEmails []*models.MailBrief
	wg.Add(len(emails))
	unread := make([]uint32, 0)
	for _, email := range emails {
		go func(mail.Email) {
			briefEmail, err := ParseEmail(email)
			if err != nil {
				fmt.Println(err)
			}
			if briefEmail.MailType == Jira {
				briefEmails = append(briefEmails, briefEmail)
			} else {
				unread = append(unread, email.UID)
			}
			wg.Done()
		}(email)
	}
	wg.Wait()
	_ = SaveEmails(store, briefEmails)
	mail.MarkAsUnread(mailboxInfo, unread)
}
