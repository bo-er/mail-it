package cmd

import (
	"fmt"
	"log"
	netMail "net/mail"
	"sync"
	"time"

	"github.com/bo-er/mail-it/mail"
	"github.com/bo-er/mail-it/user"
	"github.com/bo-er/mail-it/util"
	"github.com/spf13/cobra"
)

var getLastWeekWorkCmd = &cobra.Command{
	Use:   "lastweek",
	Short: "Gets your last week's work on jira",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		lastMonday := util.GetFirstDayOfLastWeek()
		lastSaturday := util.GetSaturdayOfLastWeek()
		keyMap := map[string]interface{}{
			"SINCE":    lastMonday.Format(dateFormat),
			"BEFORE":   lastSaturday.Format(dateFormat),
			"DMP-7610": nil,
		}
		projects, _ := user.GetLastWeekWork(mailboxInfo, func(body [][]byte) bool {
			return user.FilterByNameOfAssignee(body, mailboxInfo.Username)

		}, keyMap, []string{projectReg})
		fmt.Println(projects)

	},
}

var issueID string
var getEffectiveTimelineCmd = &cobra.Command{
	Use:   "etimeline",
	Short: "Print the Effective timeline of an issue",
	Long:  `Print the Effective timeline of an issue`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		lastMonday := util.GetFirstDayOfLastWeek()
		lastSaturday := util.GetSaturdayOfLastWeek()
		// keyMap := map[string]interface{}{
		// 	"SINCE":  lastMonday.Format(dateFormat),
		// 	"BEFORE": lastSaturday.Format(dateFormat),
		// }

		emails, _ := mail.GetWithKeyMap(mailboxInfo, nil, false, false)
		var wg sync.WaitGroup
		var briefEmails []*user.MailBrief
		wg.Add(len(emails))
		for _, email := range emails {
			e := email
			go func() {
				briefEmail, err := user.ParseEmail(e)
				if err != nil {
					fmt.Println(err)
				}
				if briefEmail.MailType == user.Jira {
					briefEmails = append(briefEmails, briefEmail)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		lastWeekWorks := map[string]string{}
		for _, bm := range briefEmails {
			if _, exists := lastWeekWorks[bm.IssueID]; exists {
				continue
			}
			if bm.FilterBriefMail(func(m *user.MailBrief) bool {
				// 判断是自己的任务
				if m.Assignee != mailboxInfo.Username {
					return false
				}
				local, _ := time.LoadLocation("Local")
				receiveTime, _ := time.ParseInLocation(user.MailBriefTimeFormat, m.Time, local)
				return receiveTime.After(lastMonday) && receiveTime.Before(lastSaturday)
			}) {
				lastWeekWorks[bm.IssueID] = bm.Link
			}

		}
		counter := 1
		for _, link := range lastWeekWorks {
			fmt.Printf("%d.%s\n", counter, link)
			fmt.Printf("%#v\n", lastWeekWorks)
			counter++
		}
	},
}

var emailBody string

var sendEmailCmd = &cobra.Command{
	Use:   "send",
	Short: "Send your email to someone",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		from := netMail.Address{Name: "", Address: mailboxInfo.User}
		sendto := netMail.Address{Name: "", Address: to}
		message := mail.Setup(from.Address, sendto.Address, mailboxInfo.Username)
		message += emailBody
		client, err := mail.Connect(mailboxInfo.User, password)
		if err != nil {
			log.Panic(err)
		}
		err = mail.Send(from.Address, sendto.Address, client, []byte(message))
		if err != nil {
			log.Panic(err)
		}
	},
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "This is the command used for testing",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		lastMonday := util.GetFirstDayOfLastWeek()
		lastSaturday := util.GetSaturdayOfLastWeek()
		keyMap := map[string]interface{}{
			// "SINCE":  lastMonday.Format(dateFormat),
			"BEFORE": lastSaturday.Format(dateFormat),
		}
		fmt.Printf("SINCE: %s, BEFORE: %s\n", lastMonday.Format(dateFormat), lastSaturday.Format(dateFormat))
		emails, _ := mail.GetWithKeyMap(mailboxInfo, keyMap, false, false)
		fmt.Printf("总共的邮件数量:%d\n", len(emails))

	},
}
