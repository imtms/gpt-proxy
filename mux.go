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
	"io"
	"log"
	kithttp "net/http"
	"strings"

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

	r.Any("/health", s.Healthy)
	r.GET("/status", s.Status)
	r.Any("/api/*path", s.Proxy)
	return r
}

func (s Server) Status(ctx *gin.Context) {
	ctx.Writer.WriteHeader(http.StatusNoContent)
}

func (s Server) Healthy(ctx *gin.Context) {
	ctx.Writer.WriteHeader(http.StatusNoContent)
}

func (s Server) Proxy(ctx *gin.Context) {
	s.client.SetCookies(ctx.Request.URL, []*http.Cookie{})

	method := ctx.Request.Method
	url := ChatOpenAI + ""

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

		}
	}

	var body io.Reader

	req, err := http.NewRequest("", "", body)
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

		log.Printf("ERR: http status code %s body resp: %s", resp.StatusCode, string(errBody))
		ctx.JSON(resp.StatusCode, New(string(errBody)))
		return
	}

	ctx.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")

}

func Auth(header kithttp.Header) string {
	xAuth := header.Get("X-Authorization")
	if xAuth == "" {
		return header.Get("Authorization")
	}
	return xAuth
}

func IsGPT4(s string) bool {
	// gpt-4
	// gpt-4-code-interpreter
	if strings.Contains(s, "gpt-4") {
		return true
	}
	return false
}
