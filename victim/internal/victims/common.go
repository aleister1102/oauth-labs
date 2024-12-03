package victims

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-rod/rod"

	"github.com/cyllective/oauth-labs/victim/internal/browser"
)

// Login performs authentication for the given victim
func login(browser *browser.Browser, vc *VictimConfig) error {
	return rod.Try(func() {
		{
			log.Printf("[victim%s]: visiting: %s\n", vc.LabName, vc.ServerURL)
			loginPage := browser.MustPage(vc.ServerURL + "/login")
			loginPage.Timeout(time.Duration(3) * time.Second)
			loginPage.MustWaitLoad()

			// Login with username and password
			log.Printf("[victim%s]: entering username and password...", vc.LabName)
			loginPage.MustElement(`#login-form input[name="username"]`).MustInput(vc.Username)
			loginPage.MustElement(`#login-form input[name="password"]`).MustInput(vc.Password)
			loginPage.MustElement(`#login-form button`).MustClick()
			loginPage.MustWaitNavigation()
			loginPage.MustWaitIdle()

			// Ensure we have successfully authenticated, we expect a session cookie to be present.
			_, ok := browser.GetCookie(fmt.Sprintf("server-%s", vc.LabName))
			if !ok {
				log.Printf("[victim%s]: failed to extract server session cookie", vc.LabName)
				panic("failed to extract server session cookie")
			}
		}

		{
			log.Printf("[victim%s]: visiting: %s", vc.LabName, vc.ClientURL)
			loginPage := browser.MustPage(vc.ClientURL + "/login")
			loginPage.Timeout(time.Duration(3) * time.Second)
			loginPage.MustWaitNavigation()
			loginPage.MustWaitIdle()

			// Click on "Login with OAuth"
			log.Printf("[victim%s]: performing oauth flow...", vc.LabName)
			loginPage.MustElement(`a.ui.huge.primary.button`).MustClick()
			loginPage.MustWaitNavigation()
			loginPage.MustWaitIdle()

			info := loginPage.MustInfo()
			if strings.HasPrefix(info.URL, vc.ServerURL) {
				// Confirm authorization prompt by clicking "authorize"
				log.Printf("[victim%s]: confirming authorization prompt...", vc.LabName)
				loginPage.MustElement(`#authorize-btn`).MustClick()
				loginPage.MustWaitNavigation()
				loginPage.MustWaitIdle()
			}

			// Ensure we have successfully authenticated, we expect a session cookie to be present.
			_, ok := browser.GetCookie(fmt.Sprintf("client-%s", vc.LabName))
			if !ok {
				log.Printf("[victim%s]: failed to extract client session cookie", vc.LabName)
				panic("failed to extract client session cookie")
			}
		}
	})
}
