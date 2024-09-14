package webaurion

import (
	"encoding/json"
	"time"
	"fmt"
	"strings"
)


type UserInfo struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Name      string `json:"name"`
	Email     string `json:"email"`
}

// NewUserInfo creates a new instance of UserInfo.
func NewUserInfo(firstName, lastName, name, email string) *UserInfo {
	return &UserInfo{
		FirstName: firstName,
		LastName:  lastName,
		Name:      name,
		Email:     email,
	}
}

// JSON returns a JSON representation of the UserInfo.
func (ui *UserInfo) JSON() string {
	data, err := json.MarshalIndent(ui, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling to JSON: %v", err)
	}
	return string(data)
}


// Grade  date, code, name, grade, absence, appreciation, and instructors.
type Grade struct {
	Date         string   `json:"date"`
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	Grade        float64  `json:"grade"`
	Absence      bool     `json:"absence"`
	Appreciation string   `json:"appreciation"`
	Instructors  []string `json:"instructors"`
}

// create a new instance of Grade.
func NewGrade(date, code, name string, grade float64, absence bool, appreciation string, instructors []string) *Grade {
	return &Grade{
		Date:         date,
		Code:         code,
		Name:         name,
		Grade:        grade,
		Absence:      absence,
		Appreciation: appreciation,
		Instructors:  instructors,
	}
}

// Return a string representation of the Grade.
func (g *Grade) String() string {
	return fmt.Sprintf("Grade(date='%s', code='%s', name='%s', grade=%f, absence=%t, appreciation='%s', instructors=%v)",
		g.Date, g.Code, g.Name, g.Grade, g.Absence, g.Appreciation, g.Instructors)
}

// Get gets the value of a specific key for the Grade.
func (g *Grade) Get(key string) (interface{}, error) {
	switch key {
	case "date":
		return g.Date, nil
	case "code":
		return g.Code, nil
	case "name":
		return g.Name, nil
	case "grade":
		return g.Grade, nil
	case "absence":
		return g.Absence, nil
	case "appreciation":
		return g.Appreciation, nil
	case "instructors":
		return g.Instructors, nil
	default:
		return nil, fmt.Errorf("invalid key: %s, valid keys are 'date', 'code', 'name', 'grade', 'absence', 'appreciation', and 'instructors'", key)
	}
}

// GradeReport represents a report about grades, including the average and data about each grade.
type GradeReport struct {
	Average float64 `json:"average"`
	Grades  []Grade `json:"data"`
}

// NewGradeReport creates a new instance of GradeReport.
func NewGradeReport(average float64, grades []Grade) *GradeReport {
	return &GradeReport{
		Average: average,
		Grades:  grades,
	}
}

// String returns a string representation of the GradeReport.
func (gr *GradeReport) String() string {
	return fmt.Sprintf("GradeReport(average=%f, data=%v)", gr.Average, gr.Grades)
}

// Get gets the value of a specific key for the GradeReport.
func (gr *GradeReport) Get(key string) (interface{}, error) {
	switch key {
	case "average":
		return gr.Average, nil
	case "data":
		return gr.Grades, nil
	default:
		return nil, fmt.Errorf("invalid key: %s, valid keys are 'average' and 'data'", key)
	}
}

func (gr *GradeReport) JSON() string {
	data, err := json.MarshalIndent(gr, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling to JSON: %v", err)
	}
	return string(data)
}

// Absence represents an absence with details including date, reason, duration, schedule, course, instructor, and subject.
type Absence struct {
	Date       string `json:"date"`
	Reason     string `json:"reason"`
	Duration   string `json:"duration"`
	Schedule   string `json:"schedule"`
	Course     string `json:"course"`
	Instructor string `json:"instructor"`
	Subject    string `json:"subject"`
}

// NewAbsence creates a new instance of Absence.
func NewAbsence(date, reason, duration, schedule, course, instructor, subject string) *Absence {
	return &Absence{
		Date:       date,
		Reason:     reason,
		Duration:   duration,
		Schedule:   schedule,
		Course:     course,
		Instructor: instructor,
		Subject:    subject,
	}
}

// String returns a string representation of the Absence.
func (a *Absence) String() string {
	return fmt.Sprintf("Absence(date='%s', reason='%s', duration='%s', schedule='%s', course='%s', instructor='%s', subject='%s')",
		a.Date, a.Reason, a.Duration, a.Schedule, a.Course, a.Instructor, a.Subject)
}

// Get gets the value of a specific key for the Absence.
func (a *Absence) Get(key string) (string, error) {
	switch key {
	case "date":
		return a.Date, nil
	case "reason":
		return a.Reason, nil
	case "duration":
		return a.Duration, nil
	case "schedule":
		return a.Schedule, nil
	case "course":
		return a.Course, nil
	case "instructor":
		return a.Instructor, nil
	case "subject":
		return a.Subject, nil
	default:
		return "", fmt.Errorf("invalid key: %s, valid keys are 'date', 'reason', 'duration', 'schedule', 'course', 'instructor', and 'subject'", key)
	}
}

// AbsenceReport represents a report about absences, including the number of absences, time, and data about each absence.
type AbsenceReport struct {
	NbAbsences int       `json:"nbAbsences"`
	Duration       int    `json:"duration"`
	Data       []Absence `json:"data"`
}

// NewAbsenceReport creates a new instance of AbsenceReport.
func NewAbsenceReport(nbAbsences int, duration int, data []Absence) *AbsenceReport {
	return &AbsenceReport{
		NbAbsences: nbAbsences,
		Duration:       duration,
		Data:       data,
	}
}

// String returns a string representation of the AbsenceReport.
func (ar *AbsenceReport) String() string {
	return fmt.Sprintf("AbsenceReport(nbAbsences=%d, duration=%d, data=%v)", ar.NbAbsences, ar.Duration, ar.Data)
}

// Get gets the value of a specific key for the AbsenceReport.
func (ar *AbsenceReport) Get(key string) (interface{}, error) {
	switch key {
	case "nbAbsences":
		return ar.NbAbsences, nil
	case "duration":
		return ar.Duration, nil
	case "data":
		return ar.Data, nil
	default:
		return nil, fmt.Errorf("invalid key: %s, valid keys are 'nbAbsences', 'time', and 'data'", key)
	}
}

func (ar *AbsenceReport) JSON() string {
	data, err := json.MarshalIndent(ar, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling to JSON: %v", err)
	}
	return string(data)
}

type Event struct {
    ID        string    `json:"id"`
    Start     time.Time `json:"start"`
    End       time.Time `json:"end"`
    AllDay    bool      `json:"allDay"`
    ClassName string    `json:"className"`
    Details   Details   `json:"details"`
}

// Details represents the details of an event.
type Details struct {
    Time         string   `json:"time"`
    Room         string   `json:"room"`
    Type         string   `json:"type"`
    Subject      string   `json:"subject"`
    Instructors  []string `json:"instructors"`
    ClassGroups  []string `json:"classGroups"`
}

// parseEventTitle parses the event title into Details.
func parseEventTitle(title string) (Details, error) {
    // Diviser le titre en parties par " - "
    parts := strings.Split(title, " - ")

    // V�rifier la longueur minimale
    if len(parts) < 6 {
        return Details{}, fmt.Errorf("incomplete event title: %v", parts)
    }

    // Extraire et nettoyer les diff�rentes parties du titre
    time := strings.TrimSpace(parts[0])
    room := strings.TrimSpace(parts[1])
    eventType := strings.TrimSpace(parts[2])
    subject := strings.TrimSpace(parts[3])
    instructors := strings.Split(strings.TrimSpace(parts[4]), "/")

    // G�rer les groupes de classes s'ils sont disponibles
    var classGroups []string
    if len(parts) > 5 {
        classGroups = strings.Split(strings.TrimSpace(parts[5]), "/")
    }

    // Si les parties essentielles sont manquantes
    if time == "" || eventType == "" || subject == "" {
        return Details{}, fmt.Errorf("incomplete details extracted from title: %v", title)
    }

    return Details{
        Time:        time,
        Room:        room,
        Type:        eventType,
        Subject:     subject,
        Instructors: instructors,
        ClassGroups: classGroups,
    }, nil
}


func NewEvent(data map[string]interface{}) (*Event, error) {
    title, ok := data["title"].(string)
    if !ok {
        return nil, fmt.Errorf("invalid or missing 'title' field")
    }

    details, err := parseEventTitle(title)
    if err != nil {
        return nil, fmt.Errorf("error parsing event title: %v", err)
    }

    startStr, ok := data["start"].(string)
    if !ok {
        return nil, fmt.Errorf("invalid or missing 'start' field")
    }
    start, err := time.Parse("2006-01-02T15:04:05-0700", startStr)
    if err != nil {
        return nil, fmt.Errorf("error parsing start time: %v", err)
    }

    endStr, ok := data["end"].(string)
    if !ok {
        return nil, fmt.Errorf("invalid or missing 'end' field")
    }
    end, err := time.Parse("2006-01-02T15:04:05-0700", endStr)
    if err != nil {
        return nil, fmt.Errorf("error parsing end time: %v", err)
    }

    allDayStr, ok := data["allDay"].(bool)
    if !ok {
		allDayStr = false
        // allDayStr, ok = data["allDay"].(string)
        // if !ok {
        //     return nil, fmt.Errorf("invalid or missing 'allDay' field")
        // }
        // allDay, err := strconv.ParseBool(allDayStr)
        // if err != nil {
        //     return nil, fmt.Errorf("error parsing allDay field: %v", err)
        // }
        // allDayStr = allDay
    }

    className, ok := data["className"].(string)
    if !ok {
        return nil, fmt.Errorf("invalid or missing 'className' field")
    }

    return &Event{
        ID:          data["id"].(string),
        Start:       start,
        End:         end,
        AllDay:      allDayStr,
        ClassName:   className,
        Details:     details,
    }, nil
}


// String returns a string representation of the Event.
func (e *Event) String() string {
    return fmt.Sprintf("Event(id='%s', start='%s', end='%s', all_day='%t', class_name='%s', time='%s', room='%s', type='%s', subject='%s', instructors='%v', class_groups='%v')",
        e.ID, e.Start, e.End, e.AllDay, e.ClassName, e.Details.Time, e.Details.Room, e.Details.Type, e.Details.Subject, e.Details.Instructors, e.Details.ClassGroups)
}

// MarshalJSON implements the json.Marshaler interface for Event.
func (e *Event) MarshalJSON() ([]byte, error) {
    return json.Marshal(struct {
        ID        string    `json:"id"`
        Start     time.Time `json:"start"`
        End       time.Time `json:"end"`
        AllDay    bool      `json:"allDay"`
        ClassName string    `json:"className"`
        Details   Details   `json:"details"`
    }{
        ID:        e.ID,
        Start:     e.Start,
        End:       e.End,
        AllDay:    e.AllDay,
        ClassName: e.ClassName,
        Details:   e.Details,
    })
}

// PlanningReport represents a report about planning events.
type PlanningReport struct {
    Events []Event `json:"events"`
}

// NewPlanningReport creates a new instance of PlanningReport.
func NewPlanningReport(events []Event) *PlanningReport {
    return &PlanningReport{
        Events: events,
    }
}

// String returns a string representation of the PlanningReport.
func (pr *PlanningReport) String() string {
    return fmt.Sprintf("PlanningReport(events=%v)", pr.Events)
}

func (pr *PlanningReport) JSON() string {
	data, err := json.MarshalIndent(pr, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling to JSON: %v", err)
	}
	return string(data)
}
