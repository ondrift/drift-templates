// @atomic route=post:notify-department auth=none env=.env
// drift:trigger queue department-queue poll=5000ms retry=3

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"drift-sdk"
)

// RequestBody is the message popped from department-queue by the trigger.
type RequestBody struct {
	Ticket      string `json:"ticket"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Category    string `json:"category"`
	Location    string `json:"location"`
	Description string `json:"description"`
}

func PostNotifyDepartment(req RequestBody) (int, string, interface{}) {
	if req.Ticket == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{"error": "ticket required"}
	}

	apiKey := drift.Env("RESEND_API_KEY")
	if apiKey == "" {
		drift.Log(fmt.Sprintf("[notify-department] no RESEND_API_KEY — skipping email for ticket %s", req.Ticket))
		return http.StatusOK, "Skipped", map[string]string{"reason": "no email key configured"}
	}

	senderEmail := drift.Env("SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "noreply@yourdomain.com"
	}
	departmentEmail := drift.Env("DEPARTMENT_EMAIL")
	if departmentEmail == "" {
		drift.Log(fmt.Sprintf("[notify-department] no DEPARTMENT_EMAIL — skipping email for ticket %s", req.Ticket))
		return http.StatusOK, "Skipped", map[string]string{"reason": "no department email configured"}
	}
	cityName := drift.Env("CITY_NAME")
	if cityName == "" {
		cityName = "City Services"
	}

	location := req.Location
	if location == "" {
		location = "Not specified"
	}

	payload := map[string]any{
		"from":    fmt.Sprintf("%s <%s>", cityName, senderEmail),
		"to":      []string{departmentEmail},
		"subject": fmt.Sprintf("New service request %s — %s", req.Ticket, req.Category),
		"html": fmt.Sprintf(`
<p>A new service request has been submitted.</p>
<table style="border-collapse:collapse;margin:1rem 0">
  <tr><td style="padding:4px 12px 4px 0;color:#666">Ticket</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Category</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Location</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Description</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Reporter</td><td>%s (%s)</td></tr>
</table>
<p>Please review and assign this request to the appropriate team.</p>
<p>– %s Automated Services</p>
`, req.Ticket, req.Category, location, req.Description, req.Name, req.Email, cityName),
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

	drift.Log(fmt.Sprintf("[notify-department] notification sent for ticket %s (%s)", req.Ticket, req.Category))
	return http.StatusOK, "Sent", map[string]string{"ticket": req.Ticket, "email": departmentEmail}
}
