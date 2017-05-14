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

/*
 * Status.go - status file parser
 */

package nagios

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

type nagiosServiceStatusSubmap map[string]map[string]string
type NagiosServiceStatus map[string]nagiosServiceStatusSubmap
type NagiosHostStatus map[string]map[string]string

// Nagios status file group types
const (
	NS_INFO          = "info"
	NS_HOSTSTATUS    = "hoststatus"
	NS_SERVICESTATUS = "servicestatus"
)

func NagiosStatusMaps(statusFile io.Reader) (hoststatuses NagiosHostStatus, servicestatuses NagiosServiceStatus, err error) {
	hoststatuses = make(NagiosHostStatus)
	servicestatuses = make(NagiosServiceStatus)

	statusFileScanner := bufio.NewScanner(statusFile)
	for statusFileScanner.Scan() {
		this_line := statusFileScanner.Text()

		if strings.HasPrefix(this_line, NS_SERVICESTATUS) || strings.HasPrefix(this_line, NS_HOSTSTATUS) {
			thisNagiosStruct, err := scanNagiosStruct(statusFileScanner)
			if err == nil {
				if strings.HasPrefix(this_line, NS_SERVICESTATUS) {
					if _, ok := servicestatuses[thisNagiosStruct["host_name"]]; !ok {
						servicestatuses[thisNagiosStruct["host_name"]] = make(nagiosServiceStatusSubmap)
					}
					servicestatuses[thisNagiosStruct["host_name"]][thisNagiosStruct["service_description"]] = thisNagiosStruct
				} else {
					hoststatuses[thisNagiosStruct["host_name"]] = thisNagiosStruct
				}
			} else {
				fmt.Println("Error scanning struct:", err)
			}
		}
	}
	return hoststatuses, servicestatuses, err
}

func scanNagiosStruct(scanner *bufio.Scanner) (structAsMap map[string]string, err error) {
	structAsMap = make(map[string]string)
	for scanner.Scan() {
		/*
		 * Here, we should be just after the { at the end of the line defining the struct
		 * As soon as we see a "}" we should back out. (scanner.Scan() will return False on
		 * EOF but scanner.Err() will return nil)
		 *
		 * Regardless of whether the contained struct has an error, we should still scan to the
		 * end. If the struct _has_ no end, we'll just hit EOF...
		 */
		thisLine := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(thisLine), "}") {
			// END OF STRUCT
			break
		}

		if err == nil {
			kvslice := strings.SplitN(strings.TrimSpace(thisLine), "=", 2)
			if len(kvslice) < 2 {
				err = errors.New("Invalid k/v pair in struct: " + thisLine)
				continue
			}
			structAsMap[kvslice[0]] = kvslice[1]
		}
	}
	if err == nil {
		// only probable error is in the scanner
		// scanner.Err() will also return 'nil' for EOF
		err = scanner.Err()
	}
	return structAsMap, err
}
