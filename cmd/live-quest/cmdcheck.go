package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/codigolandia/live-quest/log"
)

type ChallangeType string

const (
	ChallengeStatic ChallangeType = "static"
)

type ChallengeBackend string

const (
	ChallengeBackendPlayGoDev ChallengeBackend = "playGoDev"
)

type Challenge struct {
	Type    ChallangeType
	Backend ChallengeBackend

	Output string
}

type CheckResult struct {
	OK     bool
	Result string
}

type VetResponse struct {
	Body  string `json:"Body"`
	Error string `json:"Error"`
}

type CompileResponse struct {
	Errors string `json:"Errors"`
	Events []struct {
		Message string `json:"Message"`
		Kind    string `json:"Kind"`
		Delay   int    `json:"Delay"`
	} `json:"Events"`
	VetErrors string `json:"VetErrors"`
}

type CheckClient struct {
	hc http.Client
}

func NewCheckClient() *CheckClient {
	return &CheckClient{
		hc: http.Client{},
	}
}

func (c CheckClient) Call(method string, url string, body io.Reader) (out []byte, err error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "live-quest/v0.0.1-dev")
	if body != nil {
		req.Header.Set("content-type", "application/x-www-form-urlencoded")
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("check: error fetching %v: %v", url, resp.Status)
	}
	defer closeWithLog(resp.Body.Close)
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func Check(author string, shareUrl string, challenge Challenge) (cr *CheckResult, err error) {
	if challenge.Backend != ChallengeBackendPlayGoDev {
		return nil, fmt.Errorf("check: invalid backend: %v", challenge.Backend)
	}

	checkClient := NewCheckClient()
	codeSplit := strings.Split(shareUrl, "/")
	if len(codeSplit) < 1 {
		return nil, errors.New("invalid challenge shareUrl")
	}
	codeURI := fmt.Sprintf("https://go.dev/_/share?id=%s", codeSplit[len(codeSplit)-1])
	sourceCode, err := checkClient.Call(http.MethodGet, codeURI, nil)
	if err != nil {
		return nil, err
	}

	payload := url.Values{}
	payload.Set("body", string(sourceCode))
	payload.Set("imports", "true")
	vettedCode, err := checkClient.Call(
		http.MethodPost,
		"https://go.dev/_/fmt?backend=",
		strings.NewReader(payload.Encode()))
	if err != nil {
		return nil, err
	}

	var respVetted VetResponse
	err = json.Unmarshal(vettedCode, &respVetted)
	if err != nil {
		return nil, err
	}
	if respVetted.Error != "" {
		return nil, errors.New(respVetted.Error)
	}

	payload = url.Values{}
	payload.Set("body", respVetted.Body)
	payload.Set("withVet", "true")
	payload.Set("version", "2")

	compileOutput, err := checkClient.Call(
		http.MethodPost,
		"https://go.dev/_/compile?backend=",
		strings.NewReader(payload.Encode()))
	if err != nil {
		return nil, err
	}

	var compileResp CompileResponse
	err = json.Unmarshal(compileOutput, &compileResp)
	if err != nil {
		return nil, err
	}
	if len(compileResp.Errors) > 0 {
		return nil, errors.New(compileResp.Errors)
	}
	if len(compileResp.VetErrors) > 0 {
		return nil, errors.New(compileResp.VetErrors)
	}

	var output string
	for _, ev := range compileResp.Events {
		switch ev.Kind {
		case "stdout", "stderr":
			output = output + ev.Message
		default:
			log.W("check: unexpected event kind: %v", ev.Kind)
		}
	}

	log.D("check: got '%v'", string(output))
	cr = &CheckResult{}
	cr.OK = true
	cr.Result = "OK"
	if output != challenge.Output {
		cr.OK = false
		cr.Result = fmt.Sprintf("Unexpected output: %v", string(output))
	}
	return cr, nil
}

func closeWithLog(f func() error) {
	if err := f(); err != nil {
		log.W("Error when closing: %v", err)
	}
}
