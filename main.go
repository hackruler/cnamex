package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"io/ioutil"
	"strings"
	"time"
)

const currentVersion = "1.0.7"

// Function to check CNAME record for a subdomain with timeout
func checkCNAME(subdomain string, timeout time.Duration) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resolver := net.Resolver{}
	cname, err := resolver.LookupCNAME(ctx, subdomain)
	if err != nil {
		return "", false
	}
	return cname, true
}

// Function to show the help menu
func showHelp(cmd *cobra.Command, args []string) {
	fmt.Println("cnamex - Subdomain CNAME Lookup Tool - Made by Mahesh")
	fmt.Println("Usage:")
	fmt.Println("  cnamex [options] <subdomains-file>")
	fmt.Println("Options:")
	fmt.Println("  -h, --help     Show help menu")
	fmt.Println("  -f, --file     Input file containing subdomains")
	fmt.Println("  -o, --output   Output file to save subdomains with CNAME records")
	fmt.Println("  -t, --timeout  Timeout for DNS lookup in seconds (default 10)")
	fmt.Println("  -u, --update   Check for updates")
	fmt.Println("Example:")
	fmt.Println("  cnamex -f subdomains.txt -o output.txt -t 5")
}

// Function to check for updates
func checkForUpdates() {
	latestReleaseURL := "https://api.github.com/repos/hackruler/cnamex/releases/latest"
	resp, err := http.Get(latestReleaseURL)
	if err != nil {
		fmt.Println("Error checking for updates:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed to fetch release info from GitHub")
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	latestVersion := extractVersion(body)
	if latestVersion == "" {
		fmt.Println("Could not determine the latest version")
		return
	}

	if latestVersion == currentVersion {
		fmt.Println("The tool is already up-to-date.")
	} else {
		fmt.Println("Updating to the latest version...")
		downloadLatestVersion(latestVersion)
	}
}

// Function to extract version from the GitHub API response
func extractVersion(body []byte) string {
	bodyStr := string(body)
	versionTag := `"tag_name": "v`
	startIndex := strings.Index(bodyStr, versionTag)
	if startIndex == -1 {
		return ""
	}

	startIndex += len(versionTag)
	endIndex := strings.Index(bodyStr[startIndex:], `"`)
	if endIndex == -1 {
		return ""
	}

	return bodyStr[startIndex : startIndex+endIndex]
}

// Function to download and install the latest version
func downloadLatestVersion(version string) {
	downloadURL := fmt.Sprintf("https://github.com/hackruler/cnamex/releases/download/%s/cnamex-darwin-amd64", version)

	out, err := os.Create("/usr/local/bin/cnamex")
	if err != nil {
		fmt.Println("Error creating the output file:", err)
		return
	}
	defer out.Close()

	resp, err := http.Get(downloadURL)
	if err != nil {
		fmt.Println("Error downloading the release:", err)
		return
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Error writing to the output file:", err)
		return
	}

	err = os.Chmod("/usr/local/bin/cnamex", 0755)
	if err != nil {
		fmt.Println("Error making the binary executable:", err)
		return
	}

	fmt.Println("The tool has been updated to version", version)
}

// Main function to execute the program
func main() {
	var inputFile string
	var outputFile string
	var timeoutSec int

	var rootCmd = &cobra.Command{
		Use:   "cnamex",
		Short: "Subdomain CNAME Lookup Tool",
		Long:  `cnamex allows you to check CNAME records for a list of subdomains.`,
		Run: func(cmd *cobra.Command, args []string) {
			timeout := time.Duration(timeoutSec) * time.Second

			var scanner *bufio.Scanner
			var file *os.File
			var err error

			if inputFile == "" {
				scanner = bufio.NewScanner(os.Stdin)
			} else {
				file, err = os.Open(inputFile)
				if err != nil {
					fmt.Println("Error opening file:", err)
					return
				}
				defer file.Close()
				scanner = bufio.NewScanner(file)
			}

			var output *os.File
			if outputFile != "" {
				output, err = os.Create(outputFile)
				if err != nil {
					fmt.Println("Error creating output file:", err)
					return
				}
				defer output.Close()
			} else {
				output = os.Stdout
			}

			for scanner.Scan() {
				subdomain := scanner.Text()
				cname, found := checkCNAME(subdomain, timeout)
				if found {
					fmt.Fprintln(output, subdomain)
				}
			}

			if err := scanner.Err(); err != nil {
				fmt.Println("Error reading input:", err)
			}
		},
	}

	rootCmd.Flags().StringVarP(&inputFile, "file", "f", "", "Input file containing subdomains")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file to save subdomains with CNAME records")
	rootCmd.Flags().IntVarP(&timeoutSec, "timeout", "t", 10, "Timeout for DNS lookup in seconds (default 10)")
	rootCmd.Flags().BoolP("update", "u", false, "Check for updates")
	rootCmd.SetHelpFunc(showHelp)

	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Check for updates and update the tool",
		Run: func(cmd *cobra.Command, args []string) {
			checkForUpdates()
		},
	}
	rootCmd.AddCommand(updateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

