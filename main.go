package main

import (
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

// Create a map of cronjobs by name for easy lookup
func createCronJobMap(cronJobs []CronJob) map[string]CronJob {
	jobMap := make(map[string]CronJob)
	for _, job := range cronJobs {
		jobMap[job.Name] = job
	}
	return jobMap
}

// Create a map of cronjobs by command to check for duplicate commands with different names
func createCommandMap(cronJobs []CronJob) map[string]string {
	commandMap := make(map[string]string)
	for _, job := range cronJobs {
		commandMap[job.Command] = job.Name
	}
	return commandMap
}

// Helper function to normalize commands by collapsing multiple spaces
func normalizeCommand(command string) string {
	return strings.Join(strings.Fields(command), " ")
}

func compareCommands(prodFile, devFile string, prodConfig, devConfig *Config) {
	prodCronJobs := createCronJobMap(prodConfig.CronJobs)
	devCronJobs := createCronJobMap(devConfig.CronJobs)

	prodCommands := createCommandMap(prodConfig.CronJobs)
	devCommands := createCommandMap(devConfig.CronJobs)

	// Set up the tab writer for formatted output with padding
	w := tabwriter.NewWriter(os.Stdout, 10, 8, 3, ' ', 0)

	// Print header/legend
	fmt.Printf("Comparing Cron Jobs:\n")
	fmt.Printf("Production File: %s\n", prodFile)
	fmt.Printf("Development File: %s\n", devFile)
	fmt.Printf("\n")

	// Print table headers
	fmt.Fprintf(w, "%-40s\t%-70s\n", "Cron Name", "Difference")
	fmt.Fprintf(w, "%-40s\t%-70s\n", "---------", "----------")

	// Check for cronjobs in production that are missing or different in development
	for name, prodJob := range prodCronJobs {
		var differences string

		if devJob, exists := devCronJobs[name]; exists {
			// Normalize commands before comparing
			if normalizeCommand(prodJob.Command) != normalizeCommand(devJob.Command) {
				differences += fmt.Sprintf("%sCommand difference:\n  Production: %s\n  Development: %s%s\n", Red, prodJob.Command, devJob.Command, Reset)
			}
			// Compare schedules for jobs with the same name
			if prodJob.Schedule != devJob.Schedule {
				differences += fmt.Sprintf("%sSchedule difference:\n  Production: %s\n  Development: %s%s\n", Yellow, prodJob.Schedule, devJob.Schedule, Reset)
			}
		} else {
			// If name does not exist, check if command exists under a different name
			if devName, exists := devCommands[prodJob.Command]; exists {
				differences = fmt.Sprintf("Command found with different name in development: %s", devName)
			} else {
				differences = fmt.Sprintf("%sExists in production but not in development%s", LightBlue, Reset)
			}
		}

		// If there are any differences, print the cronjob info
		if differences != "" {
			fmt.Fprintf(w, "%-40s\t%-70s\n", name, differences)
		}
	}

	// Check for cronjobs in development that are missing in production
	for name, devJob := range devCronJobs {
		if _, exists := prodCronJobs[name]; !exists {
			// If name does not exist in production, check if the command exists under a different name
			if prodName, exists := prodCommands[devJob.Command]; exists {
				fmt.Fprintf(w, "%-40s\t%-70s\n", name, fmt.Sprintf("Command found with different name in production: %s", prodName))
			} else {
				fmt.Fprintf(w, "%-40s\t%-70s\n", name, fmt.Sprintf("%sExists in development but not in production%s", LightBlue, Reset))
			}
		}
	}

	// Flush the tab writer to ensure the output is printed in a formatted manner
	w.Flush()
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
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s <production-yaml> <development-yaml>\n", os.Args[0])
	}

	prodFile := os.Args[1]
	devFile := os.Args[2]

	prodConfig, err := parseYAML(prodFile)
	if err != nil {
		log.Fatalf("Error reading production YAML: %v\n", err)
	}

	devConfig, err := parseYAML(devFile)
	if err != nil {
		log.Fatalf("Error reading development YAML: %v\n", err)
	}

	compareCommands(prodFile, devFile, prodConfig, devConfig)
}