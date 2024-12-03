package browser

import (
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type Config struct {
	BrowserBin    string
	GlobalTimeout time.Duration
	Debug         bool
}

type Browser struct {
	browser *rod.Browser
	lnchr   *launcher.Launcher
	config  *Config
}

func New(config *Config) *Browser {
	lnchr := launcher.New()
	if config.BrowserBin == "auto" || config.BrowserBin == "" {
		if path, exists := launcher.LookPath(); exists {
			lnchr = lnchr.Bin(path)
		}
	} else {
		lnchr = lnchr.Bin(config.BrowserBin)
	}

	browser := rod.New()
	if !config.Debug {
		ctrlURL := lnchr.MustLaunch()
		browser = browser.ControlURL(ctrlURL)
	} else {
		ctrlURL := lnchr.Headless(false).MustLaunch()
		browser = browser.ControlURL(ctrlURL).SlowMotion(time.Duration(1) * time.Second)
	}

	browser = browser.MustConnect().Timeout(config.GlobalTimeout)
	_ = browser.IgnoreCertErrors(true)
	_ = browser.SetCookies(nil)
	return &Browser{browser, lnchr, config}
}

func (b *Browser) MustPage(url ...string) *rod.Page {
	return b.browser.MustPage(url...)
}

func (b *Browser) Visit(url string, wait time.Duration) error {
	return rod.Try(func() {
		page := b.browser.MustPage(url).Timeout(wait).MustWaitLoad()
		defer page.Close()
		time.Sleep(wait)
	})
}

func (b *Browser) Cookies() []*proto.NetworkCookie {
	return b.browser.MustGetCookies()
}

func (b *Browser) SetCookie(c *proto.NetworkCookie) {
	b.browser.MustSetCookies(c)
}

func (b *Browser) GetCookie(name string) (*proto.NetworkCookie, bool) {
	cookieJar := b.browser.MustGetCookies()
	for _, cookie := range cookieJar {
		if cookie.Name == name {
			return cookie, true
		}
	}

	return nil, false
}

func (b *Browser) Close() error {
	berr := b.browser.Close()
	b.lnchr.Cleanup()
	return berr
}
