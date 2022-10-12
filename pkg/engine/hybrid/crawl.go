package hybrid

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/katana/pkg/navigation"
	"github.com/projectdiscovery/katana/pkg/utils/queue"
	"github.com/projectdiscovery/retryablehttp-go"
)

func (c *Crawler) navigateRequest(ctx context.Context, httpclient *retryablehttp.Client, queue *queue.VarietyQueue, parseResponseCallback func(nr navigation.Request), browser *rod.Browser, request navigation.Request, rootHostname string) (*navigation.Response, error) {
	depth := request.Depth + 1
	response := &navigation.Response{
		Depth:        depth,
		Options:      c.options,
		RootHostname: rootHostname,
	}

	page, err := browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return nil, errors.Wrap(err, "could not create target")
	}
	defer page.Close()

	pageRouter := page.HijackRequests()
	if err := pageRouter.Add("*", "", c.makeRoutingHandler(queue, depth, rootHostname, httpclient, parseResponseCallback)); err != nil {
		return nil, errors.Wrap(err, "could not add router")
	}
	go pageRouter.Run()
	defer func() {
		if err := pageRouter.Stop(); err != nil {
			gologger.Warning().Msgf("%s\n", err)
		}
	}()

	timeout := time.Duration(c.options.Options.Timeout) * time.Second
	page = page.Timeout(timeout)

	// wait the page to be fully loaded and becoming idle
	waitNavigation := page.WaitNavigation(proto.PageLifecycleEventNameDOMContentLoaded)

	if err := page.Navigate(request.URL); err != nil {
		return nil, errors.Wrap(err, "could not navigate target")
	}
	waitNavigation()

	// Wait for the window.onload event
	if err := page.WaitLoad(); err != nil {
		gologger.Warning().Msgf("\"%s\" on wait load: %s\n", request.URL, err)
	}

	// wait for idle the network requests
	if err := page.WaitIdle(timeout); err != nil {
		gologger.Warning().Msgf("\"%s\" on wait idle: %s\n", request.URL, err)
	}

	body, err := page.HTML()
	if err != nil {
		return nil, errors.Wrap(err, "could not get html")
	}

	parsed, _ := url.Parse(request.URL)
	response.Resp = &http.Response{Header: make(http.Header), Request: &http.Request{URL: parsed}}
	response.Body = []byte(body)
	response.Reader, err = goquery.NewDocumentFromReader(bytes.NewReader(response.Body))
	if err != nil {
		return nil, errors.Wrap(err, "could not parse html")
	}

	return response, nil
}
