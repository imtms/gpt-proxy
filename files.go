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
	kithttp "net/http"
	"strings"
)

type CreateFilesResponse struct {
	Status    string `json:"status"`
	UploadUrl string `json:"upload_url"`
	FileId    string `json:"file_id"`
}

func Auth(header kithttp.Header) string {
	xAuth := header.Get("X-Authorization")
	if xAuth == "" {
		return header.Get("Authorization")
	}
	return xAuth
}

func IsGPT4(s string) bool {
	if s == "gpt-4" {
		return true
	}
	return false
}

func IsGPT4Code(s string) bool {
	if s == "gpt-4-code-interpreter" {
		return true
	}
	return false
}

func URL(ctx *gin.Context) string {
	var url string
	if strings.HasPrefix(ctx.Param("path"), "/conversion") {
		url = ChatOpenAI + "/public-api" + ctx.Param("path") + "?" + ctx.Request.URL.RawQuery
	} else if ctx.Request.URL.RawQuery != "" {
		url = ChatOpenAI + "/backend-api" + ctx.Param("path") + "?" + ctx.Request.URL.RawQuery
	} else {
		url = ChatOpenAI + "/backend-api" + ctx.Param("path")
	}
	return url
}

func IsConversation(path string) bool {
	if path == "/conversation" {
		return true
	}
	return false
}
