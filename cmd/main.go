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

package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	tlsclient "github.com/bogdanfinn/tls-client"
	gpt_proxy "yho.io/gptproxy"
)

func main() {
	rand.Seed(time.Now().Unix())

	config, err := gpt_proxy.Environ()
	if err != nil {
		log.Panic(fmt.Sprintf("load env config err:%s", err))
	}

	if err := config.Validate(); err != nil {
		log.Panic(fmt.Sprintf("config validate err: %s", err))
	}

	options := []tlsclient.HttpClientOption{
		tlsclient.WithTimeoutSeconds(360),
		tlsclient.WithClientProfile(tlsclient.Safari_IOS_16_0),
		tlsclient.WithNotFollowRedirects(),
		tlsclient.WithCookieJar(tlsclient.NewCookieJar()), // create cookieJar instance and pass it as argument
	}
	cc, err := tlsclient.NewHttpClient(tlsclient.NewNoopLogger(), options...)
	if err != nil {
		log.Panic(fmt.Sprintf("tls new client err:%s", err))
	}

	if config.HttpProxy != "" {
		if err := cc.SetProxy(config.HttpProxy); err != nil {
			log.Panic(fmt.Sprintf("tls set proxy err:%s", err))
		}
		log.Println("set proxy success: ", config.HttpProxy)
	}

	mux := gpt_proxy.New(cc, config.ArkoseURL, config.ReportURL)
	if err := mux.Handler().Run(); err != nil {
		log.Panic(fmt.Sprintf("http server run err:%s", err))
	}
}
