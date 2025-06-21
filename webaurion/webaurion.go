package webaurion

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type WebAurion struct {
	BaseURL          string
	Cookies          []*http.Cookie
	ViewState        string
	Link             map[string]string
	LastRequetTime   time.Time
	Name             string
	LoggedIn         bool
	Client           *http.Client
	GradeLink        string
	AbsenceLink      string
	PlanningLink     string
	IdInit           string
	IdBasic          string
	Payload          string
	ProxyEndpoints   []string
	currentProxyIndex int
}


func NewWebAurion() *WebAurion {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	return &WebAurion{
		BaseURL:           "https://web.isen-ouest.fr",
		Link:              make(map[string]string),
		LoggedIn:          false,
		Client:            client,
		ProxyEndpoints:    []string{},
		currentProxyIndex: 0,
	}
}


func NewWebAurionWithProxies(proxyEndpoints []string) *WebAurion {
	w := NewWebAurion()
	w.SetProxies(proxyEndpoints)
	return w
}


func (w *WebAurion) SetProxies(proxyEndpoints []string) error {
	w.ProxyEndpoints = proxyEndpoints
	w.currentProxyIndex = 0
	
	if len(proxyEndpoints) > 0 {
		// proxy aléatoire
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(w.ProxyEndpoints), func(i, j int) {
			w.ProxyEndpoints[i], w.ProxyEndpoints[j] = w.ProxyEndpoints[j], w.ProxyEndpoints[i]
		})
		
		return w.updateClientWithProxy()
	}
	
	return nil
}


func (w *WebAurion) updateClientWithProxy() error {
	if len(w.ProxyEndpoints) == 0 { // pas de proxy

		jar, _ := cookiejar.New(nil)
		w.Client = &http.Client{
			Jar: jar,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		return nil
	}
	
	proxyURL, err := url.Parse(w.ProxyEndpoints[w.currentProxyIndex])
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %v", err)
	}
	
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	
	w.Client = &http.Client{
		Jar:       jar,
		Transport: transport,
		Timeout:   30 * time.Second,
	}
	
	return nil
}


func (w *WebAurion) rotateProxy() error {
	if len(w.ProxyEndpoints) <= 1 { // si que 1 proxy, pas besoin de rotation
		return nil
	}
	
	w.currentProxyIndex = (w.currentProxyIndex + 1) % len(w.ProxyEndpoints)
	return w.updateClientWithProxy()
}


func (w *WebAurion) getCurrentProxy() string {
	if len(w.ProxyEndpoints) == 0 {
		return "Aucun proxy"
	}
	return w.ProxyEndpoints[w.currentProxyIndex]
}


func (w *WebAurion) testProxy(proxyURL string) error {
	proxyParsed, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %v", err)
	}
	
	transport := &http.Transport{
		Proxy:           http.ProxyURL(proxyParsed),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}
	
	resp, err := client.Get("https://httpbin.org/ip")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("proxy test failed with status: %d", resp.StatusCode)
	}
	
	return nil
}


func (w *WebAurion) TestAllProxies() map[string]error {
	results := make(map[string]error)
	
	for _, proxy := range w.ProxyEndpoints {
		results[proxy] = w.testProxy(proxy)
	}
	
	return results
}

func (w *WebAurion) Login(username, password string) (bool, error) {
	return w.LoginWithRetry(username, password, 3)
}


func (w *WebAurion) LoginWithRetry(username, password string, maxRetries int) (bool, error) {
	var lastError error
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 && len(w.ProxyEndpoints) > 1 {
			// change proxy
			w.rotateProxy()
			fmt.Printf("Tentative %d avec proxy: %s\n", attempt+1, w.getCurrentProxy())
		}
		
		success, err := w.performLogin(username, password)
		if success {
			return true, nil
		}
		
		lastError = err
		
		// wait a bit before retrying
		if attempt < maxRetries-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}
	
	return false, fmt.Errorf("login failed after %d attempts. Last error: %v", maxRetries, lastError)
}

func (w *WebAurion) performLogin(username, password string) (bool, error) {
	payload := url.Values{}
	payload.Set("username", username)
	payload.Set("password", password)
	payload.Set("j_idt27", "")

	req, err := http.NewRequest("POST", w.BaseURL+"/webAurion/login", strings.NewReader(payload.Encode()))
	if err != nil {
		return false, errors.New("error creating login request")
	}

	w.setRequestHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := w.Client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error during login request: %v", err)
	}
	defer resp.Body.Close()

	w.Cookies = resp.Cookies()

	// get the main page
	req, err = http.NewRequest("GET", w.BaseURL+"/webAurion/", nil)
	if err != nil {
		return false, errors.New("error creating main page request")
	}
	w.setRequestHeaders(req)

	resp, err = w.Client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error getting main page: %v", err)
	}
	defer resp.Body.Close()

	w.ViewState, err = w.getViewState(resp.Body, true)
	if err != nil {
		return false, errors.New("username or password incorrect")
	}

	w.LoggedIn = true
	w.LastRequetTime = time.Now()
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
	return w.DoRequestWithRetry(payload, 3, referer...)
}

// DoRequestWithRetry effectue une requête avec retry automatique
func (w *WebAurion) DoRequestWithRetry(payload string, maxRetries int, referer ...string) ([]byte, error) {
	var lastError error
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 && len(w.ProxyEndpoints) > 1 {
			
			w.rotateProxy()
		}
		
		data, err := w.performRequest(payload, referer...)
		if err == nil {
			return data, nil
		}
		
		lastError = err
		
		
		if attempt < maxRetries-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}
	
	return nil, fmt.Errorf("request failed after %d attempts. Last error: %v", maxRetries, lastError)
}

func (w *WebAurion) performRequest(payload string, referer ...string) ([]byte, error) {
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
		return nil, fmt.Errorf("request failed with proxy %s: %v", w.getCurrentProxy(), err)
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
			if strings.Contains(s.Text(), "note") {
				w.GradeLink = id
			} else if strings.Contains(s.Text(), "Absences") {
				w.AbsenceLink = id
			} else if strings.Contains(s.Text(), "Planning") {
				w.PlanningLink = id
			}
		})

		doc.Find("input").Each(func(i int, s *goquery.Selection) {
			name, _ := s.Attr("name")
			value, _ := s.Attr("value")
			w.Payload += fmt.Sprintf("%s=%s&", name, value)
		})

		w.IdBasic, _ = doc.Find("input[value='basicDay']").Attr("id")
		w.Payload += "form:j_idt820_input=275805"
	}

	viewState, exists := doc.Find("input[name='javax.faces.ViewState']").Attr("value")
	if !exists {
		return "", fmt.Errorf("ViewState not found")
	}

	return viewState, nil
}

func (w *WebAurion) setRequestHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "ISENGO https://github.com/CorentinMre/isengo")
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
	return fmt.Sprintf("%s&%s=%s", w.Payload, w.PlanningLink, w.PlanningLink)
}

func (w *WebAurion) GetPlanningPayload2(viewState string) string {
	startDate := time.Now().AddDate(0, -3, 0)
	endDate := time.Now().AddDate(0, 10, 0)
	startTimestamp := startDate.UnixNano() / int64(time.Millisecond)
	endTimestamp := endDate.UnixNano() / int64(time.Millisecond)

	return fmt.Sprintf("javax.faces.partial.ajax=true&javax.faces.source=form%%3Aj_idt118&javax.faces.partial.execute=form%%3Aj_idt118&javax.faces.partial.render=form%%3Aj_idt118&form%%3Aj_idt118=form%%3Aj_idt118&form%%3Aj_idt118_start=%d&form%%3Aj_idt118_end=%d&form=form&form%%3AlargeurDivCenter=&form%%3AidInit=%s&form%%3Adate_input=27%%2F05%%2F2024&form%%3Aweek=22-2024&form%%3Aj_idt118_view=agendaWeek&form%%3AoffsetFuseauNavigateur=-7200000&form%%3Aonglets_activeIndex=0&form%%3Aonglets_scrollState=0&form%%3Aj_idt244_focus=&form%%3Aj_idt244_input=275805&javax.faces.ViewState=%s", startTimestamp, endTimestamp, w.IdInit, viewState)
}

func (w *WebAurion) GetGrades() (*GradeReport, error) {
	data, err := w.DoRequest(w.GetGradesPayload())
	if err != nil {
		return nil, err
	}

	beautifulGrade := &BeautifulGrade{}
	gradeReport, err := beautifulGrade.ParseGrades(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing grades: %v", err)
	}

	w.LastRequetTime = time.Now()
	return gradeReport, nil
}

func (w *WebAurion) GetAbsences() (*AbsenceReport, error) {
	data, err := w.DoRequest(w.GetAbsencesPayload(), "")
	if err != nil {
		return nil, err
	}

	beautifulAbsences := &BeautifulAbsences{}
	absenceReport, err := beautifulAbsences.ParseAbsences(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing absences: %v", err)
	}

	w.LastRequetTime = time.Now()
	return absenceReport, nil
}

func (w *WebAurion) GetPlanning() (*PlanningReport, error) {
	resp, err := w.DoRequest(w.GetPlanningPayload())
	if err != nil {
		return nil, fmt.Errorf("error getting initial planning page: %v", err)
	}

	newViewState, err := w.getViewState(strings.NewReader(string(resp)), false)
	if err != nil {
		return nil, fmt.Errorf("error getting new view state: %v", err)
	}

	planningData, err := w.DoRequest(w.GetPlanningPayload2(newViewState), "/webAurion/faces/Planning.xhtml")
	if err != nil {
		return nil, fmt.Errorf("error getting planning data: %v", err)
	}

	beautifulPlanning := &BeautifulPlanning{}
	planningReport, err := beautifulPlanning.ParsePlanning(planningData)
	if err != nil {
		return nil, fmt.Errorf("error parsing planning data: %v", err)
	}

	w.LastRequetTime = time.Now()
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

	userInfo := NewUserInfo(firstNameStr, lastNameStr, w.Name, w.RemoveAccents(email))
	return userInfo, nil
}

func (w *WebAurion) Refresh() error {
	if time.Since(w.LastRequetTime) > 30*time.Minute {
		_, err := w.DoRequest(w.GetGradesPayload())
		if err != nil {
			return err
		}
		w.LastRequetTime = time.Now()
	}
	return nil
}