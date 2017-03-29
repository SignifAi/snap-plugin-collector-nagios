/*
 * Status.go - status file parser
 */

package nagios

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type nagiosServiceStatusSubmap map[string]uint64
type NagiosServiceStatus map[string]nagiosServiceStatusSubmap
type NagiosHostStatus map[string]uint64

// Nagios status file group types
const (
	NS_INFO          = "info"
	NS_HOSTSTATUS    = "hoststatus"
	NS_SERVICESTATUS = "servicestatus"
)

func NagiosStatusMaps(filename string) (hoststatuses NagiosHostStatus, servicestatuses NagiosServiceStatus, err error) {
	hoststatuses = make(NagiosHostStatus)
	servicestatuses = make(NagiosServiceStatus)
	statusFile, err := os.Open(filename)

	if err != nil {
		return nil, nil, err
	}

	statusFileScanner := bufio.NewScanner(statusFile)
	for statusFileScanner.Scan() {
		this_line := statusFileScanner.Text()

		if strings.HasPrefix(this_line, NS_SERVICESTATUS) || strings.HasPrefix(this_line, NS_HOSTSTATUS) {
			thisNagiosStruct, err := scanNagiosStruct(statusFileScanner)
			if err == nil {
				if strings.HasPrefix(this_line, NS_SERVICESTATUS) {
					considered_state := thisNagiosStruct["current_state"]
					if thisNagiosStruct["state_type"] == "0" {
						considered_state = thisNagiosStruct["last_hard_state"]
					}
					if servicestatuses[thisNagiosStruct["host_name"]] == nil {
						servicestatuses[thisNagiosStruct["host_name"]] = make(nagiosServiceStatusSubmap)
					}
					servicestatuses[thisNagiosStruct["host_name"]][thisNagiosStruct["service_description"]], err = strconv.ParseUint(considered_state, 10, 64)
				} else {
					considered_state := thisNagiosStruct["current_state"]
					if thisNagiosStruct["state_type"] == "0" {
						considered_state = thisNagiosStruct["last_hard_state"]
					}
					hoststatuses[thisNagiosStruct["host_name"]], err = strconv.ParseUint(considered_state, 10, 64)
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
