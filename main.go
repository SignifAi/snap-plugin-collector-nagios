package main

import (
	"fmt"
	"github.com/signifai/snap-plugin-collector-nagios/nagios"
)

func main() {
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
	hoststatuses, servicestatuses, err := nagios.NagiosStatusMaps("/usr/local/nagios/var/status.dat")
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
