// Package cmd provides the command line functions of the crunchy CLI
package cmd

/*
 Copyright 2017-2018 Crunchy Data Solutions, Inc.
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

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	msgs "github.com/crunchydata/postgres-operator/apiservermsgs"
	"github.com/crunchydata/postgres-operator/pgo/api"
	"os"
)

var BackrestOpts string

// createBackrestBackup ....
func createBackrestBackup(args []string, backupOpts string) {
	log.Debugf("createBackrestBackup called %v %s\n", args, backupOpts)

	request := new(msgs.CreateBackrestBackupRequest)
	request.Args = args
	request.Selector = Selector
	request.BackupOpts = backupOpts

	response, err := api.CreateBackrestBackup(httpclient, &SessionCredentials, request)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		os.Exit(2)
	}

	if response.Status.Code == msgs.Ok {
		for k := range response.Results {
			fmt.Println(response.Results[k])
		}
	} else {
		fmt.Println("Error: " + response.Status.Msg)
		os.Exit(2)
	}

	if len(response.Results) == 0 {
		fmt.Println("No clusters found.")
		return
	}

}

// showBackrest ....
func showBackrest(args []string) {
	log.Debugf("showBackrest called %v\n", args)

	for _, v := range args {
		response, err := api.ShowBackrest(httpclient, v, Selector, &SessionCredentials)
		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(2)
		}

		if response.Status.Code != msgs.Ok {
			fmt.Println("Error: " + response.Status.Msg)
			os.Exit(2)
		}

		if len(response.Items) == 0 {
			fmt.Println("No pgBackRest found.")
			return
		}

		log.Debugf("response = %v\n", response)
		log.Debugf("len of items = %d\n", len(response.Items))

		for _, backup := range response.Items {
			printBackrest(&backup)
		}

	}

}

// printBackrest
func printBackrest(result *msgs.ShowBackrestDetail) {
	fmt.Printf("%s%s\n", "", "")
	fmt.Printf("%s%s\n", "", "backrest : "+result.Name)
	fmt.Printf("%s%s\n", "", result.Info)

}
