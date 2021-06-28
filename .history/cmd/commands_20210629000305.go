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
		emails, _ := mail.GetWithKeyMap(mailboxInfo, keyMap, false, false)
		var wg sync.WaitGroup
		var briefEmails = make([]user.MailBrief, len(emails))
		wg.Add(len(emails))
		for i, email := range emails {
			go func() {
				briefEmails[i], _ = user.ParseEmail(email)
				wg.Done()
			}()
		}
		wg.Wait()
		fmt.Println("解析成功的邮件数量为:%d",len())
	},
}
