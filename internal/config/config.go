// Package config reads and writes the ~/.aic credentials and config files.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/ini.v1"
)

// Profile is one named set of credentials + config.
type Profile struct {
	Name           string
	AccessToken    string
	RefreshToken   string
	ExpiresAt      time.Time
	DefaultProject string
	Output         string
	APIEndpoint    string
}

const timeFormat = time.RFC3339

func dir() (string, error) {
	if d := os.Getenv("AIC_CONFIG_DIR"); d != "" {
		return d, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".aic"), nil
}

func paths() (credPath, cfgPath string, err error) {
	d, err := dir()
	if err != nil {
		return "", "", err
	}
	return filepath.Join(d, "credentials"), filepath.Join(d, "config"), nil
}

// Save writes the profile to both files, creating the dir if needed.
func Save(p *Profile) error {
	d, err := dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, 0o700); err != nil {
		return err
	}
	credPath, cfgPath, _ := paths()

	cred := loadOrNew(credPath)
	sec := cred.Section(p.Name)
	sec.Key("access_token").SetValue(p.AccessToken)
	sec.Key("refresh_token").SetValue(p.RefreshToken)
	if !p.ExpiresAt.IsZero() {
		sec.Key("expires_at").SetValue(p.ExpiresAt.UTC().Format(timeFormat))
	}
	if err := cred.SaveTo(credPath); err != nil {
		return err
	}
	if err := os.Chmod(credPath, 0o600); err != nil {
		return err
	}

	cfg := loadOrNew(cfgPath)
	csec := cfg.Section(p.Name)
	csec.Key("default_project").SetValue(p.DefaultProject)
	csec.Key("output").SetValue(p.Output)
	csec.Key("api_endpoint").SetValue(p.APIEndpoint)
	return cfg.SaveTo(cfgPath)
}

// Load reads a named profile. Returns an error if the profile is absent.
func Load(name string) (*Profile, error) {
	credPath, cfgPath, err := paths()
	if err != nil {
		return nil, err
	}
	cred, err := ini.Load(credPath)
	if err != nil {
		return nil, fmt.Errorf("no credentials found (run `aic login`): %w", err)
	}
	if _, err := cred.GetSection(name); err != nil {
		return nil, fmt.Errorf("profile %q not found (run `aic login`)", name)
	}
	sec := cred.Section(name)

	p := &Profile{
		Name:         name,
		AccessToken:  sec.Key("access_token").String(),
		RefreshToken: sec.Key("refresh_token").String(),
	}
	if v := sec.Key("expires_at").String(); v != "" {
		if t, e := time.Parse(timeFormat, v); e == nil {
			p.ExpiresAt = t
		}
	}
	if cfg, e := ini.Load(cfgPath); e == nil {
		c := cfg.Section(name)
		p.DefaultProject = c.Key("default_project").String()
		p.Output = c.Key("output").String()
		p.APIEndpoint = c.Key("api_endpoint").String()
	}
	return p, nil
}

// Delete removes a profile's section from both files.
func Delete(name string) error {
	credPath, cfgPath, err := paths()
	if err != nil {
		return err
	}
	if f, err := ini.Load(credPath); err == nil {
		f.DeleteSection(name)
		if err := f.SaveTo(credPath); err != nil {
			return err
		}
		_ = os.Chmod(credPath, 0o600)
	}
	if f, err := ini.Load(cfgPath); err == nil {
		f.DeleteSection(name)
		if err := f.SaveTo(cfgPath); err != nil {
			return err
		}
	}
	return nil
}

func loadOrNew(path string) *ini.File {
	if f, err := ini.Load(path); err == nil {
		return f
	}
	return ini.Empty()
}
