/*
http://www.apache.org/licenses/LICENSE-2.0.txt
Copyright 2017 SignifAI Inc
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
	"os"

	"github.com/signifai/snap-plugin-collector-nagios/nagios"
)

func mainOld() {
	/*
		fd, err := os.Open("someFile.txt")
		var readBuffer []byte = make([]byte, 512)
		var BytesRead int

		for BytesRead, err = fd.Read(readBuffer); err != nil && err != io.EOF; BytesRead, err = fd.Read(readBuffer) {
			fmt.Print(string(readBuffer[0:BytesRead]))
		}
		// One last output because EOF terminates the for
		fmt.Print(string(readBuffer[3:5]))

		if err != nil {
			fmt.Println("Got err", err)
		}

		fmt.Println(BytesRead, "bytes read")
	*/
	statusFile, err := os.Open("/home/zcarlson/FakeNagios/var/status.dat")
	if err == nil {
		hoststatuses, servicestatuses, err := nagios.NagiosStatusMaps(statusFile)
		if err == nil {
			fmt.Println("Host statuses:")
			for host, hostStatus := range hoststatuses {
				fmt.Println("  ", host+":", hostStatus)
			}
			fmt.Println("====")
			fmt.Println("Service statuses:")
			for host, serviceMap := range servicestatuses {
				for service, serviceStatus := range serviceMap {
					fmt.Println("  ", host, service+":", serviceStatus)
				}
			}
		}
	}
}
