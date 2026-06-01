package cmd

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/spf13/cobra"
)

func newMailMessagesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "messages", Short: "List and read stored mail messages"}
	cmd.AddCommand(newMailMessagesListCmd(), newMailMessagesShowCmd())
	return cmd
}

func newMailMessagesListCmd() *cobra.Command {
	var direction, inbox string
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List stored messages (newest first)",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			list, err := a.Client.ListMailMessages(cmd.Context(), a.Team, a.Project, direction, inbox, limit)
			if err != nil {
				return err
			}
			return a.Out.Print(list, []string{"ID", "DIRECTION", "FROM", "SUBJECT", "STATUS", "CREATED"}, func(v any) []string {
				x := v.(api.MailMessage)
				return []string{x.ID, x.Direction, x.From, x.Subject, x.Status, x.CreatedAt}
			})
		},
	}
	cmd.Flags().StringVar(&direction, "direction", "", "filter: sent|received")
	cmd.Flags().StringVar(&inbox, "inbox", "", "filter by inbox address or id")
	cmd.Flags().IntVar(&limit, "limit", 0, "max rows (default 50)")
	return cmd
}

func newMailMessagesShowCmd() *cobra.Command {
	var rawOut string
	cmd := &cobra.Command{
		Use:   "show <message-id>",
		Short: "Show a message's metadata; optionally write its raw .eml to a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			d, err := a.Client.ShowMailMessage(cmd.Context(), a.Team, a.Project, args[0])
			if err != nil {
				return err
			}
			if rawOut != "" {
				raw, err := base64.StdEncoding.DecodeString(d.RawBase64)
				if err != nil {
					return err
				}
				if err := os.WriteFile(rawOut, raw, 0o644); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "wrote %d bytes to %s\n", len(raw), rawOut)
			}
			return a.Out.Print(d.MailMessage, []string{"ID", "DIRECTION", "FROM", "SUBJECT", "STATUS", "SENT_AT"}, func(v any) []string {
				x := v.(api.MailMessage)
				return []string{x.ID, x.Direction, x.From, x.Subject, x.Status, x.SentAt}
			})
		},
	}
	cmd.Flags().StringVar(&rawOut, "raw-out", "", "write the raw .eml to this path")
	return cmd
}
