package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v2"
)

// ANSI color codes
const (
	Reset       = "\033[0m"
	Yellow      = "\033[33m"
	Red         = "\033[31m"
	LightBlue   = "\033[94m"
)

type CronJob struct {
	Command  string `yaml:"command"`
	Name     string `yaml:"name"`
	Schedule string `yaml:"schedule"`
}

type Config struct {
	CronJobs []CronJob `yaml:"cronjobs"`
}

// JSON structure to hold differences
type JobDifference struct {
	CronName    string `json:"cron_name"`
	Type        string `json:"type"`
	Production  string `json:"production,omitempty"`
	Development string `json:"development,omitempty"`
}

func createCronJobMap(cronJobs []CronJob) map[string]CronJob {
	jobMap := make(map[string]CronJob)
	for _, job := range cronJobs {
		jobMap[job.Name] = job
	}
	return jobMap
}

// Helper function to normalize commands by collapsing multiple spaces
func normalizeCommand(command string) string {
	return strings.Join(strings.Fields(command), " ")
}

func compareCommands(prodFile, devFile string, prodConfig, devConfig *Config, jsonOutput bool) {
	prodCronJobs := createCronJobMap(prodConfig.CronJobs)
	devCronJobs := createCronJobMap(devConfig.CronJobs)

	// List to hold differences in case of JSON output
	var differences []JobDifference

	if jsonOutput {
		// Collect differences in JSON format
		for name, prodJob := range prodCronJobs {
			if devJob, exists := devCronJobs[name]; exists {
				if normalizeCommand(prodJob.Command) != normalizeCommand(devJob.Command) {
					differences = append(differences, JobDifference{
						CronName:   name,
						Type:       "Command Difference",
						Production: prodJob.Command,
						Development: devJob.Command,
					})
				}
				if prodJob.Schedule != devJob.Schedule {
					differences = append(differences, JobDifference{
						CronName:   name,
						Type:       "Schedule Difference",
						Production: prodJob.Schedule,
						Development: devJob.Schedule,
					})
				}
			} else {
				differences = append(differences, JobDifference{
					CronName: name,
					Type:     "Exists in production but not in development",
				})
			}
		}

		for name := range devCronJobs {
			if _, exists := prodCronJobs[name]; !exists {
				differences = append(differences, JobDifference{
					CronName: name,
					Type:     "Exists in development but not in production",
				})
			}
		}

		// Output as JSON
		jsonData, err := json.MarshalIndent(differences, "", "  ")
		if err != nil {
			log.Fatalf("Error marshalling to JSON: %v", err)
		}
		fmt.Println(string(jsonData))

	} else {
		// Human-readable output
		w := tabwriter.NewWriter(os.Stdout, 10, 8, 3, ' ', 0)

		fmt.Printf("Comparing Cron Jobs:\n")
		fmt.Printf("Production File: %s\n", prodFile)
		fmt.Printf("Development File: %s\n", devFile)
		fmt.Printf("\n")

		fmt.Fprintf(w, "%-40s\t%-70s\n", "Cron Name", "Difference")
		fmt.Fprintf(w, "%-40s\t%-70s\n", "---------", "----------")

		for name, prodJob := range prodCronJobs {
			var differences string

			if devJob, exists := devCronJobs[name]; exists {
				if normalizeCommand(prodJob.Command) != normalizeCommand(devJob.Command) {
					differences += fmt.Sprintf("%sCommand difference:\n  Production: %s\n  Development: %s%s\n", Red, prodJob.Command, devJob.Command, Reset)
				}
				if prodJob.Schedule != devJob.Schedule {
					differences += fmt.Sprintf("%sSchedule difference:\n  Production: %s\n  Development: %s%s\n", Yellow, prodJob.Schedule, devJob.Schedule, Reset)
				}
			} else {
				differences = fmt.Sprintf("%sExists in production but not in development%s", LightBlue, Reset)
			}

			if differences != "" {
				fmt.Fprintf(w, "%-40s\t%-70s\n", name, differences)
			}
		}

		for name := range devCronJobs {
			if _, exists := prodCronJobs[name]; !exists {
				fmt.Fprintf(w, "%-40s\t%-70s\n", name, fmt.Sprintf("%sExists in development but not in production%s", LightBlue, Reset))
			}
		}

		w.Flush()
	}
}

func parseYAML(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func main() {
	// Add --json flag to switch output to JSON format
	jsonOutput := flag.Bool("json", false, "Output differences in JSON format")
	flag.Parse()

	if len(flag.Args()) < 2 {
		log.Fatalf("Usage: %s <production-yaml> <development-yaml>\n", os.Args[0])
	}

	prodFile := flag.Arg(0)
	devFile := flag.Arg(1)

	prodConfig, err := parseYAML(prodFile)
	if err != nil {
		log.Fatalf("Error reading production YAML: %v\n", err)
	}

	devConfig, err := parseYAML(devFile)
	if err != nil {
		log.Fatalf("Error reading development YAML: %v\n", err)
	}

	compareCommands(prodFile, devFile, prodConfig, devConfig, *jsonOutput)
}