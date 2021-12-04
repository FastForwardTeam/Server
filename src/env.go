/*
Copyright 2021 NotAProton

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
	"log"
	"os"
	"strings"
)

var port string
var dbUser string
var dbPassword string
var dbName string
var ipList []string
var USER_DIR string
var panelDir string
var privPEM string
var pubPEM string

func parseEnv() {
	envVars := []string{"port", "privPEM", "pubPEM", "dbName", "dbUser", "dbPassword"}
	for _, evar := range envVars {
		os.LookupEnv(evar)
		_, set := os.LookupEnv(evar)
		if !set {
			log.Fatalln("Environment variable " + evar + " is not set")
		}
	}
	port = os.Getenv("port")
	privPEM = os.Getenv("privPEM")
	pubPEM = os.Getenv("pubPEM")
	ipList = strings.Split(os.Getenv("banned_ip_list"), "\n")
	dbName = os.Getenv("dbName")
	dbUser = os.Getenv("dbUser")
	dbPassword = os.Getenv("dbPassword")
	panelDir = "./static/"
}
