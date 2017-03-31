package main

import(
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/signifai/snap-plugin-collector-nagios/nagios"
)

func main() {
	plugin.StartCollector(nagios.NagiosPlugin{}, nagios.Name, nagios.Version, plugin.RoutingStrategy(plugin.StickyRouter))
}