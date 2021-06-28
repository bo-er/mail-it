package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"

	"github.com/bo-er/mail-it/mail"
	"github.com/bo-er/mail-it/util"
	"github.com/spf13/cobra"
)

const (
	dateFormat = "2006-01-02"
)

var projectReg = "DMP-[0-9]+"
var contentReg = "-----------------.*>"

var (
	configFile string
	username   string
	password   string
	to         string
	rootCmd    = &cobra.Command{
		Use:   "mail",
		Short: "A command for starting your email service",
		Long:  `This command is used for Starting your email service.`,
		Run: func(cmd *cobra.Command, args []string) {
			content, err := ioutil.ReadFile(configFile)
			if err != nil {
				log.Panic(err)
			}
			var mailboxInfo mail.MailboxInfo
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
				if username == "" {
					log.Panic("must provide a username")
				}
				mailboxInfo.User = username
			}
			fmt.Println(util.GetFirstDateOfWeek())
			mails, err := mail.GetUnread(mailboxInfo, false, false)
			if err != nil {
				log.Panic(err)
			}
			pr := regexp.MustCompile(projectReg)
			cr := regexp.MustCompile(contentReg)
			for _, mail := range mails {
				fmt.Println("---------------------------------------------------")
				content, _ := mail.VisibleText()
				pv := pr.Find(content[0])
				fmt.Println(string(pv))
				cr.Find(content[0])
				fmt.Println
				// fmt.Println(string(content[0]))
				fmt.Println("---------------------------------------------------")
			}

			// from := netMail.Address{"", username}
			// sendto := netMail.Address{"", to}
			// message := mail.Setup(from.Address, sendto.Address)
			// client, err := mail.Connect(username, password)
			// if err != nil {
			// 	log.Panic(err)
			// }
			// err = mail.Send(from.Address, sendto.Address, client, []byte(message))
			// if err != nil {
			// 	log.Panic(err)
			// }
		},
	}
)

func GetUsername() string {
	return username
}

func GetPassword() string {
	return password
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "name of mailbox owner")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "password of a user")
	rootCmd.PersistentFlags().StringVarP(&to, "sendto", "t", "", "the target email address")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config file path", "c", "", "the path of the config file")

}

func initConfig() {

}
