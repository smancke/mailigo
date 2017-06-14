package main

import (
	"flag"
	. "github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestConfig_ReadConfigDefaults(t *testing.T) {
	originalArgs := os.Args
	os.Args = []string{"goletter"}
	defer func() { os.Args = originalArgs }()

	defaultConfig := DefaultConfig()
	gotConfig := ReadConfig()
	Equal(t, defaultConfig, gotConfig)
}

func TestConfig_ReadConfig(t *testing.T) {
	input := []string{
		"--host=host",
		"--port=port",
		"--log-level=loglevel",
		"--text-logging=true",
		"--jwt-secret=jwtsecret",
		"--db-driver=dbdriver",
		"--db-datasource=dbdatasource",
		"--mail-host=mailhost",
		"--mail-port=mailport",
		"--mail-username=mailusername",
		"--mail-password=mailpassword",
		"--mail-ssl=true",
		"--grace-period=4s",
	}

	expected := &Config{
		Host:         "host",
		Port:         "port",
		LogLevel:     "loglevel",
		TextLogging:  true,
		JwtSecret:    "jwtsecret",
		DBDriver:     "dbdriver",
		DBDataSource: "dbdatasource",
		MailHost:     "mailhost",
		MailPort:     "mailport",
		MailUsername: "mailusername",
		MailPassword: "mailpassword",
		MailSSL:      true,
		GracePeriod:  4 * time.Second,
	}

	cfg, err := readConfig(flag.NewFlagSet("", flag.ContinueOnError), input)
	NoError(t, err)
	Equal(t, expected, cfg)
}

func TestConfig_ReadConfigFromEnv(t *testing.T) {
	NoError(t, os.Setenv("GOLETTER_HOST", "host"))
	NoError(t, os.Setenv("GOLETTER_PORT", "port"))
	NoError(t, os.Setenv("GOLETTER_LOG_LEVEL", "loglevel"))
	NoError(t, os.Setenv("GOLETTER_TEXT_LOGGING", "true"))
	NoError(t, os.Setenv("GOLETTER_JWT_SECRET", "jwtsecret"))
	NoError(t, os.Setenv("GOLETTER_DB_DRIVER", "dbdriver"))
	NoError(t, os.Setenv("GOLETTER_DB_DATASOURCE", "dbdatasource"))
	NoError(t, os.Setenv("GOLETTER_MAIL_HOST", "mailhost"))
	NoError(t, os.Setenv("GOLETTER_MAIL_PORT", "mailport"))
	NoError(t, os.Setenv("GOLETTER_MAIL_USERNAME", "mailusername"))
	NoError(t, os.Setenv("GOLETTER_MAIL_PASSWORD", "mailpassword"))
	NoError(t, os.Setenv("GOLETTER_MAIL_SSL", "true"))
	NoError(t, os.Setenv("GOLETTER_GRACE_PERIOD", "4s"))

	defer func() {
		os.Unsetenv("GOLETTER_HOST")
		os.Unsetenv("GOLETTER_PORT")
		os.Unsetenv("GOLETTER_LOG_LEVEL")
		os.Unsetenv("GOLETTER_TEXT_LOGGING")
		os.Unsetenv("GOLETTER_JWT_SECRET")
		os.Unsetenv("GOLETTER_DB_DRIVER")
		os.Unsetenv("GOLETTER_DB_DATASOURCE")
		os.Unsetenv("GOLETTER_MAIL_HOST")
		os.Unsetenv("GOLETTER_MAIL_PORT")
		os.Unsetenv("GOLETTER_MAIL_USERNAME")
		os.Unsetenv("GOLETTER_MAIL_PASSWORD")
		os.Unsetenv("GOLETTER_MAIL_SSL")
		os.Unsetenv("GOLETTER_GRACE_PERIOD")
	}()

	expected := &Config{
		Host:         "host",
		Port:         "port",
		LogLevel:     "loglevel",
		TextLogging:  true,
		JwtSecret:    "jwtsecret",
		DBDriver:     "dbdriver",
		DBDataSource: "dbdatasource",
		MailHost:     "mailhost",
		MailPort:     "mailport",
		MailUsername: "mailusername",
		MailPassword: "mailpassword",
		MailSSL:      true,
		GracePeriod:  4 * time.Second,
	}

	cfg, err := readConfig(flag.NewFlagSet("", flag.ContinueOnError), []string{})
	NoError(t, err)
	Equal(t, expected, cfg)
}
