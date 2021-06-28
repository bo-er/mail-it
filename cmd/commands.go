package cmd

import (
	"bytes"
	"fmt"

	"github.com/bo-er/mail-it/user"
	"github.com/bo-er/mail-it/util"
	"github.com/spf13/cobra"
)

var getLastWeekWorkCmd = &cobra.Command{
	Use:   "lw",
	Short: "Gets your last week's work on jira",
	RunE: func(cmd *cobra.Command, args []string) error {
		initConfig()
		lastMonday := util.GetFirstDayOfLastWeek()
		lastSaturday := util.GetSaturdayOfLastWeek()
		keyMap := map[string]interface{}{
			"SINCE":  lastMonday.Format(dateFormat),
			"BEFORE": lastSaturday.Format(dateFormat),
		}
		projects, _ := user.GetLastWeekWork(mailboxInfo, func(body [][]byte) bool {
			fmt.Println(len(body))
			return bytes.ContainsAny(body[0],"经办人: "+mailboxInfo.Username)

		}, keyMap, []string{projectReg})
		fmt.Println(projects)
		return nil
	},
}
