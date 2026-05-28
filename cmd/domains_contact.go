package cmd

import (
	"fmt"

	"github.com/learners-superpumped/aicompany-platform/cli/internal/api"
	"github.com/spf13/cobra"
)

func newDomainsContactCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "contact", Short: "Manage WHOIS contact profiles for domain registration"}
	cmd.AddCommand(
		newDomainsContactCreateCmd(),
		newDomainsContactListCmd(),
		newDomainsContactShowCmd(),
		newDomainsContactUpdateCmd(),
		newDomainsContactDeleteCmd(),
		newDomainsContactSetDefaultCmd(),
	)
	return cmd
}

// contactFlags binds CLI flags into a DomainContact. Used by create and update.
// Empty values mean "no value provided"; update sends them as-is (no merge —
// PATCH on the API rewrites the editable fields as a unit, so the user supplies
// the full intended contact).
func contactFlags(cmd *cobra.Command, c *api.DomainContact) {
	cmd.Flags().StringVar(&c.FirstName, "first-name", "", "registrant first name")
	cmd.Flags().StringVar(&c.LastName, "last-name", "", "registrant last name")
	cmd.Flags().StringVar(&c.Organization, "organization", "", "company name (presence => ContactType=Company)")
	cmd.Flags().StringVar(&c.Email, "email", "", "registrant email")
	cmd.Flags().StringVar(&c.Phone, "phone", "", "E.164 dot-notation (e.g. +82.1012345678)")
	cmd.Flags().StringVar(&c.AddressLine1, "address1", "", "street address line 1")
	cmd.Flags().StringVar(&c.AddressLine2, "address2", "", "street address line 2 (optional)")
	cmd.Flags().StringVar(&c.City, "city", "", "city")
	cmd.Flags().StringVar(&c.State, "state", "", "state/region (required for US/CA)")
	cmd.Flags().StringVar(&c.Zip, "zip", "", "postal code")
	cmd.Flags().StringVar(&c.Country, "country", "", "ISO-3166-1 alpha-2 country code (e.g. KR, US)")
}

func contactRows() ([]string, func(any) []string) {
	return []string{"NAME", "DEFAULT", "TYPE", "EMAIL", "PHONE", "COUNTRY"}, func(v any) []string {
		c := v.(api.DomainContact)
		typ := "Person"
		if c.Organization != "" {
			typ = "Company"
		}
		return []string{c.Name, fmt.Sprintf("%t", c.IsDefault), typ, c.Email, c.Phone, c.Country}
	}
}

func newDomainsContactCreateCmd() *cobra.Command {
	var c api.DomainContact
	var isDefault bool
	cmd := &cobra.Command{
		Use: "create --name=<name> [flags]", Short: "Create a WHOIS contact profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			c.IsDefault = isDefault
			out, err := a.Client.CreateDomainContact(cmd.Context(), a.Team, c)
			if err != nil {
				return err
			}
			cols, row := contactRows()
			return a.Out.Print(*out, cols, row)
		},
	}
	cmd.Flags().StringVar(&c.Name, "name", "", "profile name (unique per team; e.g. default, client-acme) [required]")
	cmd.Flags().BoolVar(&isDefault, "default", false, "mark this profile as the team's default (first profile auto-becomes default)")
	contactFlags(cmd, &c)
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newDomainsContactListCmd() *cobra.Command {
	return &cobra.Command{
		Use: "list", Short: "List WHOIS contact profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			items, err := a.Client.ListDomainContacts(cmd.Context(), a.Team)
			if err != nil {
				return err
			}
			cols, row := contactRows()
			return a.Out.Print(items, cols, row)
		},
	}
}

func newDomainsContactShowCmd() *cobra.Command {
	return &cobra.Command{
		Use: "show <name>", Short: "Show a WHOIS contact profile", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			c, err := a.Client.GetDomainContact(cmd.Context(), a.Team, args[0])
			if err != nil {
				return err
			}
			cols, row := contactRows()
			return a.Out.Print(*c, cols, row)
		},
	}
}

func newDomainsContactUpdateCmd() *cobra.Command {
	var c api.DomainContact
	cmd := &cobra.Command{
		Use: "update <name>", Short: "Update a WHOIS contact profile (provide full set of fields)", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			out, err := a.Client.UpdateDomainContact(cmd.Context(), a.Team, args[0], c)
			if err != nil {
				return err
			}
			cols, row := contactRows()
			return a.Out.Print(*out, cols, row)
		},
	}
	contactFlags(cmd, &c)
	return cmd
}

func newDomainsContactDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use: "delete <name>", Short: "Delete a WHOIS contact profile (refuses if used by any domain)", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			if err := a.Client.DeleteDomainContact(cmd.Context(), a.Team, args[0]); err != nil {
				return err
			}
			fmt.Println("Deleted.")
			return nil
		},
	}
}

func newDomainsContactSetDefaultCmd() *cobra.Command {
	return &cobra.Command{
		Use: "set-default <name>", Short: "Mark a profile as the team's default", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireTeam(); err != nil {
				return err
			}
			if err := a.Client.SetDefaultDomainContact(cmd.Context(), a.Team, args[0]); err != nil {
				return err
			}
			fmt.Printf("%s is now the default.\n", args[0])
			return nil
		},
	}
}
