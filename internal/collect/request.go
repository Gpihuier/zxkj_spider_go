package collect

import (
	"context"
	"fmt"

	"github.com/antchfx/htmlquery"
	"github.com/imroc/req/v3"
	"golang.org/x/net/html"
)

type Request struct {
	Client *req.Client
}

func NewRequest() *Request {
	return &Request{
		Client: req.C().
			EnableDumpEachRequest().
			SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36").
			SetTLSFingerprintChrome().
			ImpersonateChrome().
			SetCommonHeaders(map[string]string{
				"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
				"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
				"Cookie":          "Hm_lvt_b249ff02d75c7db5d3b272499b5ad516=1716184757; Hm_lvt_134f43513d609acd8dbd02453e16940                                            bmx f=1716185436; Hm_lvt_0f54047fed93dd36c9ed5f16f6c080b8=1717409270,1717468394; Hm_lpvt_0f54047fed93dd36c9ed5f16f6c080b8=1717470526",
			}).
			SetCommonRetryCount(3).
			OnAfterResponse(func(client *req.Client, resp *req.Response) error {
				if resp.Err != nil { // Ignore when there is an underlying error, e.g. network error.
					return nil
				}
				// Treat non-successful responses as errors, record raw dump content in error message.
				if !resp.IsSuccessState() { // Status code is not between 200 and 299.
					resp.Err = fmt.Errorf("bad response, raw content:\n%s", resp.Dump())
				}
				return nil
			}),
	}
}

func (r *Request) SetDevMode() *Request {
	r.Client = r.Client.DevMode()
	return r
}

func (r *Request) Get(ctx context.Context, url string) (*req.Response, error) {
	return r.Client.R().SetContext(ctx).Get(url)
}

func (r *Request) GetByParse(ctx context.Context, url string, callback func(doc *html.Node) error) error {
	resp, err := r.Client.R().
		SetContext(ctx).
		Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Pass resp.Body to goquery.
	doc, err := htmlquery.Parse(resp.Body)
	if err != nil { // Append raw dump content to error message if goquery parse failed to help troubleshoot.
		return fmt.Errorf("failed to parse html: %s, raw content:\n%s", err.Error(), resp.Dump())
	}
	err = callback(doc)
	if err != nil {
		err = fmt.Errorf("%s, raw content:\n%s", err.Error(), resp.Dump())
	}

	return err
}
