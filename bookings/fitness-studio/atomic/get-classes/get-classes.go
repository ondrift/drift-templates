// @atomic route=get:get-classes auth=none env=.env
// Returns the class schedule with live availability.
// The schedule is stored in Backbone Cache under key "class-schedule" as a JSON string.
// Per-class spot counts are tracked in cache keys "class:{id}:{date}:count".

package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	drift "github.com/ondrift/drift-sdk"
)

type classInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Instructor  string   `json:"instructor"`
	Time        string   `json:"time"`
	Duration    string   `json:"duration"`
	MaxCapacity int      `json:"max_capacity"`
	Location    string   `json:"location"`
	Days        []string `json:"days"`
}

type schedule struct {
	Classes []classInfo `json:"classes"`
}

func GetGetClasses() (int, string, interface{}) {
	raw, err := drift.Cache.Get("class-schedule")
	if err != nil || len(raw) == 0 {
		return http.StatusOK, "OK", defaultSchedule()
	}

	var sched schedule
	if err := json.Unmarshal(raw, &sched); err != nil {
		return http.StatusOK, "OK", defaultSchedule()
	}

	if len(sched.Classes) == 0 {
		return http.StatusOK, "OK", defaultSchedule()
	}

	// Build response with availability for the next 7 days.
	today := time.Now()
	type classAvail struct {
		classInfo
		Upcoming []map[string]any `json:"upcoming"`
	}

	var result []classAvail
	for _, c := range sched.Classes {
		ca := classAvail{classInfo: c}
		daySet := map[string]bool{}
		for _, d := range c.Days {
			daySet[d] = true
		}

		for offset := 0; offset < 7; offset++ {
			day := today.AddDate(0, 0, offset)
			if !daySet[day.Weekday().String()] {
				continue
			}
			dateStr := day.Format("2006-01-02")
			spotsUsed := 0
			countRaw, err := drift.Cache.Get("class:" + c.ID + ":" + dateStr + ":count")
			if err == nil && len(countRaw) > 0 {
				spotsUsed, _ = strconv.Atoi(string(countRaw))
			}
			remaining := c.MaxCapacity - spotsUsed
			if remaining < 0 {
				remaining = 0
			}
			ca.Upcoming = append(ca.Upcoming, map[string]any{
				"date":            dateStr,
				"day":             day.Weekday().String(),
				"spots_remaining": remaining,
			})
		}
		result = append(result, ca)
	}

	return http.StatusOK, "OK", map[string]any{"classes": result}
}

func defaultSchedule() map[string]any {
	return map[string]any{
		"classes": []map[string]any{
			{"id": "yoga-morning", "name": "Morning Yoga", "instructor": "Sarah", "time": "07:00", "duration": "60min", "max_capacity": 20, "location": "Studio A", "days": []string{"Monday", "Wednesday", "Friday"}, "upcoming": []any{}},
			{"id": "hiit-express", "name": "HIIT Express", "instructor": "Marcus", "time": "12:00", "duration": "30min", "max_capacity": 15, "location": "Studio B", "days": []string{"Tuesday", "Thursday"}, "upcoming": []any{}},
			{"id": "spin-evening", "name": "Evening Spin", "instructor": "Jade", "time": "18:00", "duration": "45min", "max_capacity": 25, "location": "Spin Room", "days": []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}, "upcoming": []any{}},
			{"id": "pilates-core", "name": "Pilates Core", "instructor": "Emma", "time": "09:30", "duration": "50min", "max_capacity": 12, "location": "Studio A", "days": []string{"Tuesday", "Saturday"}, "upcoming": []any{}},
			{"id": "boxing-basics", "name": "Boxing Basics", "instructor": "Marcus", "time": "17:00", "duration": "60min", "max_capacity": 16, "location": "Studio B", "days": []string{"Wednesday", "Friday"}, "upcoming": []any{}},
		},
	}
}
