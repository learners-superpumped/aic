package cmd

import (
	"fmt"
	"os"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/spf13/cobra"
)

func newMessagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "messages",
		Aliases: []string{"msg", "mail"},
		Short:   "Send and read inbox messages",
	}
	cmd.AddCommand(
		newMessagesSendCmd(),
		newMessagesListCmd(),
		newMessagesShowCmd(),
	)
	return cmd
}

func newMessagesSendCmd() *cobra.Command {
	var inbox, to, subject, body, bodyFile string
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a message from an inbox",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			if inbox == "" || to == "" {
				return fmt.Errorf("--inbox and --to are required")
			}
			if bodyFile != "" {
				data, err := os.ReadFile(bodyFile)
				if err != nil {
					return err
				}
				body = string(data)
			}
			m, err := a.Client.SendMessage(cmd.Context(), a.Project, inbox, to, subject, body)
			if err != nil {
				return err
			}
			return a.Out.Print(*m, []string{"MESSAGE ID"}, func(v any) []string {
				x := v.(api.Message)
				return []string{x.MessageID}
			})
		},
	}
	cmd.Flags().StringVar(&inbox, "inbox", "", "sending inbox address")
	cmd.Flags().StringVar(&to, "to", "", "recipient address")
	cmd.Flags().StringVar(&subject, "subject", "", "subject line")
	cmd.Flags().StringVar(&body, "body", "", "message body")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "read message body from a file")
	return cmd
}

func newMessagesListCmd() *cobra.Command {
	var inbox string
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List messages in an inbox",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			if inbox == "" {
				return fmt.Errorf("--inbox is required")
			}
			items, err := a.Client.ListMessages(cmd.Context(), a.Project, inbox)
			if err != nil {
				return err
			}
			return a.Out.Print(items, []string{"ID", "FROM", "SUBJECT", "RECEIVED"}, func(v any) []string {
				m := v.(api.Message)
				return []string{m.MessageID, m.From, m.Subject, m.ReceivedAt.Format("2006-01-02 15:04")}
			})
		},
	}
	cmd.Flags().StringVar(&inbox, "inbox", "", "inbox address")
	return cmd
}

func newMessagesShowCmd() *cobra.Command {
	var inbox string
	cmd := &cobra.Command{
		Use:     "show <id>",
		Aliases: []string{"read"},
		Short:   "Show a message",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			if inbox == "" {
				return fmt.Errorf("--inbox is required")
			}
			m, err := a.Client.GetMessage(cmd.Context(), a.Project, inbox, args[0])
			if err != nil {
				return err
			}
			return a.Out.Print(*m, []string{"ID", "FROM", "TO", "SUBJECT"}, func(v any) []string {
				x := v.(api.Message)
				return []string{x.MessageID, x.From, x.To, x.Subject}
			})
		},
	}
	cmd.Flags().StringVar(&inbox, "inbox", "", "inbox address")
	return cmd
}
