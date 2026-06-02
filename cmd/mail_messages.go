package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/learners-superpumped/aic/internal/api"
	"github.com/spf13/cobra"
)

func newMailMessagesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "messages", Short: "List and read stored mail messages"}
	cmd.AddCommand(newMailMessagesListCmd(), newMailMessagesShowCmd(), newMailMessagesAttachmentCmd())
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
		Short: "Show a message's metadata and body; optionally write its raw .eml to a file",
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
				if err := os.WriteFile(rawOut, raw, 0o600); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "wrote %d bytes to %s\n", len(raw), rawOut)
			}
			// json/yaml: emit the full detail (recipients, attachments, raw_base64).
			if a.Out.Format() != "table" {
				return a.Out.Print(*d, nil, nil)
			}
			// table: envelope row, then recipients + attachments sections.
			if err := a.Out.Print(d.MailMessage, []string{"ID", "DIRECTION", "FROM", "SUBJECT", "STATUS", "SENT_AT"}, func(v any) []string {
				x := v.(api.MailMessage)
				return []string{x.ID, x.Direction, x.From, x.Subject, x.Status, x.SentAt}
			}); err != nil {
				return err
			}
			w := a.Out.Writer()
			if d.TextBody != "" {
				fmt.Fprintln(w, "\nBody:")
				fmt.Fprintln(w, d.TextBody)
			}
			if len(d.Recipients) > 0 {
				fmt.Fprintln(w, "\nRecipients:")
				for _, r := range d.Recipients {
					fmt.Fprintf(w, "  %-3s %s (%s)\n", r.Kind, r.Address, r.Status)
				}
			}
			if len(d.Attachments) > 0 {
				fmt.Fprintln(w, "\nAttachments:")
				for _, at := range d.Attachments {
					fmt.Fprintf(w, "  %s  %s  %s  %d bytes\n", at.ID, at.Filename, at.ContentType, at.SizeBytes)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&rawOut, "raw-out", "", "write the raw .eml to this path")
	return cmd
}

func newMailMessagesAttachmentCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "attachment <message-id> <attachment-id>",
		Short: "Download a message attachment (id from `messages show`)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			data, filename, err := a.Client.GetMailAttachment(cmd.Context(), a.Team, a.Project, args[0], args[1])
			if err != nil {
				return err
			}
			if out == "-" { // stream to stdout
				_, err := cmd.OutOrStdout().Write(data)
				return err
			}
			dest := out
			if dest == "" { // default to the server-suggested filename
				base, err := safeAttachmentName(filename)
				if err != nil {
					return err
				}
				dest = base
			}
			if err := os.WriteFile(dest, data, 0o600); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "wrote %d bytes to %s\n", len(data), dest)
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "write to this path (default: attachment filename; '-' for stdout)")
	return cmd
}

// safeAttachmentName reduces a server-suggested attachment filename to a safe
// basename. The name ultimately comes from the (untrusted) email sender, so a
// crafted value like "../../x" must not let the download escape the cwd. Returns
// an error directing the user to pass --out when the name can't be made safe.
func safeAttachmentName(filename string) (string, error) {
	base := filepath.Base(filename)
	if base == "." || base == ".." || base == string(filepath.Separator) ||
		base == "" || strings.ContainsAny(base, `/\`) {
		return "", fmt.Errorf("server-suggested filename %q is unsafe; pass --out <path>", filename)
	}
	return base, nil
}
