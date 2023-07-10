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

import "github.com/kelseyhightower/envconfig"

const (
	UserAgent  = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36"
	ChatOpenAI = "https://chat.openai.com"
)

type Config struct {
	HttpProxy string `envconfig:"HTTP_PROXY"`
	ReportURL string `envconfig:"REPORT_URL"`
	ArkoseURL string `envconfig:"ARKOSE_URL"`
}

// Environ returns the settings from the environment.
func Environ() (Config, error) {
	cfg := Config{}
	err := envconfig.Process("", &cfg)
	return cfg, err
}

func (c Config) Validate() error {
	return nil
}

// Error represents a json-encoded API error.
type Error struct {
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

// New returns a new error message.
func New(text string) error {
	return &Error{Message: text}
}

type OpenAIChatRequest struct {
	Action                     string         `json:"action"`
	Messages                   OpenAIMessages `json:"messages"`
	Model                      string         `json:"model"`
	ConversationId             string         `json:"conversation_id,omitempty"`
	ParentMessageId            string         `json:"parent_message_id"`
	TimezoneOffsetMin          int            `json:"timezone_offset_min,omitempty"`
	HistoryAndTrainingDisabled bool           `json:"history_and_training_disabled,omitempty"`
	ArkoseToken                string         `json:"arkose_token"`
	PluginIDS                  []string       `json:"plugin_ids,omitempty"`
}

type OpenAIMessages []OpenAIMessage

type OpenAIMessage struct {
	ID       string      `json:"id"`
	Author   Author      `json:"author"`
	Content  Content     `json:"content"`
	Metadata interface{} `json:"metadata"`
}

type Author struct {
	Role string `json:"role"`
}

type Content struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}
