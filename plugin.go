package plugin

import (
	"bytes"
	"fmt"
	"github.com/jessevdk/go-flags"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Plugin struct {
	name               string
	status             Status
	messages           []string
	metrics            checkMetrics
	Version            string
	Preamble           string
	Description        string
	AllMetricsInOutput bool
	MessageSeparator   string
}

type checkMetric struct {
	value    interface{}
	uom      string
	warn     string
	critical string
}

type checkMetrics map[string]*checkMetric

func New(name, version string) *Plugin {
	return &Plugin{
		name:               name,
		status:             OK,
		messages:           make([]string, 0),
		metrics:            make(checkMetrics),
		Version:            version,
		AllMetricsInOutput: false,
		MessageSeparator:   ", ",
	}
}

func (p *Plugin) AddMetric(name string, value interface{}, args ...string) error {
	args_count := len(args)
	_, ok := p.metrics[name]
	if !ok {
		p.metrics[name] = &checkMetric{}
	} else {
		return fmt.Errorf("Duplicated metric %s;", name)
	}

	p.metrics[name].value = value
	if args_count >= 1 {
		p.metrics[name].uom = args[0]
	}

	val, err := i2f(value)
	if err != nil {
		return fmt.Errorf("Invalid value of %s: %v;", name, value)
	}

	var alert_message string

	if args_count == 2 || args_count == 3 {
		var threshold_breached bool
		for i, a := range args[1:] {
			var threshold_name string
			var invert bool

			if len(a) == 0 {
				continue
			}

			arg := strings.TrimPrefix(a, "@")
			if a != arg {
				invert = true
			}

			thresh := strings.Split(arg, ":")

			switch i {
			case 0:
				threshold_name = "warning"
				p.metrics[name].warn = a
			case 1:
				threshold_name = "critical"
				p.metrics[name].critical = a
			}

			switch len(thresh) {
			case 1:
				// v < X
				t_max, err := strconv.ParseFloat(thresh[0], 64)
				if err != nil {
					return fmt.Errorf("Invalid format of %s threshold %s: %s", threshold_name, name, a)
				}
				threshold_breached = val < 0 || val > t_max
			case 2:
				switch {
				case thresh[0] == "~":
					t_max, err := strconv.ParseFloat(thresh[1], 64)
					if err != nil {
						return fmt.Errorf("Invalid format of %s threshold %s: %s", threshold_name, name, a)
					}
					threshold_breached = val > t_max
				case thresh[1] == "":
					t_min, err := strconv.ParseFloat(thresh[0], 64)
					if err != nil {
						return fmt.Errorf("Invalid format of %s threshold %s: %s", threshold_name, name, a)
					}
					threshold_breached = val < t_min
				default:
					t_min, err := strconv.ParseFloat(thresh[0], 64)
					if err != nil {
						return fmt.Errorf("Invalid format of %s threshold %s: %s", threshold_name, name, a)
					}
					t_max, err := strconv.ParseFloat(thresh[1], 64)
					if err != nil {
						return fmt.Errorf("Invalid format of %s threshold %s: %s", threshold_name, name, a)
					}
					if t_min > t_max {
						return fmt.Errorf("Invalid format of %s threshold %s: %s", threshold_name, name, a)
					}
					threshold_breached = val < t_min || val > t_max
				}
			default:
				return fmt.Errorf("Invalid format of %s threshold %s: %s", threshold_name, name, a)
			}

			if invert {
				threshold_breached = !threshold_breached
			}

			if threshold_breached {
				p.UpdateStatus(Status(i + 1)) // i=0 warning, i=1 critical
				if invert {
					alert_message = fmt.Sprintf("%s is %v (inside %s)", name, value, a)
				} else {
					alert_message = fmt.Sprintf("%s is %v (outside %s)", name, value, a)
				}
			}

		}
	} else {
		return fmt.Errorf("Too many arguments")
	}

	if len(alert_message) > 0 {
		p.AddMessage(alert_message)
	} else if p.AllMetricsInOutput {
		p.AddMessage(fmt.Sprintf("%s is %v", name, value))
	}

	return nil
}

func (p *Plugin) AddMessage(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.messages = append(p.messages, msg)
}

func (p *Plugin) AddResult(code Status, format string, args ...interface{}) {
	p.UpdateStatus(code)
	p.AddMessage(format, args...)
}

func (p *Plugin) Final() {
	fmt.Printf("%s: ", p.status.String())
	fmt.Printf(strings.Join(p.messages, p.MessageSeparator))
	if len(p.metrics) > 0 {
		var sorted []string
		sorted = make([]string, 0, len(p.metrics))

		fmt.Printf(" |")
		for k := range p.metrics {
			sorted = append(sorted, k)
		}
		sort.Strings(sorted)
		for _, k := range sorted {
			fmt.Printf(" %s=%v%s;%s;%s;;",
				k,
				p.metrics[k].value,
				p.metrics[k].uom,
				p.metrics[k].warn,
				p.metrics[k].critical,
			)
		}
	}
	fmt.Printf("\n")
	os.Exit(p.status.ExitCode())
}

func (p *Plugin) SetMessage(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	p.messages = []string{msg}
}

func (p *Plugin) exit(code Status, format string, args ...interface{}) {
	p.status = code
	p.SetMessage(format, args...)
	p.Final()
}

func (p *Plugin) ExitOK(format string, args ...interface{}) {
	p.exit(OK, format, args...)
}

func (p *Plugin) ExitUnknown(format string, args ...interface{}) {
	p.exit(UNKNOWN, format, args...)
}

func (p *Plugin) ExitWarning(format string, args ...interface{}) {
	p.exit(WARNING, format, args...)
}

func (p *Plugin) ExitCritical(format string, args ...interface{}) {
	p.exit(CRITICAL, format, args...)
}

func (p *Plugin) ParseArgs(opts interface{}) error {
	var err error

	var builtin struct {
		Help bool `short:"h" long:"help" description:"Show this help message"`
	}
	parser := flags.NewParser(opts, 0)
	_, err = parser.AddGroup("Default Options", "", &builtin)

	g := parser.Command.Group.Find("Application Options")
	if g != nil {
		g.ShortDescription = "Plugin Options"
	}

	_, err = parser.Parse()

	if builtin.Help {
		fmt.Printf("%s v%s\n", p.name, strings.TrimPrefix(p.Version, "v"))
		if len(p.Preamble) > 0 {
			fmt.Println(p.Preamble)
		}
		parser.Options = flags.HelpFlag
		var b bytes.Buffer
		parser.WriteHelp(&b)
		fmt.Println(b.String())

		if len(p.Description) > 0 {
			fmt.Println(p.Description)
		}
		os.Exit(0)
	}

	return err
}

func (p *Plugin) UpdateStatus(status Status) {
	if status >= 0 && status <= 3 {
		if int(status) > int(p.status) {
			p.status = status
		}
	} else {
		panic("Invalid status code")
	}
}

func (p *Plugin) Status() Status {
	return p.status
}

func i2f(v interface{}) (float64, error) {
	var f float64
	var err error

	switch v.(type) {
	case float32:
		f = float64(v.(float32))
	case float64:
		f = v.(float64)
	default:
		f, err = strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
	}
	return f, err
}
