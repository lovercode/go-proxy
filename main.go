package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/syumai/workers"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		base := "https://generativelanguage.googleapis.com/v1beta/openai"
		origPath := req.URL.Path
		origRawQuery := req.URL.RawQuery
		targetURL := base + origPath
		if origRawQuery != "" {
			targetURL += "?" + origRawQuery
		}

		// 构造 fetch 请求参数
		headers := map[string]string{}
		for k, vs := range req.Header {
			for _, v := range vs {
				headers[k] = v
			}
		}

		// 读取 body
		var body []byte
		if req.Body != nil {
			body, _ = io.ReadAll(req.Body)
		}

		// 用 workers.Fetch 发起请求
		resp, err := workers.Fetch(targetURL, &workers.FetchOptions{
			Method:  req.Method,
			Headers: headers,
			Body:    body,
		})
		if err != nil {
			http.Error(w, "fetch error: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// 透传响应头
		for k, vs := range resp.Header {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})
	workers.Serve(nil)
}
