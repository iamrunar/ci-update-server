package main

//https://pkg.go.dev

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type TagDescription struct {
	Name string `json:"name"`
}

type ContainerMetadataDescription struct {
	Tag TagDescription `json:"tag"`
}

type PackageVersionDescription struct {
	PackageUrl        string                       `json:"package_url"`
	ContainerMetadata ContainerMetadataDescription `json:"container_metadata"`
}

type PackageDescription struct {
	PackageVersion PackageVersionDescription `json:"package_version"`
}

type ActionDescription struct {
	Action  string             `json:"action"`
	Package PackageDescription `json:"package"`
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Println("got / request")
	io.WriteString(w, "Root requested\n")
}

func postEventHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("got /event_handler request")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Could not read body %s\n", err)
		return
	}
	if body == nil {
		log.Printf("Body is empty")
	}

	githubSecretToken := os.Getenv("GITHUB_WEBHOOK_SECRET_TOKEN")
	log.Printf("GITHUB_WEBHOOK_SECRET_TOKEN: %s", githubSecretToken)
	signatureErr := verifySignature(body, r.Header.Get("X-Hub-Signature-256"), githubSecretToken)
	if signatureErr != nil {
		log.Printf("event_handler failed: %s", signatureErr)
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "Forbidden.")
		return
	}

	processXGitHubEvent(strings.ToLower(r.Header.Get("X-GitHub-Event")), body)

	io.WriteString(w, "OK")
}

func verifySignature(body []byte, githubSignature256 string, githubSecretToken string) error {
	//https://pkg.go.dev/crypto/hmac#example-New
	sha256Text := fmt.Sprintf("sha256=%x", hmacSHA256(githubSecretToken, body))

	// https://codahale.com/a-lesson-in-timing-attacks/
	compare := subtle.ConstantTimeCompare([]byte(sha256Text), []byte(githubSignature256))
	if compare == 0 {
		return errors.New("Signatures are not equal")
	}
	return nil
}

func hmacSHA256(secretToken string, body []byte) []byte {
	hash := hmac.New(sha256.New, []byte(secretToken))
	hash.Write(body)
	return hash.Sum(nil)
}

func executeDockerCommand(arg ...string) error {
	log.Printf("Execute command: docker %q\n", arg)
	cmd := exec.Command("docker", arg...)

	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out

	cmdErr := cmd.Run()
	if cmdErr != nil {
		return errors.New(fmt.Sprintf("Error executing command: %s\nOutput: %s\n", cmdErr, out.String()))
	}
	return nil
}

func processXGitHubEvent(githubEventTypeText string, body []byte) {
	switch githubEventTypeText {
	default:
		log.Printf("Unknown X-GitHub-Event = %s", githubEventTypeText)
	case "ping":
		log.Println("Process ping event% N/A")
	case "package":
		log.Println("Process package event...")
		actionDescription, err := convertBodyToArgs(body)
		if err != nil {
			return
		}
		processXGitHubPackageEvent(actionDescription)
	}
}

func updateDockerContainerAndRerun(packageUrl string) {
	executingCommandsSequence := [][]string{
		{"pull", packageUrl},
		{"stop", "video-converter-back"},
		{"rm", "video-converter-back"},
		{"run", "-d", "--name", "video-converter-back", "-p", "5104:80", packageUrl},
	}
	for _, command := range executingCommandsSequence {
		err := executeDockerCommand(command...)
		if err != nil {
			//fixme: ignore error No such container: video-converter-back (docker ["stop" "video-converter-back"])
			log.Printf("Warning! command docker %q\n return ignorable(?) error: %s", command, err)
		}
	}
	log.Printf("Update docker container complited")
}

func processXGitHubPackageEvent(actionDescription *ActionDescription) {

	log.Printf("Receive package: %s\n", actionDescription)

	packageUrl := actionDescription.Package.PackageVersion.PackageUrl
	tagName := actionDescription.Package.PackageVersion.ContainerMetadata.Tag.Name
	log.Printf("package url: %s\ntag name: %s\n", packageUrl, tagName)
	if tagName == "" {
		log.Printf("Tag name is empty\n")
		return
	}

	go updateDockerContainerAndRerun(packageUrl)
}

func convertBodyToArgs(body []byte) (*ActionDescription, error) {
	var actionDescription *ActionDescription
	err := json.Unmarshal(body, &actionDescription)
	if err != nil {
		return nil, err
	}
	return actionDescription, nil
}

func main() {
	listenAddress := ":27075"
	log.Printf("Hello! I'm Github webhook server. Listen to %s", listenAddress)

	http.HandleFunc("/", getRoot)
	http.HandleFunc("/event_handler", postEventHandler)

	err := http.ListenAndServe(listenAddress, nil)
	if errors.Is(err, http.ErrServerClosed) {
		log.Print("Server closed")
	} else if err != nil {
		log.Fatal(fmt.Sprintf("Err %s", err))
	}
}
