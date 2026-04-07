// @atomic route=post:notify-agent auth=none env=.env
// drift:trigger queue agent-queue poll=5000ms retry=3

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	drift "github.com/ondrift/drift-sdk"
)

type RequestBody struct {
	Type      string `json:"type"`
	InquiryID string `json:"inquiry_id"`
	RsvpID    string `json:"rsvp_id"`
	ListingID string `json:"listing_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Message   string `json:"message"`
	PartySize int    `json:"party_size"`
}

func PostNotifyAgent(req RequestBody) (int, string, interface{}) {
	apiKey := drift.Env("RESEND_API_KEY")
	if apiKey == "" {
		drift.Log(fmt.Sprintf("[notify-agent] no RESEND_API_KEY — skipping email for %s", req.Type))
		return http.StatusOK, "Skipped", map[string]string{"reason": "no email key configured"}
	}

	senderEmail := drift.Env("SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "listings@yourdomain.com"
	}
	agentName := drift.Env("AGENT_NAME")
	if agentName == "" {
		agentName = "Agent"
	}
	agentEmail := drift.Env("AGENT_EMAIL")
	if agentEmail == "" {
		drift.Log("[notify-agent] no AGENT_EMAIL configured — skipping")
		return http.StatusOK, "Skipped", map[string]string{"reason": "no agent email configured"}
	}

	var subject string
	var htmlBody string

	if req.Type == "inquiry" {
		subject = fmt.Sprintf("New inquiry on %s from %s", req.ListingID, req.Name)
		htmlBody = fmt.Sprintf(`
<p>Hi %s,</p>
<p>You have a new property inquiry.</p>
<table style="border-collapse:collapse;margin:1rem 0">
  <tr><td style="padding:4px 12px 4px 0;color:#666">Reference</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Listing</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">From</td><td><strong>%s</strong> (%s)</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Phone</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Message</td><td>%s</td></tr>
</table>
<p>Reply directly to this email to respond to the inquiry.</p>
`, agentName, req.InquiryID, req.ListingID, req.Name, req.Email, req.Phone, req.Message)
	} else if req.Type == "rsvp" {
		subject = fmt.Sprintf("Open house RSVP on %s from %s", req.ListingID, req.Name)
		htmlBody = fmt.Sprintf(`
<p>Hi %s,</p>
<p>Someone has RSVP'd for an open house.</p>
<table style="border-collapse:collapse;margin:1rem 0">
  <tr><td style="padding:4px 12px 4px 0;color:#666">Reference</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Listing</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Name</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Party size</td><td><strong>%d guest(s)</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Phone</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Email</td><td>%s</td></tr>
</table>
`, agentName, req.RsvpID, req.ListingID, req.Name, req.PartySize, req.Phone, req.Email)
	} else {
		return http.StatusBadRequest, "Bad Request", map[string]string{"error": "unknown notification type"}
	}

	payload := map[string]any{
		"from":     fmt.Sprintf("%s Listings <%s>", agentName, senderEmail),
		"to":       []string{agentEmail},
		"reply_to": req.Email,
		"subject":  subject,
		"html":     htmlBody,
	}

	body, _ := json.Marshal(payload)
	resp, err := drift.HTTPRequest(
		http.MethodPost,
		"https://api.resend.com/emails",
		map[string]string{
			"Authorization": "Bearer " + apiKey,
			"Content-Type":  "application/json",
		},
		body,
	)
	if err != nil {
		return http.StatusInternalServerError, "Email error", map[string]string{"error": err.Error()}
	}
	if resp.Status >= 400 {
		return http.StatusInternalServerError, "Email error", map[string]string{"error": fmt.Sprintf("resend returned %d", resp.Status)}
	}

	drift.Log(fmt.Sprintf("[notify-agent] %s notification sent to %s", req.Type, agentEmail))
	return http.StatusOK, "Sent", map[string]string{"type": req.Type, "agent": agentEmail}
}
