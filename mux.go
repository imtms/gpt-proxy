/*
Copyright 2022 The deepauto-io LLC.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gpt_proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"log"

	"github.com/tidwall/gjson"

	"github.com/acheong08/funcaptcha"
	http "github.com/bogdanfinn/fhttp"
	tlscc "github.com/bogdanfinn/tls-client"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	cors "github.com/rs/cors/wrapper/gin"
)

type Server struct {
	client    tlscc.HttpClient
	arkoseURL string
	reportURL string
}

func NewServer(cc tlscc.HttpClient, arkoseURL, reportURL string) *Server {
	return &Server{
		client:    cc,
		arkoseURL: arkoseURL,
		reportURL: reportURL,
	}
}

func (s Server) Handler() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.Default())

	r.Any("/health", s.Healthy)  // 健康检查
	r.GET("/status", s.Status)   // 上报当前代理状态 [异常，正常]
	r.Any("/api/*path", s.Proxy) // 代理GPT-4对话
	r.GET("/files", s.Files)     // gpt-4-code-interpreter
	return r
}

func (s Server) Status(ctx *gin.Context) {
	ctx.Writer.WriteHeader(http.StatusNoContent)
}

func (s Server) Healthy(ctx *gin.Context) {
	ctx.Writer.WriteHeader(http.StatusNoContent)
}

func (s Server) Files(ctx *gin.Context) {

}

func (s Server) Proxy(ctx *gin.Context) {
	s.client.SetCookies(ctx.Request.URL, []*http.Cookie{})

	var (
		url    = URL(ctx)
		method = ctx.Request.Method
		body   io.Reader
	)

	if IsConversation(ctx.Param("path")) {
		in := new(OpenAIChatRequest)
		if err := ctx.BindJSON(in); err != nil {
			log.Printf("JSON: bind json err: %s", err)
			ctx.JSON(http.StatusBadRequest, New(err.Error()))
			return
		}

		if len(in.Messages) != 0 {
			for _, message := range in.Messages {
				if message.ID == "" {
					message.ID = uuid.New().String()
				}
				if message.Author.Role == "" {
					message.Author.Role = "user"
				}
			}
		}

		if IsGPT4(in.Model) {
			if s.arkoseURL == "" {
				arkoseToken, err := funcaptcha.GetOpenAIToken()
				if err != nil {
					log.Printf("ERR: funcaptcha get arkose token err:%s", err)
					ctx.JSON(http.StatusInternalServerError, New(err.Error()))
					return
				}
				in.ArkoseToken = arkoseToken
			} else {
				arkreq, err := http.NewRequest("GET", s.arkoseURL, http.NoBody)
				if err != nil {
					log.Printf("ERR: %s get arkose token err:%s", s.arkoseURL, err)
					ctx.JSON(http.StatusInternalServerError, New(err.Error()))
					return
				}

				resp, err := s.client.Do(arkreq)
				if err != nil {
					log.Printf("ERR: %s do arkose token err:%s", s.arkoseURL, err)
					ctx.JSON(http.StatusInternalServerError, New(err.Error()))
					return
				}
				defer resp.Body.Close()

				jsonBuf, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Printf("ERR: %s do arkose token err:%s", s.arkoseURL, err)
					ctx.JSON(http.StatusInternalServerError, New(err.Error()))
					return
				}

				arkoseToken := gjson.Get(string(jsonBuf), "token").String()
				if arkoseToken == "" {
					ctx.JSON(http.StatusInternalServerError, New("arkose token is empty"))
					return
				}
				in.ArkoseToken = arkoseToken
			}
		}
		jsonBytes, _ := json.Marshal(in)
		body = bytes.NewBuffer(jsonBytes)
	} else {
		body = ctx.Request.Body
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Printf("ERR: http new request err: %s", err)
		ctx.JSON(http.StatusInternalServerError, New(err.Error()))
		return
	}

	req.Header.Set("Authorization", Auth(ctx.Request.Header))
	req.Header.Set("user-agent", UserAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("ERR: http do request err: %s", err)
		ctx.JSON(http.StatusInternalServerError, New(err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("ERR: http status code %s err: %s", resp.StatusCode, err)
			ctx.JSON(resp.StatusCode, New(err.Error()))
			return
		}

		log.Printf("ERR: http status code %d body resp: %s", resp.StatusCode, string(errBody))
		ctx.JSON(resp.StatusCode, New(string(errBody)))
		return
	}

	ctx.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")

	_, err = io.Copy(ctx.Writer, resp.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, New(err.Error()))
		return
	}
}
