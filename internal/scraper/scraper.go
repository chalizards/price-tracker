package scraper

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

func FetchHTML(ctx context.Context, url string) (string, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", true),
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	taskCtx, cancel = context.WithTimeout(taskCtx, 30*time.Second)
	defer cancel()

	var html string
	err := chromedp.Run(taskCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(`Object.defineProperty(navigator, 'webdriver', {get: () => undefined})`, nil).Do(ctx)
		}),
		chromedp.Navigate(url),
		chromedp.Sleep(5*time.Second),
		chromedp.OuterHTML("body", &html),
	)
	if err != nil {
		return "", err
	}

	return html, nil
}
