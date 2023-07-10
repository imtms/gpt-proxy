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
	"github.com/gin-gonic/gin"
	"net/http"
)

type Server struct {
	httpProxy string
	arkoseURL string
	reportURL string
}

func (s Server) Handler() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Any("/health", s.Healthy)
	r.GET("/status", s.Status)
	r.Any("/api/*path", s.Proxy)
	return r
}

func (s Server) Status(ctx *gin.Context) {

}

func (s Server) Healthy(ctx *gin.Context) {
	ctx.Writer.WriteHeader(http.StatusNoContent)
}

func (s Server) Proxy(ctx *gin.Context) {

}
