package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/syumai/workers"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// 构造目标 URL：把原请求的 path+query 拼到 Gemini 的 OpenAI‑兼容 base_url
		// 注意这个 base_url 来自官方文档
		// https://ai.google.dev/gemini-api/docs/openai :contentReference[oaicite:0]{index=0}
		base := "https://generativelanguage.googleapis.com/v1beta/openai"
		// 原始请求示例： /v1/chat/completions?stream=true
		origPath := req.URL.Path         // e.g. "/v1/chat/completions"
		origRawQuery := req.URL.RawQuery // e.g. "stream=true"
		targetURL := base + origPath
		if origRawQuery != "" {
			targetURL += "?" + origRawQuery
		}

		// 新建转发请求，保留原方法、Headers、Body
		outReq, err := http.NewRequest(req.Method, targetURL, req.Body)
		if err != nil {
			http.Error(w, "构造请求失败："+err.Error(), http.StatusInternalServerError)
			return
		}
		// 复制所有 header（包括 Authorization: Bearer GEMINI_KEY）
		for k, vs := range req.Header {
			for _, v := range vs {
				outReq.Header.Add(k, v)
			}
		}
		// 确保 Host 为 Google 域名
		outReq.Host = (&url.URL{Host: "generativelanguage.googleapis.com"}).Host

		// 发起请求
		client := http.DefaultClient
		resp, err := client.Do(outReq)
		if err != nil {
			http.Error(w, "转发请求失败："+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// 把响应头状态码透传回去
		for k, vs := range resp.Header {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)

		// 流式转发 body（支持 stream=true）
		io.Copy(w, resp.Body)
	})
	workers.Serve(nil) // use http.DefaultServeMux
}
