package cmd

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newStorageCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "storage", Short: "Manage AIC storage buckets and objects"}
	cmd.AddCommand(newStorageBucketsCmd(), newStorageCpCmd(), newStorageLsCmd(),
		newStorageCatCmd(), newStorageRmCmd(), newStoragePresignCmd())
	return cmd
}

func newStorageBucketsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "buckets", Short: "Manage buckets"}
	cmd.AddCommand(
		&cobra.Command{
			Use: "create <name>", Short: "Create a bucket", Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				a, err := appFromCmd(cmd)
				if err != nil {
					return err
				}
				if err := a.RequireProject(); err != nil {
					return err
				}
				if err := a.Client.CreateBucket(cmd.Context(), a.Team, a.Project, args[0]); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "created bucket %s\n", args[0])
				return nil
			},
		},
		&cobra.Command{
			Use: "ls", Short: "List buckets", Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				a, err := appFromCmd(cmd)
				if err != nil {
					return err
				}
				if err := a.RequireProject(); err != nil {
					return err
				}
				bs, err := a.Client.ListBuckets(cmd.Context(), a.Team, a.Project)
				if err != nil {
					return err
				}
				for _, b := range bs {
					fmt.Fprintln(cmd.OutOrStdout(), b.Name)
				}
				return nil
			},
		},
		&cobra.Command{
			Use: "rm <name>", Short: "Delete an empty bucket", Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				a, err := appFromCmd(cmd)
				if err != nil {
					return err
				}
				if err := a.RequireProject(); err != nil {
					return err
				}
				return a.Client.DeleteBucket(cmd.Context(), a.Team, a.Project, args[0])
			},
		},
	)
	return cmd
}

func splitRef(ref string) (bucket, key string, err error) {
	i := strings.IndexByte(ref, '/')
	if i <= 0 || i == len(ref)-1 {
		return "", "", fmt.Errorf("expected <bucket>/<key>, got %q", ref)
	}
	return ref[:i], ref[i+1:], nil
}

func fileExists(p string) bool { _, err := os.Stat(p); return err == nil }

func contentTypeOf(path string) string {
	if ct := mime.TypeByExtension(filepath.Ext(path)); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

func newStorageCpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cp <src> <dst>",
		Short: "Upload or download an object",
		Long:  "Upload:   aic storage cp ./file.txt <bucket>/<key>\nDownload: aic storage cp <bucket>/<key> ./file.txt",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			src, dst := args[0], args[1]
			if fileExists(src) {
				bucket, key, err := splitRef(dst)
				if err != nil {
					return err
				}
				body, err := os.ReadFile(src)
				if err != nil {
					return err
				}
				_, err = a.Client.PutObject(cmd.Context(), a.Team, a.Project, bucket, key, body, contentTypeOf(src))
				return err
			}
			bucket, key, err := splitRef(src)
			if err != nil {
				return err
			}
			body, _, err := a.Client.GetObject(cmd.Context(), a.Team, a.Project, bucket, key)
			if err != nil {
				return err
			}
			return os.WriteFile(dst, body, 0o644)
		},
	}
}

func newStorageLsCmd() *cobra.Command {
	return &cobra.Command{
		Use: "ls <bucket>[/<prefix>]", Short: "List objects", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			bucket, prefix := args[0], ""
			if i := strings.IndexByte(args[0], '/'); i >= 0 {
				bucket, prefix = args[0][:i], args[0][i+1:]
			}
			objs, err := a.Client.ListObjects(cmd.Context(), a.Team, a.Project, bucket, prefix)
			if err != nil {
				return err
			}
			for _, o := range objs {
				fmt.Fprintf(cmd.OutOrStdout(), "%-40s %d\n", o.Key, o.SizeBytes)
			}
			return nil
		},
	}
}

func newStorageCatCmd() *cobra.Command {
	return &cobra.Command{
		Use: "cat <bucket>/<key>", Short: "Print an object to stdout", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			bucket, key, err := splitRef(args[0])
			if err != nil {
				return err
			}
			body, _, err := a.Client.GetObject(cmd.Context(), a.Team, a.Project, bucket, key)
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(body)
			return err
		},
	}
}

func newStorageRmCmd() *cobra.Command {
	return &cobra.Command{
		Use: "rm <bucket>/<key>", Short: "Delete an object", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			bucket, key, err := splitRef(args[0])
			if err != nil {
				return err
			}
			return a.Client.DeleteObject(cmd.Context(), a.Team, a.Project, bucket, key)
		},
	}
}

func newStoragePresignCmd() *cobra.Command {
	var upload bool
	var expires string
	cmd := &cobra.Command{
		Use: "presign <bucket>/<key>", Short: "Create a shareable, time-limited link", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := appFromCmd(cmd)
			if err != nil {
				return err
			}
			if err := a.RequireProject(); err != nil {
				return err
			}
			bucket, key, err := splitRef(args[0])
			if err != nil {
				return err
			}
			op := "download"
			if upload {
				op = "upload"
			}
			signed, err := a.Client.SignURL(cmd.Context(), a.Team, a.Project, bucket, key, op, expires)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), signed)
			return nil
		},
	}
	cmd.Flags().BoolVar(&upload, "upload", false, "create an upload link instead of download")
	cmd.Flags().StringVar(&expires, "expires", "", "link lifetime, e.g. 15m, 1h (default 1h)")
	return cmd
}
