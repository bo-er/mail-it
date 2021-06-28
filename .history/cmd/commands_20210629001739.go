package cmd

import (
	"fmt"
	"sync"

	"github.com/bo-er/mail-it/mail"
	"github.com/bo-er/mail-it/user"
	"github.com/bo-er/mail-it/util"
	"github.com/spf13/cobra"
)

var getLastWeekWorkCmd = &cobra.Command{
	Use:   "lw",
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
	Use:   "et",
	Short: "Print the Effective timeline of an issue",
	Long:  `Print the Effective timeline of an issue`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		lastMonday := util.GetFirstDayOfLastWeek()
		lastSaturday := util.GetSaturdayOfLastWeek()
		keyMap := map[string]interface{}{
			"SINCE":    lastMonday.Format(dateFormat),
			"BEFORE":   lastSaturday.Format(dateFormat),
			"DMP-7610": nil,
		}
		fmt.Println("开始获取邮件")
		emails, _ := mail.GetWithKeyMap(mailboxInfo, keyMap, false, false)
		fmt.Println(len(emails))
		var wg sync.WaitGroup
		var briefEmails []user.MailBrief
		wg.Add(len(emails))
		for _, email := range emails {
			e := email
			go func() {
				briefEmail, err := user.ParseEmail(e)
				fmt.Printf("%#v", briefEmail)
				if err != nil {
					fmt.Println(err)
				}
				briefEmails = append(briefEmails, briefEmail)
				wg.Done()
			}()
		}
		wg.Wait()
		fmt.Println(briefEmails[0])
	},
}
