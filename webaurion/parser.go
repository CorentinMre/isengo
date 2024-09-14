package webaurion

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	
	"github.com/PuerkitoBio/goquery"
)

type BeautifulGrade struct{}
type BeautifulAbsences struct{}
type BeautifulPlanning struct{}

func (b *BeautifulGrade) ParseGrades(html []byte) (*GradeReport, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		return nil, err
	}

	title := doc.Find("title").Text()
	if strings.TrimSpace(title) != "Mes notes" {
		return nil, fmt.Errorf("not connected to WebAurion")
	}

	var grades []Grade
	var totalGrade float64
	var gradeCount int

	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td").Map(func(_ int, s *goquery.Selection) string {
			return s.Text()
		})

		if len(tds) < 7 || tds[0] == "Aucun enregistrement" {
			return
		}

		gradeValue, err := strconv.ParseFloat(strings.Replace(tds[3], ",", ".", -1), 64)
		if err != nil {
			gradeValue = 0
		}

		absence := false
		if tds[4] == "Oui" {
			absence = true
		}

		grade := Grade{
			Date:         tds[0],
			Code:         tds[1],
			Name:         tds[2],
			Grade:        gradeValue,
			Absence:      absence,
			Appreciation: tds[5],
			Instructors:  strings.Split(tds[6], "/"),
		}

		grades = append(grades, grade)

		if !absence {
			totalGrade += gradeValue
			gradeCount++
		}
	})

	average := 0.0
	if gradeCount > 0 {
		average = totalGrade / float64(gradeCount)
	}

	gradeReport := &GradeReport{
		Average: average,
		Grades:  grades,
	}

	return gradeReport, nil
}

func (b *BeautifulAbsences) ParseAbsences(html []byte) (*AbsenceReport, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		return nil, err
	}

	// check for security hehe
	title := doc.Find("title").Text()

	// fmt.Println(title)


	if strings.TrimSpace(title) != "Mes absences" {
		return nil, fmt.Errorf("not connected to WebAurion or not on the absences page")
	}

	var absences []Absence
	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td").Map(func(_ int, s *goquery.Selection) string {
			return strings.TrimSpace(s.Text())
		})

		// skip unwanted rows
		if len(tds) < 7 || tds[0] == "Date" {
			return
		}

		absence := Absence{
			Date:       tds[0],
			Reason:     tds[1],
			Duration:   tds[2],
			Schedule:   tds[3],
			Course:     tds[4],
			Instructor: tds[5],
			Subject:    tds[6],
		}
		absences = append(absences, absence)
	})

	// all periods of absence
	duration := 0
	for _, absence := range absences {
		// split duration into hours and minutes
		durationParts := strings.Split(absence.Duration, ":")
		if len(durationParts) != 2 {
			continue
		}

		hours, err := strconv.Atoi(durationParts[0])
		if err != nil {
			continue
		}

		minutes, err := strconv.Atoi(durationParts[1])
		if err != nil {
			continue
		}

		duration += hours*60 + minutes
	}

	absenceReport := &AbsenceReport{
		NbAbsences: len(absences),
		Duration:       duration,
		Data:       absences,
	}

	return absenceReport, nil
}

func (b *BeautifulPlanning) ParsePlanning(html []byte) (*PlanningReport, error) {
	// delete CDATA tags
	htmlString := strings.ReplaceAll(string(html), `<![CDATA[`, "")
	htmlString = strings.ReplaceAll(htmlString, `]]>`, "")
	
	// laod the HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlString))
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %v", err)
	}

	// extract the JSON data
	jsonData := doc.Find("#form\\:j_idt118").Text()
	if jsonData == "" {
		return nil, fmt.Errorf("no JSON data found, you may not be connected to WebAurion")
	}

	// parse the JSON data
	var rawData struct {
		Events []map[string]interface{} `json:"events"`
	}

	// fmt.Println(jsonData)

	if err := json.Unmarshal([]byte(jsonData), &rawData); err != nil {
		return nil, fmt.Errorf("error parsing planning JSON: %v", err)
	}

	// for all evenements, create an Event object
	var events []Event
	for _, rawEvent := range rawData.Events {
		event, err := NewEvent(rawEvent)
		if err != nil {
			return nil, fmt.Errorf("error creating event: %v", err)
		}
		events = append(events, *event)
	}

	//return the PlanningReport
	return NewPlanningReport(events), nil
}