package cmd

import (
	"fmt"
	"log"
	netMail "net/mail"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/bo-er/mail-it/mail"
	"github.com/bo-er/mail-it/models"
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
		store := user.GetRedisStore()
		if store == nil {
			fmt.Fprintf(os.Stderr, "failed to get db")
		}
		lastMonday := util.GetFirstDayOfLastWeek()
		lastSaturday := util.GetSaturdayOfLastWeek()
		// keyMap := map[string]interface{}{
		// 	"SINCE":  lastMonday.Format(dateFormat),
		// 	"BEFORE": lastSaturday.Format(dateFormat),
		// }

		emails, _ := mail.GetWithKeyMap(mailboxInfo, nil, false, false)
		fmt.Printf("收到了%d封新邮件\n", len(emails))
		var wg sync.WaitGroup
		var briefEmails []*models.MailBrief
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
		_ = user.SaveEmails(store, briefEmails)
		for _, mb := range briefEmails {
			var mailID = strconv.FormatUint(uint64(mb.UID), 10)
			err := store.Set(mailID, *mb)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
			}
			if _, exists := lastWeekWorks[mb.IssueID]; exists {
				continue
			}
			if mb.FilterBriefMail(func(m *models.MailBrief) bool {
				// 判断是自己的任务
				if m.Assignee != mailboxInfo.Username {
					return false
				}
				local, _ := time.LoadLocation("Local")
				receiveTime, _ := time.ParseInLocation(user.MailBriefTimeFormat, m.Time, local)
				return receiveTime.After(lastMonday) && receiveTime.Before(lastSaturday)
			}) {

				lastWeekWorks[mb.IssueID] = mb.Link
				result, _ := store.Get(mb.IssueID, "Reporter")
				fmt.Printf("%#v\n", result...)

			}

		}
		// results, err := store.LGetAll("DMP-7566")
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, err.Error())
		// }
		// fmt.Println(results)
		counter := 1
		for _, link := range lastWeekWorks {
			fmt.Printf("%d.%s\n", counter, link)
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
			"SINCE": "2021-06-30",
			// "BEFORE": lastSaturday.Format(dateFormat),
		}
		fmt.Printf("SINCE: %s, BEFORE: %s\n", lastMonday.Format(dateFormat), lastSaturday.Format(dateFormat))
		emails, _ := mail.GetWithKeyMap(mailboxInfo, keyMap, false, false)
		fmt.Printf("总共的邮件数量:%d\n", len(emails))

		// store := user.GetRedisStore()
		// if store == nil {
		// 	fmt.Fprintf(os.Stderr, "failed to get db")
		// }
		// resuls, err := store.Get("472", "Assignee", "Project")
		// if err != nil {
		// 	fmt.Println(err)
		// }
		// fmt.Println(resuls...)



		// store := user.GetRedisStore()
		// if store == nil {
		// 	fmt.Fprintf(os.Stderr, "failed to get db")
		// }
		// results, _ := user.GetEmailWithDescTimeline(store, "DMP-7566")
		// user.PrintEmailWithTimeline(store, results)
	},
}
