// @atomic route=post:notify-hiring auth=none env=.env
// drift:trigger queue hiring-queue poll=5000ms retry=3

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"drift-sdk"
)

// RequestBody is the message popped from hiring-queue by the trigger.
type RequestBody struct {
	ApplicationID string `json:"application_id"`
	PositionID    string `json:"position_id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	CoverLetter   string `json:"cover_letter"`
	LinkedInURL   string `json:"linkedin_url"`
}

func PostNotifyHiring(req RequestBody) (int, string, interface{}) {
	if req.Email == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{"error": "email required"}
	}

	apiKey := drift.Env("RESEND_API_KEY")
	if apiKey == "" {
		drift.Log(fmt.Sprintf("[notify-hiring] no RESEND_API_KEY — skipping email for application %s", req.ApplicationID))
		return http.StatusOK, "Skipped", map[string]string{"reason": "no email key configured"}
	}

	senderEmail := drift.Env("SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "careers@yourdomain.com"
	}
	companyName := drift.Env("COMPANY_NAME")
	if companyName == "" {
		companyName = "Our Company"
	}
	hiringEmail := drift.Env("HIRING_EMAIL")
	if hiringEmail == "" {
		drift.Log(fmt.Sprintf("[notify-hiring] no HIRING_EMAIL — skipping notification for application %s", req.ApplicationID))
		return http.StatusOK, "Skipped", map[string]string{"reason": "no hiring email configured"}
	}

	phoneRow := ""
	if req.Phone != "" {
		phoneRow = fmt.Sprintf(`<tr><td style="padding:4px 12px 4px 0;color:#666">Phone</td><td><strong>%s</strong></td></tr>`, req.Phone)
	}
	coverLetterRow := ""
	if req.CoverLetter != "" {
		coverLetterRow = fmt.Sprintf(`<tr><td style="padding:4px 12px 4px 0;color:#666;vertical-align:top">Cover Letter</td><td>%s</td></tr>`, req.CoverLetter)
	}
	linkedInRow := ""
	if req.LinkedInURL != "" {
		linkedInRow = fmt.Sprintf(`<tr><td style="padding:4px 12px 4px 0;color:#666">LinkedIn</td><td><a href="%s">%s</a></td></tr>`, req.LinkedInURL, req.LinkedInURL)
	}

	payload := map[string]any{
		"from":     fmt.Sprintf("%s Careers <%s>", companyName, senderEmail),
		"to":       []string{hiringEmail},
		"reply_to": req.Email,
		"subject":  fmt.Sprintf("New application — %s for %s", req.Name, req.PositionID),
		"html": fmt.Sprintf(`
<p>New application received at <strong>%s</strong>.</p>
<table style="border-collapse:collapse;margin:1rem 0">
  <tr><td style="padding:4px 12px 4px 0;color:#666">Application ID</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Position</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Name</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Email</td><td><a href="mailto:%s">%s</a></td></tr>
  %s
  %s
  %s
</table>
<p>Reply directly to this email to contact the applicant.</p>
<p>– %s Careers</p>
`, companyName, req.ApplicationID, req.PositionID, req.Name, req.Email, req.Email, phoneRow, coverLetterRow, linkedInRow, companyName),
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

	drift.Log(fmt.Sprintf("[notify-hiring] notification sent to %s for application %s", hiringEmail, req.ApplicationID))
	return http.StatusOK, "Sent", map[string]string{"email": hiringEmail, "application_id": req.ApplicationID}
}
