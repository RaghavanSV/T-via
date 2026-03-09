package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

var ()

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
	var cmd string = ""
	var output []map[string]interface{}

	for end != "end" {
		fmt.Print("command> ")
		fmt.Scan(&cmd)

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
		time.Sleep(50 * time.Second)

		//for reading the output

		req3, _ := http.NewRequest("GET", obj.UrlC, nil)
		req3.Header.Add("Authorization", "Bearer "+obj.Access_token)
		req3.Header.Add("Accept", "application/vnd.github+json")

		response3, err3 := client.Do(req3)

		if err3 != nil {
			fmt.Println("[-] Error in reading the output")
			fmt.Println(err3)
		}

		body, _ := io.ReadAll(response3.Body)
		json.Unmarshal(body, &output)
		fmt.Println("[+] Command Output : ", output[len(output)-1]["body"])

		//json.NewDecoder(response3.Body).Decode(&output)

		//fmt.Println("[+] Output : ", output[len(output)-1])

		fmt.Print("Continue (end/quit): ")
		fmt.Scan(&end)
	}
}

func main() {
	ForGithub()
}
