package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"io/ioutil"
	"strings"
)

const currentVersion = "1.1.0" // Set the current version of the tool

// Function to check CNAME record for a subdomain
func checkCNAME(subdomain string) string {
	cname, err := net.LookupCNAME(subdomain)
	if err != nil {
		return ""
	}
	return cname
}

// Function to handle the help command
func showHelp(cmd *cobra.Command, args []string) {
	fmt.Println("cnamex - Subdomain CNAME Lookup Tool - Made by Mahesh")
	fmt.Println("Usage:")
	fmt.Println("  cnamex [options] <subdomains-file>")
	fmt.Println("Options:")
	fmt.Println("  -h, --help     Show help menu")
	fmt.Println("  -f, --file     Input file containing subdomains")
	fmt.Println("  -o, --output   Output file to save subdomains with CNAME records")
	fmt.Println("  -u, --update   Check for updates")
	fmt.Println("Example:")
	fmt.Println("  cnamex -f subdomains.txt -o output.txt")
}

// Function to check for updates
func checkForUpdates() {
	// GitHub API URL to fetch the latest release version
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

	// Parse the response body to get the version info
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	// Extract the version from the response
	latestVersion := extractVersion(body)
	if latestVersion == "" {
		fmt.Println("Could not determine the latest version")
		return
	}

	// Compare with the current version
	if latestVersion == currentVersion {
		fmt.Println("The tool is already up-to-date.")
	} else {
		fmt.Println("Updating to the latest version...")

		// Download the latest release binary (assuming it's hosted as a binary on GitHub)
		downloadLatestVersion(latestVersion)
	}
}

// Function to extract version from the GitHub API response
func extractVersion(body []byte) string {
	// Extract version from the GitHub API response (using simple string manipulation here)
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
	// GitHub download URL for the latest release binary
	downloadURL := fmt.Sprintf("https://github.com/hackruler/cnamex/releases/download/%s/cnamex-darwin-amd64", version) // Modify for your platform

	// Create the output file to save the binary
	out, err := os.Create("/usr/local/bin/cnamex")
	if err != nil {
		fmt.Println("Error creating the output file:", err)
		return
	}
	defer out.Close()

	// Get the latest release binary
	resp, err := http.Get(downloadURL)
	if err != nil {
		fmt.Println("Error downloading the release:", err)
		return
	}
	defer resp.Body.Close()

	// Copy the downloaded content to the output file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Error writing to the output file:", err)
		return
	}

	// Make the file executable
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

	// Create a new cobra command
	var rootCmd = &cobra.Command{
		Use:   "cnamex",
		Short: "Subdomain CNAME Lookup Tool",
		Long:  `cnamex allows you to check CNAME records for a list of subdomains.`,
		Run: func(cmd *cobra.Command, args []string) {
			var scanner *bufio.Scanner
			var file *os.File
			var err error

			if inputFile == "" {
				// If no file is provided, use standard input (piped input)
				scanner = bufio.NewScanner(os.Stdin)
			} else {
				// If a file is provided, read from the file
				file, err = os.Open(inputFile)
				if err != nil {
					fmt.Println("Error opening file:", err)
					return
				}
				defer file.Close()
				scanner = bufio.NewScanner(file)
			}

			// Prepare output (either file or terminal)
			var output *os.File
			if outputFile != "" {
				// If an output file is provided, open the file for writing
				output, err = os.Create(outputFile)
				if err != nil {
					fmt.Println("Error creating output file:", err)
					return
				}
				defer output.Close()
				fmt.Fprintln(output, "Subdomains with CNAME records:")
			} else {
				// If no output file, print to terminal
				output = os.Stdout
				fmt.Fprintln(output, "Subdomains with CNAME records:")
			}

			// Read each subdomain and check for CNAME
			for scanner.Scan() {
				subdomain := scanner.Text()
				cname := checkCNAME(subdomain)
				if cname != "" {
					// Output subdomain only if CNAME record exists
					fmt.Fprintln(output, subdomain)
				}
			}

			// Check for scanning errors
			if err := scanner.Err(); err != nil {
				fmt.Println("Error reading input:", err)
			}
		},
	}

	// Define flags for the command
	rootCmd.Flags().StringVarP(&inputFile, "file", "f", "", "Input file containing subdomains")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file to save subdomains with CNAME records")
	rootCmd.Flags().BoolP("update", "u", false, "Check for updates")
	rootCmd.MarkFlagRequired("file")
	rootCmd.SetHelpFunc(showHelp)

	// Add update command
	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Check for updates and update the tool",
		Run: func(cmd *cobra.Command, args []string) {
			checkForUpdates()
		},
	}
	rootCmd.AddCommand(updateCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
