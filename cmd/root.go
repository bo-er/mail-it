package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/bo-er/mail-it/mail"
	"github.com/spf13/cobra"
)

const (
	dateFormat = "2006-01-02"
)

var projectReg = "DMP-[0-9]+"
var contentReg = `-{3,}`
var mailboxInfo mail.MailboxInfo

var (
	configFile string
	password   string
	to         string
	rootCmd    = &cobra.Command{
		Use:   "mail",
		Short: "A command for starting your email service",
		Long:  `This command is used for Starting your email service.`,
		Run: func(cmd *cobra.Command, args []string) {
			// start := time.Date(2021, 06, 28, 1, 0, 0, 0, time.UTC)
			// result, _ := user.GetLastWeekWork(mailboxInfo, start, []string{projectReg})
			// fmt.Println(result)

			// mails, err := mail.GetUnread(mailboxInfo, false, false)
			// if err != nil {
			// 	log.Panic(err)
			// }
			// // pr := regexp.MustCompile(projectReg)
			// cr := regexp.MustCompile(contentReg)
			// for _, mail := range mails {
			// 	fmt.Println("---------------------------------------------------")
			// 	c, _ := mail.VisibleText()
			// 	content := string(c[0])
			// 	// pv := pr.Find(c[0])
			// 	result := cr.Find(c[0])
			// 	begin := strings.Index(content, string(result))
			// 	end := strings.Index(content, ">")
			// 	fmt.Println(strings.Trim(content[begin:end], "-"))
			// 	fmt.Println("---------------------------------------------------")
			// }

		},
	}
)

func GetPassword() string {
	return password
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "password of a user")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config file path", "c", "", "the path of the config file")
	getEffectiveTimelineCmd.Flags().StringVarP(&issueID, "issueID", "i", "", "The issue's ID, e.g., DMP-7566")
	rootCmd.AddCommand(getLastWeekWorkCmd)
	rootCmd.AddCommand(getEffectiveTimelineCmd)
	sendEmailCmd.Flags().StringVarP(&emailBody, "emailBody", "b", "", "This is the email body")
	sendEmailCmd.PersistentFlags().StringVarP(&to, "sendto", "t", "", "the target email address")
	rootCmd.AddCommand(sendEmailCmd)
	rootCmd.AddCommand(testCmd)

}

func initConfig() {
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(content, &mailboxInfo)
	if err != nil {
		log.Panic(err)
	}
	if mailboxInfo.Pwd == "" {
		if password == "" {
			log.Panic("must provide a password")
		}
		mailboxInfo.Pwd = password
	}
	if mailboxInfo.User == "" {
		log.Panic("must provide a username")
	}
}
