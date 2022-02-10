/*
Copyright 2021, 2022 NotAProton, mockuser404

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
	"regexp"

	anyascii "github.com/anyascii/go"
	"github.com/microcosm-cc/bluemonday"
)

var (
	domainRegex             = regexp.MustCompile(`(?:[a-zA-Z0-9](?:[a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,6}`)
	pathRegex               = regexp.MustCompile(`[^a-z\d&=+.\-_~/?#'%]+`)
	bmStrict                = bluemonday.StrictPolicy()
	whiteSpaceRegex         = regexp.MustCompile(`(\r?\n)|( {2,})|(\t)`)
	usernameSanitizingRegex = regexp.MustCompile(`[^a-zA-Z0-9_]+`)
)

func sanitizeForLog(input string) string {

	sanitizedInput := anyascii.Transliterate(input)
	//no newline
	sanitizedInput = whiteSpaceRegex.ReplaceAllString(sanitizedInput, "")
	//truncate, aribitrary length because why not
	if len(sanitizedInput) > 70 {
		sanitizedInput = sanitizedInput[:70]
	}
	sanitizedInput = `"` + sanitizedInput + `"`
	return sanitizedInput
}

func sanitizePath(s string) string {
	return pathRegex.ReplaceAllString(s, "")
}

func sanitizeDomain(s string) string {
	return domainRegex.FindString(s)
}

func sanitizeUsername(s string) string {
	if len(s) > 20 {
		s = s[:20]
	}
	return usernameSanitizingRegex.ReplaceAllString(s, "")
}
