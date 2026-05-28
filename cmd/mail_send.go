package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/spf13/cobra"
)

func newMailSendCmd() *cobra.Command {
	var (
		from, subject, text, textFile, html, htmlFile string
		to, cc, bcc, replyTo, attach                  []string
	)
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send mail via SES",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			if from == "" || len(to) == 0 {
				return fmt.Errorf("--from and at least one --to are required")
			}
			if textFile != "" {
				b, err := os.ReadFile(textFile)
				if err != nil {
					return err
				}
				text = string(b)
			}
			if htmlFile != "" {
				b, err := os.ReadFile(htmlFile)
				if err != nil {
					return err
				}
				html = string(b)
			}
			if text == "" && html == "" {
				return fmt.Errorf("at least one of --text/--text-file/--html/--html-file is required")
			}
			atts := []api.MailAttachment{}
			for _, path := range attach {
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				atts = append(atts, api.MailAttachment{
					Filename:   filepath.Base(path),
					DataBase64: base64.StdEncoding.EncodeToString(data),
				})
			}
			res, err := a.Client.SendMail(cmd.Context(), a.Team, a.Project, api.SendMessageRequest{
				From: from, To: to, CC: cc, BCC: bcc, ReplyTo: replyTo,
				Subject: subject, Text: text, HTML: html, Attachments: atts,
			})
			if err != nil {
				return err
			}
			return a.Out.Print(*res, []string{"MESSAGE_ID", "FROM", "SENT_AT"}, func(v any) []string {
				x := v.(api.SendMessageResponse)
				return []string{x.SESMessageID, x.From, x.SentAt.Format("2006-01-02 15:04:05")}
			})
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "sending address (must be an inbox you've created)")
	cmd.Flags().StringSliceVar(&to, "to", nil, "recipient address (repeatable)")
	cmd.Flags().StringSliceVar(&cc, "cc", nil, "CC recipient (repeatable)")
	cmd.Flags().StringSliceVar(&bcc, "bcc", nil, "BCC recipient (repeatable)")
	cmd.Flags().StringSliceVar(&replyTo, "reply-to", nil, "Reply-To address (repeatable)")
	cmd.Flags().StringVar(&subject, "subject", "", "subject line")
	cmd.Flags().StringVar(&text, "text", "", "plain-text body")
	cmd.Flags().StringVar(&textFile, "text-file", "", "read plain-text body from file")
	cmd.Flags().StringVar(&html, "html", "", "HTML body")
	cmd.Flags().StringVar(&htmlFile, "html-file", "", "read HTML body from file")
	cmd.Flags().StringSliceVar(&attach, "attach", nil, "attachment file path (repeatable)")
	return cmd
}
