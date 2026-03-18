package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

// FetchHTML navigates to a URL and returns the rendered HTML.
// If CHROME_WS_URL is set, connects to an existing Chrome instance via DevTools Protocol.
// Otherwise, launches a local headless Chrome.
func FetchHTML(ctx context.Context, targetURL string) (string, error) {
	wsURL := os.Getenv("CHROME_WS_URL")
	if wsURL != "" {
		return fetchViaRemoteChrome(ctx, wsURL, targetURL)
	}
	return fetchViaHeadless(ctx, targetURL)
}

// discoverWSURL fetches the browser WebSocket URL from Chrome DevTools,
// using Host: localhost to satisfy Chrome's origin check.
func discoverWSURL(baseWSURL string) (string, error) {
	parsed, err := url.Parse(baseWSURL)
	if err != nil {
		return "", fmt.Errorf("invalid CHROME_WS_URL: %w", err)
	}

	httpURL := fmt.Sprintf("http://%s/json/version", parsed.Host)

	req, err := http.NewRequest("GET", httpURL, nil)
	if err != nil {
		return "", err
	}
	req.Host = "localhost"

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("chrome discovery failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("chrome discovery decode: %w", err)
	}

	// The returned WS URL points to localhost; replace with the original host
	// so Docker containers can reach it.
	discoveredURL, err := url.Parse(result.WebSocketDebuggerURL)
	if err != nil {
		return "", err
	}
	discoveredURL.Host = parsed.Host
	return discoveredURL.String(), nil
}

func fetchViaRemoteChrome(ctx context.Context, baseWSURL string, targetURL string) (string, error) {
	wsURL, err := discoverWSURL(baseWSURL)
	if err != nil {
		return "", fmt.Errorf("remote chrome discovery: %w", err)
	}

	allocCtx, cancel := chromedp.NewRemoteAllocator(ctx, wsURL)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	taskCtx, cancel = context.WithTimeout(taskCtx, 45*time.Second)
	defer cancel()

	var html string
	err = chromedp.Run(taskCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			script := `
				Object.defineProperty(navigator, 'webdriver', {get: () => undefined});
				window.chrome = {runtime: {}};
				Object.defineProperty(navigator, 'plugins', {get: () => [1, 2, 3, 4, 5]});
				Object.defineProperty(navigator, 'languages', {get: () => ['pt-BR', 'pt', 'en-US', 'en']});
			`
			return chromedp.Evaluate(script, nil).Do(ctx)
		}),
		chromedp.Navigate(targetURL),
		chromedp.Sleep(10*time.Second),
		chromedp.OuterHTML("body", &html),
	)
	if err != nil {
		return "", fmt.Errorf("remote chrome: %w", err)
	}

	if len(html) < 5000 {
		return "", fmt.Errorf("page content too small (%d chars), likely blocked by bot detection", len(html))
	}

	return html, nil
}

func fetchViaHeadless(ctx context.Context, targetURL string) (string, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.WindowSize(1920, 1080),
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	taskCtx, cancel = context.WithTimeout(taskCtx, 45*time.Second)
	defer cancel()

	var html string
	err := chromedp.Run(taskCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			script := `
				Object.defineProperty(navigator, 'webdriver', {get: () => undefined});
				window.chrome = {runtime: {}};
				Object.defineProperty(navigator, 'plugins', {get: () => [1, 2, 3, 4, 5]});
				Object.defineProperty(navigator, 'languages', {get: () => ['pt-BR', 'pt', 'en-US', 'en']});
			`
			return chromedp.Evaluate(script, nil).Do(ctx)
		}),
		chromedp.Navigate(targetURL),
		chromedp.Sleep(8*time.Second),
		chromedp.OuterHTML("body", &html),
	)
	if err != nil {
		return "", err
	}

	if len(html) < 5000 {
		return "", fmt.Errorf("page content too small (%d chars), likely blocked by bot detection", len(html))
	}

	return html, nil
}
