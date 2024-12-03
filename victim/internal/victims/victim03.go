package victims

import (
	"errors"
	"strings"
	"time"

	"github.com/cyllective/oauth-labs/victim/internal/browser"
)

type Victim03 struct {
	config VictimConfig
}

func (v Victim03) Number() int {
	return v.config.LabNumber
}

func (v Victim03) Name() string {
	return v.config.LabName
}

func (v Victim03) Config() VictimConfig {
	return v.config
}

func (v Victim03) CheckURL(u string) error {
	if !strings.HasPrefix(u, v.config.ServerURL) && !strings.HasPrefix(u, v.config.ClientURL) {
		return errors.New("victim03 will only visit client-03.oauth.labs or server-03.oauth.labs URLs")
	}
	return nil
}

func (v Victim03) Handle(browser *browser.Browser, u string, wait time.Duration) error {
	if err := login(browser, &v.config); err != nil {
		return err
	}
	return browser.Visit(u, wait)
}
