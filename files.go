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

type createFileRequest struct {
	FileName string `json:"file_name"`
	FileSize int    `json:"file_size"`
	UseCase  string `json:"use_case"`
}

type processUploadRequest struct {
	ConversationId string `json:"conversation_id"`
	FileId         string `json:"file_id"`
	FileName       string `json:"file_name"`
}

type CreateFilesResponse struct {
	Status    string `json:"status"`
	UploadUrl string `json:"upload_url"`
	FileId    string `json:"file_id"`
}

type processUploadResponse struct {
	Status      string `json:"status"`
	DownloadUrl string `json:"download_url"`
}

func (r *createFileRequest) Validate() error {
	if r.FileName == "" {
		return New("file name is empty")
	}

	if r.FileSize <= 0 {
		return New("file size is <=0")
	}

	if r.UseCase == "" {
		return New("file use case is empty")
	}
	return nil
}

func (r *processUploadRequest) Validate() error {
	if r.FileId == "" {
		return New("file id is empty")
	}

	if r.FileName == "" {
		return New("file name is empty")
	}

	return nil
}

func Auth(header kithttp.Header) string {
	xAuth := header.Get("X-Authorization")
	if xAuth == "" {
		return header.Get("Authorization")
	}
	return xAuth
}

func IsGPT4(s string) bool {
	if strings.Contains(s, "gpt-4") {
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
