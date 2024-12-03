package victims

import (
	"errors"
	"strings"
	"time"

	"github.com/cyllective/oauth-labs/victim/internal/browser"
)

type Victim02 struct {
	config VictimConfig
}

func (v Victim02) Number() int {
	return v.config.LabNumber
}

func (v Victim02) Name() string {
	return v.config.LabName
}

func (v Victim02) Config() VictimConfig {
	return v.config
}

func (v Victim02) CheckURL(u string) error {
	if !strings.HasPrefix(u, v.config.ServerURL) && !strings.HasPrefix(u, v.config.ClientURL) {
		return errors.New("victim02 will only visit client-02.oauth.labs or server-02.oauth.labs URLs")
	}
	return nil
}

func (v Victim02) Handle(browser *browser.Browser, url string, wait time.Duration) error {
	if err := login(browser, &v.config); err != nil {
		return err
	}
	return browser.Visit(url, wait)
}
