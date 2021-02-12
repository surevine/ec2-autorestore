package main

import (
	"ec2-autorestore/commands"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
	"log"
)

func main() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable}))
	ec2c := ec2.New(sess)

	commands.InitialiseCommands(ec2c)
	var rootCmd = &cobra.Command{Use: "app"}
	rootCmd.AddCommand(commands.BackupCommand(), commands.RestoreCommand())
	err := rootCmd.Execute()

	if err != nil {
		log.Fatal(err)
	}
}
