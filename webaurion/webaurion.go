package webaurion

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type WebAurion struct {
	BaseURL    string
	Cookies    []*http.Cookie
	ViewState  string
	Link       map[string]string
	TimeConn   time.Time
	Name       string
	LoggedIn   bool
	Client     *http.Client
	GradeLink  string
	AbsenceLink string
	PlanningLink string
	IdInit	 string
	IdBasic string
	Payload string
}

func NewWebAurion() *WebAurion {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}
	return &WebAurion{
		BaseURL:   "https://web.isen-ouest.fr",
		Link:      make(map[string]string),
		LoggedIn:  false,
		Client:    client,
	}
}

func (w *WebAurion) Login(username, password string) (bool, error) {
	payload := url.Values{}
	payload.Set("username", username)
	payload.Set("password", password)
	payload.Set("j_idt27", "")


	req, err := http.NewRequest("POST", w.BaseURL+"/webAurion/login", strings.NewReader(payload.Encode()))
	if err != nil {
		return false, errors.New("error getting login page")
	}

	w.setRequestHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := w.Client.Do(req)
	
	// fmt.Printf("%v\n", resp)

	if err != nil {
		return false, errors.New("error getting main page")
	}
	defer resp.Body.Close()

	// if resp.StatusCode != http.StatusFound {
	// 	return false, fmt.Errorf("login failed: unexpected status code %d", resp.StatusCode)
	// }

	w.Cookies = resp.Cookies()

	// get the main page
	req, err = http.NewRequest("GET", w.BaseURL+"/webAurion/", nil)
	if err != nil {
		return false, errors.New("error getting main page")
	}
	w.setRequestHeaders(req)

	resp, err = w.Client.Do(req)
	if err != nil {
		return false, errors.New("error getting main page")
	}
	defer resp.Body.Close()

	w.ViewState, err = w.getViewState(resp.Body, true)
	// w.ViewState = data[0]
	// w.GradeLink = ui-menu-child
	// w.AbsenceLink = ui-menu-child
	// w.PlanningLink = ui-menu-child

	if err != nil {
		return false, errors.New("username or password incorrect")
	}


	// fmt.Print(w.GradeLink, w.AbsenceLink, w.PlanningLink, "\n")

	// fmt.Println(w.IdInit)

	w.LoggedIn = true
	w.TimeConn = time.Now()
	return true, nil
}

func (w *WebAurion) RemoveAccents(str string) string {
	// todo better
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u",
		"à", "a", "è", "e", "ì", "i", "ò", "o", "ù", "u",
		"â", "a", "ê", "e", "î", "i", "ô", "o", "û", "u",
		"ä", "a", "ë", "e", "ï", "i", "ö", "o", "ü", "u",
	)
	return replacer.Replace(str)
}

func (w *WebAurion) DoRequest(payload string, referer ...string) ([]byte, error) {

	targetURL := w.BaseURL + "/webAurion/faces/MainMenuPage.xhtml"
	if len(referer) > 0 && referer[0] != "" {
		targetURL = w.BaseURL + referer[0]
	}


	req, err := http.NewRequest("POST", targetURL, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}

	w.setRequestHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := w.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}





func (w *WebAurion) getViewState(body io.Reader, first bool) (string, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return "", err
	}

	if first {
		w.Name = doc.Find("div.menuMonCompte h3").Text()
		doc.Find("a.lien-cliquable").Each(func(i int, s *goquery.Selection) {
			id, _ := s.Attr("id")
			// fmt.Print(id, "\n")
			// fmt.Print(s.Text(), "\n")
			if strings.Contains(s.Text(), "note") {
				w.GradeLink = id//strings.Replace(id, ":", "%%3A", -1)
			} else if strings.Contains(s.Text(), "Absences") {
				w.AbsenceLink = id//strings.Replace(id, ":", "%%3A", -1)
			} else if strings.Contains(s.Text(), "Planning") {
				w.PlanningLink = id//strings.Replace(id, ":", "%%3A", -1)
			}
		})

		// for all input autocomplete="off" get the id and value
		doc.Find("input").Each(func(i int, s *goquery.Selection) {
			name, _ := s.Attr("name")
			value, _ := s.Attr("value")
			w.Payload += fmt.Sprintf("%s=%s&", name, value)
		})

		// fmt.Print(w.Payload, "\n")

		w.IdBasic, _ = doc.Find("input[value='basicDay']").Attr("id")

		// idInit, _ := doc.Find("input[name='form:idInit']").Attr("value")

		w.Payload += "form:j_idt820_input=275805"
		// w.Payload = strings.ReplaceAll(w.Payload, ":", "%%3A")
		

		// fmt.Print(w.Payload, "\n")

	}

	viewState, exists := doc.Find("input[name='javax.faces.ViewState']").Attr("value")
	if !exists {
		return "", fmt.Errorf("ViewState not found")
	}

	return viewState, nil
}

func (w *WebAurion) setRequestHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Sec-Ch-Ua", "\"Chromium\";v=\"124\", \"Google Chrome\";v=\"124\", \"Not-A.Brand\";v=\"99\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "\"macOS\"")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Referer", "https://web.isen-ouest.fr/webAurion/")
}

func (w *WebAurion) GetGradesPayload() string {

	
	return fmt.Sprintf("%s&%s=%s", w.Payload, w.GradeLink, w.GradeLink)
}

func (w *WebAurion) GetAbsencesPayload() string {
	
	return fmt.Sprintf("%s&%s=%s", w.Payload, w.AbsenceLink, w.AbsenceLink)
}

func (w *WebAurion) GetPlanningPayload() string {
	
	return fmt.Sprintf("%s&%s=%s",w.Payload, w.PlanningLink, w.PlanningLink)
}

func (w *WebAurion) GetPlanningPayload2(viewState string) string {
	startDate := time.Now().AddDate(0, -3, 0)
	endDate := time.Now().AddDate(0, 10, 0)
	startTimestamp := startDate.UnixNano() / int64(time.Millisecond)
	endTimestamp := endDate.UnixNano() / int64(time.Millisecond)

	// fmt.Printf("%v\n", startDate)
	// fmt.Printf("%d\n", startTimestamp)
	// fmt.Printf("%d\n", endTimestamp)

	
	return fmt.Sprintf("javax.faces.partial.ajax=true&javax.faces.source=form%%3Aj_idt118&javax.faces.partial.execute=form%%3Aj_idt118&javax.faces.partial.render=form%%3Aj_idt118&form%%3Aj_idt118=form%%3Aj_idt118&form%%3Aj_idt118_start=%d&form%%3Aj_idt118_end=%d&form=form&form%%3AlargeurDivCenter=&form%%3AidInit=%s&form%%3Adate_input=27%%2F05%%2F2024&form%%3Aweek=22-2024&form%%3Aj_idt118_view=agendaWeek&form%%3AoffsetFuseauNavigateur=-7200000&form%%3Aonglets_activeIndex=0&form%%3Aonglets_scrollState=0&form%%3Aj_idt244_focus=&form%%3Aj_idt244_input=275805&javax.faces.ViewState=%s", startTimestamp, endTimestamp,w.IdInit, viewState)
}

func (w *WebAurion) GetGrades() (*GradeReport, error) {
	data, err := w.DoRequest(w.GetGradesPayload())
	// fmt.Printf("%v\n", string(data))
	if err != nil {
		return nil, err
	}

	beautifulGrade := &BeautifulGrade{}
	gradeReport, err := beautifulGrade.ParseGrades(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing grades: %v", err)
	}

	return gradeReport, nil
}

func (w *WebAurion) GetAbsences() (*AbsenceReport, error) {
	data, err := w.DoRequest(w.GetAbsencesPayload(), "")
	// fmt.Printf("%v\n", string(data))
	if err != nil {
		return nil, err
	}

	// fmt.Printf("%v\n", string(data))

	beautifulAbsences := &BeautifulAbsences{}
	absenceReport, err := beautifulAbsences.ParseAbsences(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing absences: %v", err)
	}

	return absenceReport, nil
}

func (w *WebAurion) GetPlanning() (*PlanningReport, error) {
	resp, err := w.DoRequest(w.GetPlanningPayload())
	if err != nil {
		return nil, fmt.Errorf("error getting initial planning page: %v", err)
	}

	// fmt.Printf("%v\n", string(resp))

	newViewState, err := w.getViewState(strings.NewReader(string(resp)), false)
	if err != nil {
		return nil, fmt.Errorf("error getting new view state: %v", err)
	}

	planningData, err := w.DoRequest(w.GetPlanningPayload2(newViewState), "/webAurion/faces/Planning.xhtml")
	if err != nil {
		return nil, fmt.Errorf("error getting planning data: %v", err)
	}

	// fmt.Printf("%v\n", string(planningData))

	beautifulPlanning := &BeautifulPlanning{}
	planningReport, err := beautifulPlanning.ParsePlanning(planningData)

	// fmt.Printf("%v\n", string(newViewState))
	if err != nil {
		return nil, fmt.Errorf("error parsing planning data: %v", err)
	}

	return planningReport, nil
}

func (w *WebAurion) UserInfo() (*UserInfo, error) {
	nameParts := strings.Fields(w.Name)
	var firstName, lastName []string

	for _, word := range nameParts {
		if word == strings.ToUpper(word) {
			lastName = append(lastName, word)
		} else if string(word[0]) == strings.ToUpper(string(word[0])) {
			firstName = append(firstName, word)
		}
	}

	firstNameStr := strings.Join(firstName, " ")
	lastNameStr := strings.Join(lastName, " ")
	email := fmt.Sprintf("%s.%s@isen-ouest.yncrea.fr", strings.ToLower(strings.Join(firstName, "-")), strings.ToLower(strings.Join(lastName, "-")))


	

	// userInfo := map[string]string{
	// 	"firstname": firstNameStr,
	// 	"lastname":  lastNameStr,
	// 	"name":      w.Name,
	// 	"email":     w.RemoveAccents(email),
	// }

	// if len(lastName) > 1 {
	// 	specEmail := fmt.Sprintf("%s.%s@isen-ouest.yncrea.fr", strings.ToLower(strings.Join(firstName, "-")), strings.ToLower(lastName[0]))
	// 	userInfo["specEmail"] = w.RemoveAccents(specEmail)
	// }

	userInfo := NewUserInfo(firstNameStr, lastNameStr, w.Name, w.RemoveAccents(email))

	return userInfo , nil
}
