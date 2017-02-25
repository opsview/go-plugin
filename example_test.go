package plugin

func Example() {
	// example command line options
	var opts struct {
		Hostname string `short:"H" long:"hostname" description:"Host" default:"localhost"`
		Port     int    `short:"p" long:"port" description:"Port" default:"123"`
	}

	// initialize check
	check := New("check_service", "1.0.0")

	// parse command line arguments
	if err := check.ParseArgs(&opts); err != nil {
		check.ExitCritical("Error parsing arguments: %s", err)
	}

	// defer exit (also handles panics)
	defer check.Final()

	// add message to output
	check.AddMessage("Service %s:%d", opts.Hostname, opts.Port)

	// add metrics
	check.AddMetric("uptime", 123456, "s")
	check.AddMetric("processes", 789)
}
