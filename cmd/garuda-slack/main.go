// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

var (
	slackBotToken  = os.Getenv("SLACK_BOT_TOKEN")
	slackAppToken  = os.Getenv("SLACK_APP_TOKEN")
	garudaAPIURL   = os.Getenv("GARUDA_API_URL")
	garudaAPIToken = os.Getenv("GARUDA_API_TOKEN")
)

func main() {
	if slackBotToken == "" || slackAppToken == "" {
		log.Fatal("SLACK_BOT_TOKEN and SLACK_APP_TOKEN must be set")
	}
	if garudaAPIURL == "" {
		garudaAPIURL = "http://localhost:8080"
	}

	// Create Slack client
	client := slack.New(slackBotToken, slack.OptionAppLevelToken(slackAppToken))
	socketClient := socketmode.New(client,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "slack: ", log.Lshortfile)),
	)

	// Create Garuda API client
	garuda := &GarudaClient{
		BaseURL: garudaAPIURL,
		Token:   garudaAPIToken,
	}

	// Handle socket mode events
	go func() {
		for envelope := range socketClient.Events {
			// Ack the event
			socketClient.Ack(*envelope.Request)
		}
	}()

	// Handle slash commands
	http.HandleFunc("/slack/command", func(w http.ResponseWriter, r *http.Request) {
		// Parse the request from Slack
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		command := r.FormValue("command")
		text := r.FormValue("text")
		userID := r.FormValue("user_id")
		responseURL := r.FormValue("response_url")

		switch command {
		case "/decide":
			handleDecide(w, r, text, userID, garuda, responseURL)
		case "/check":
			handleCheck(w, r, text, garuda, responseURL)
		case "/approve":
			handleApprove(w, r, text, userID, garuda, responseURL)
		case "/supersede":
			handleSupersede(w, r, text, userID, garuda, responseURL)
		case "/refund":
			handleRefund(w, r, text, garuda, responseURL)
		default:
			http.Error(w, "Unknown command", http.StatusBadRequest)
		}
	})

	// Start the server
	go func() {
		log.Println("Slack bot server listening on :8082")
		if err := http.ListenAndServe(":8082", nil); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")
}

// GarudaClient is a simple client for the Garuda API.
type GarudaClient struct {
	BaseURL string
	Token   string
}

// CreateDecision creates a new decision in Garuda.
func (c *GarudaClient) CreateDecision(rec *DecisionRequest) (*DecisionResponse, error) {
	body, _ := json.Marshal(rec)
	req, _ := http.NewRequest("POST", c.BaseURL+"/v1/decisions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result DecisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}
	return &result, nil
}

// GetDecision retrieves a decision from Garuda.
func (c *GarudaClient) GetDecision(id string) (*DecisionResponse, error) {
	req, _ := http.NewRequest("GET", c.BaseURL+"/v1/decisions/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result DecisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TransitionDecision transitions a decision to a new status.
func (c *GarudaClient) TransitionDecision(id, status string) error {
	body, _ := json.Marshal(map[string]string{"status": status})
	req, _ := http.NewRequest("POST", c.BaseURL+"/v1/decisions/"+id+"/transition", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("API error: %d", resp.StatusCode)
	}
	return nil
}

// RefundRequest represents a refund request.
type RefundRequest struct {
	OrderID       string  `json:"order_id"`
	CustomerEmail string  `json:"customer_email"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Reason        string  `json:"reason"`
}

// RefundResponse represents a refund response.
type RefundResponse struct {
	Status     string `json:"status"`
	DecisionID string `json:"decision_id"`
	Reason     string `json:"reason"`
}

// ProcessRefund processes a refund via the refund service.
func (c *GarudaClient) ProcessRefund(req *RefundRequest) (*RefundResponse, error) {
	body, _ := json.Marshal(req)
	// Use the refund service on port 8081
	reqHTTP, _ := http.NewRequest("POST", "http://localhost:8081/v1/refund", bytes.NewReader(body))
	reqHTTP.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(reqHTTP)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result RefundResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DecisionRequest represents the request to create a decision.
type DecisionRequest struct {
	ID         string            `json:"id"`
	Decision   string            `json:"decision"`
	Rationale  string            `json:"rationale,omitempty"`
	Scope      map[string]string `json:"scope"`
	Owner      string            `json:"owner"`
	Approvers  []string          `json:"approvers"`
	Confidence float64           `json:"confidence"`
}

// DecisionResponse represents the response from Garuda.
type DecisionResponse struct {
	ID             string            `json:"id"`
	Decision       string            `json:"decision"`
	Status         string            `json:"status"`
	Scope          map[string]string `json:"scope"`
	Message        string            `json:"message,omitempty"`
	Contradictions []interface{}     `json:"contradictions,omitempty"`
}

// handleDecide handles the /decide command.
func handleDecide(w http.ResponseWriter, r *http.Request, text, userID string, garuda *GarudaClient, responseURL string) {
	// Parse the decision text
	if text == "" {
		sendSlackResponse(responseURL, "Usage: `/decide <decision text>`")
		return
	}

	// Create a simple decision
	decision := &DecisionRequest{
		ID:         "D-" + time.Now().Format("20060102150405"),
		Decision:   text,
		Scope:      map[string]string{"domain": "general"},
		Owner:      userID,
		Approvers:  []string{userID},
		Confidence: 0.8,
	}

	resp, err := garuda.CreateDecision(decision)
	if err != nil {
		sendSlackResponse(responseURL, "❌ Error: "+err.Error())
		return
	}

	if resp.Contradictions != nil && len(resp.Contradictions) > 0 {
		sendSlackResponse(responseURL, fmt.Sprintf("⚠️ Contradiction detected! Decision rejected.\n%v", resp.Contradictions))
		return
	}

	sendSlackResponse(responseURL, fmt.Sprintf("✅ Decision `%s` created with status: %s", resp.ID, resp.Status))
}

// handleCheck handles the /check command.
func handleCheck(w http.ResponseWriter, r *http.Request, text string, garuda *GarudaClient, responseURL string) {
	if text == "" {
		sendSlackResponse(responseURL, "Usage: `/check <decision ID>`")
		return
	}

	decision, err := garuda.GetDecision(text)
	if err != nil {
		sendSlackResponse(responseURL, "❌ Error: "+err.Error())
		return
	}

	sendSlackResponse(responseURL, fmt.Sprintf("📄 Decision `%s`\nStatus: %s\nDecision: %s", decision.ID, decision.Status, decision.Decision))
}

// handleApprove handles the /approve command.
func handleApprove(w http.ResponseWriter, r *http.Request, text, userID string, garuda *GarudaClient, responseURL string) {
	if text == "" {
		sendSlackResponse(responseURL, "Usage: `/approve <decision ID>`")
		return
	}

	if err := garuda.TransitionDecision(text, "APPROVED"); err != nil {
		sendSlackResponse(responseURL, "❌ Error: "+err.Error())
		return
	}
	sendSlackResponse(responseURL, fmt.Sprintf("✅ Decision `%s` approved.", text))
}

// handleSupersede handles the /supersede command.
func handleSupersede(w http.ResponseWriter, r *http.Request, text, userID string, garuda *GarudaClient, responseURL string) {
	// Implement supersede logic here
	sendSlackResponse(responseURL, "🔄 Supersede command not yet implemented")
}

// handleRefund handles the /refund command.
func handleRefund(w http.ResponseWriter, r *http.Request, text string, garuda *GarudaClient, responseURL string) {
	// Parse amount from text
	var amount float64
	_, err := fmt.Sscanf(text, "%f", &amount)
	if err != nil {
		sendSlackResponse(responseURL, "Usage: `/refund <amount>`")
		return
	}

	refundReq := &RefundRequest{
		OrderID:       "SLACK-" + time.Now().Format("20060102150405"),
		CustomerEmail: "slack-user@example.com",
		Amount:        amount,
		Currency:      "USD",
		Reason:        "slack-request",
	}

	resp, err := garuda.ProcessRefund(refundReq)
	if err != nil {
		sendSlackResponse(responseURL, "❌ Error: "+err.Error())
		return
	}

	if resp.Status == "approved" {
		sendSlackResponse(responseURL, fmt.Sprintf("✅ Refund of $%.2f approved (decision: %s)", amount, resp.DecisionID))
	} else {
		sendSlackResponse(responseURL, fmt.Sprintf("❌ Refund of $%.2f denied: %s", amount, resp.Reason))
	}
}

// sendSlackResponse sends a response to Slack via the response URL.
func sendSlackResponse(url, text string) {
	payload := map[string]string{"text": text}
	body, _ := json.Marshal(payload)
	http.Post(url, "application/json", bytes.NewReader(body))
}
