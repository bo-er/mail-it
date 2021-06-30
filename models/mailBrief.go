package models

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
	Time      string
	//MailType 邮件的类型 "JIRA"表示来自jira, "Slack"表示来自slack的消息, "Confluence"表示来自confluence上的文档变更,"GitLab"表示来自gitlab
	MailType string
	UID      uint32
}

func (mb *MailBrief) MapFormat() (m map[string]interface{}) {
	m = map[string]interface{}{
		"Operator":  mb.Operator,
		"IssueID":   mb.IssueID,
		"Link":      mb.Link,
		"Project":   mb.Project,
		"IssueType": mb.IssueType,
		"Assignee":  mb.Assignee,
		"Reporter":  mb.Reporter,
		"Tag":       mb.Tag,
		"Version":   mb.Version,
		"BriefBody": mb.BriefBody,
		"Time":      mb.Time,
		"MailType":  mb.MailType,
		"UID":       mb.UID,
	}
	return m
}

type BriefMailFilter func(mb *MailBrief) bool

func (mb *MailBrief) FilterBriefMail(bf BriefMailFilter) bool {
	return bf(mb)
}
