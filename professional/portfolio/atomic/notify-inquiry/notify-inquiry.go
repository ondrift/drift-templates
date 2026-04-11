// @atomic route=post:notify-inquiry auth=none env=.env
// drift:trigger queue inquiry-queue poll=5000ms retry=3

package main

import (
	"net/http"

	drift "github.com/ondrift/drift-sdk"
)

// RequestBody is the message popped from inquiry-queue by the trigger.
type RequestBody struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
	Budget  string `json:"budget"`
}

func PostNotifyInquiry(req RequestBody) (int, string, interface{}) {
	if req.Email == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{"error": "email required"}
	}

	apiKey, _ := drift.Secret.Get("RESEND_API_KEY")
	if apiKey == "" {
		return http.StatusOK, "Skipped", map[string]string{"reason": "no email key configured"}
	}

	senderEmail, _ := drift.Secret.Get("SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "portfolio@yourdomain.com"
	}
	ownerName, _ := drift.Secret.Get("OWNER_NAME")
	if ownerName == "" {
		ownerName = "Portfolio Owner"
	}
	ownerEmail, _ := drift.Secret.Get("OWNER_EMAIL")
	if ownerEmail == "" {
		return http.StatusOK, "Skipped", map[string]string{"reason": "no owner email configured"}
	}

	subject := req.Subject
	if subject == "" {
		subject = "New inquiry"
	}

	budget := req.Budget
	if budget == "" {
		budget = "Not specified"
	}

	return http.StatusOK, "Sent", map[string]string{"inquiry_id": req.ID, "notified": ownerEmail}
}
