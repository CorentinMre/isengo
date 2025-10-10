package catalog

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// load all available catalogs from WebAurion (main method)
func LoadCatalogsFromWebAurion(w WebAurionClient) ([]Catalog, error) {
	// first, load the main page to get the "Divers" submenu ID
	req, err := http.NewRequest("GET", w.GetBaseURL()+"/webAurion/", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	w.SetRequestHeaders(req)

	resp, err := w.GetClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("error loading page: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %v", err)
	}

	// find the submenu ID of "Divers"
	diversSubmenuID := findDiversSubmenuID(doc)
	if diversSubmenuID == "" {
		return nil, fmt.Errorf("menu 'Divers' not found")
	}

	// get ViewState for AJAX request
	viewState, err := w.GetViewState(strings.NewReader(string(bodyBytes)), false)
	if err != nil {
		return nil, fmt.Errorf("error getting ViewState: %v", err)
	}

	// extract necessary parameters from payload
	payloadParts := strings.Split(w.GetPayload(), "&")
	var largeurDivCenter, idInit, jIdt837Input string

	for _, part := range payloadParts {
		if strings.HasPrefix(part, "form:largeurDivCenter=") {
			largeurDivCenter = strings.TrimPrefix(part, "form:largeurDivCenter=")
		} else if strings.HasPrefix(part, "form:idInit=") {
			idInit = strings.TrimPrefix(part, "form:idInit=")
		} else if strings.HasPrefix(part, "form:j_idt820_input=") {
			jIdt837Input = strings.TrimPrefix(part, "form:j_idt820_input=")
		}
	}

	payload := fmt.Sprintf("javax.faces.partial.ajax=true&javax.faces.source=form:j_idt52&javax.faces.partial.execute=form:j_idt52&javax.faces.partial.render=form:sidebar&form:j_idt52=form:j_idt52&webscolaapp.Sidebar.ID_SUBMENU=%s&form=form&form:largeurDivCenter=%s&form:idInit=%s&form:sauvegarde=&form:j_idt774:j_idt776_page=0&form:j_idt822:j_idt825_view=basicDay&form:j_idt837_focus=&form:j_idt837_input=%s&javax.faces.ViewState=%s",
		diversSubmenuID, largeurDivCenter, idInit, jIdt837Input, url.QueryEscape(viewState))

	// make AJAX request
	req2, err := http.NewRequest("POST", w.GetBaseURL()+"/webAurion/faces/MainMenuPage.xhtml", strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("error creating AJAX request: %v", err)
	}

	w.SetRequestHeaders(req2)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req2.Header.Set("Faces-Request", "partial/ajax")
	req2.Header.Set("X-Requested-With", "XMLHttpRequest")
	req2.Header.Set("Accept", "application/xml, text/xml, */*; q=0.01")

	resp2, err := w.GetClient().Do(req2)
	if err != nil {
		return nil, fmt.Errorf("error loading submenu: %v", err)
	}
	defer resp2.Body.Close()

	bodyBytes2, err := io.ReadAll(resp2.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading submenu response: %v", err)
	}

	// extract HTML from CDATA in XML response
	responseStr := string(bodyBytes2)

	cdataStart := strings.Index(responseStr, "<![CDATA[")
	cdataEnd := strings.Index(responseStr, "]]>")

	if cdataStart == -1 || cdataEnd == -1 {
		return nil, fmt.Errorf("CDATA not found in response")
	}

	htmlContent := responseStr[cdataStart+9 : cdataEnd]

	// parse extracted HTML
	doc2, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("error parsing submenu HTML: %v", err)
	}

	return parseCatalogsFromDoc(doc2), nil
}

// find the submenu ID of "Divers" menu
func findDiversSubmenuID(doc *goquery.Document) string {
	var submenuID string
	doc.Find("li.ui-menu-parent").Each(func(i int, parent *goquery.Selection) {
		text := strings.TrimSpace(parent.Find("span.ui-menuitem-text").First().Text())
		if text == "Divers" {
			class, _ := parent.Attr("class")
			classes := strings.Split(class, " ")
			for _, c := range classes {
				if strings.HasPrefix(c, "submenu_") {
					submenuID = c
					return
				}
			}
		}
	})
	return submenuID
}

// parse catalogs from document (finds all catalog menu items)
func parseCatalogsFromDoc(doc *goquery.Document) []Catalog {
	catalogs := []Catalog{}

	doc.Find("a.ui-menuitem-link").Each(func(i int, link *goquery.Selection) {
		name := strings.TrimSpace(link.Find("span.ui-menuitem-text").Text())

		if !strings.Contains(name, "Catalogue") {
			return
		}

		onclick, onclickExists := link.Attr("onclick")
		if !onclickExists {
			return
		}

		params := parseOnclickButtonParams(onclick)
		menuID, menuExists := params["form:sidebar_menuid"]
		if !menuExists {
			return
		}

		// find parent submenu by traversing DOM
		submenuID := ""
		item := link.Parent() // li.ui-menuitem
		if item.Length() > 0 {
			ul := item.Parent() // ul.ui-menu-child
			if ul.Length() > 0 {
				parent := ul.Parent() // li.ui-menu-parent
				if parent.Length() > 0 {
					class, _ := parent.Attr("class")
					classes := strings.Split(class, " ")
					for _, c := range classes {
						if strings.HasPrefix(c, "submenu_") {
							submenuID = c
							break
						}
					}
				}
			}
		}

		catalogs = append(catalogs, *NewCatalog(name, submenuID, menuID))
	})

	return catalogs
}

// parse onclick button parameters (handles PrimeFaces format)
func parseOnclickButtonParams(onclick string) map[string]string {
	params := make(map[string]string)

	parts := strings.Split(onclick, "PrimeFaces.addSubmitParam")
	if len(parts) < 2 {
		return params
	}

	paramPart := parts[1]
	startIndex := strings.Index(paramPart, "{")
	endIndex := strings.Index(paramPart, "}")
	if startIndex == -1 || endIndex == -1 || endIndex <= startIndex {
		return params
	}

	paramString := paramPart[startIndex+1 : endIndex]

	// manual parsing respecting single quotes
	i := 0
	for i < len(paramString) {
		// find start of key
		if i >= len(paramString) || paramString[i] != '\'' {
			i++
			continue
		}
		i++ // skip '

		// read key until closing '
		keyStart := i
		for i < len(paramString) && paramString[i] != '\'' {
			i++
		}
		if i >= len(paramString) {
			break
		}
		key := paramString[keyStart:i]
		i++ // skip '

		// skip : between key and value
		for i < len(paramString) && (paramString[i] == ':' || paramString[i] == ' ') {
			i++
		}

		// find start of value
		if i >= len(paramString) || paramString[i] != '\'' {
			continue
		}
		i++ // skip '

		// read value until closing '
		valueStart := i
		for i < len(paramString) && paramString[i] != '\'' {
			i++
		}
		if i >= len(paramString) {
			break
		}
		value := paramString[valueStart:i]
		i++ // skip '

		params[key] = value

		// skip , between pairs
		for i < len(paramString) && (paramString[i] == ',' || paramString[i] == ' ') {
			i++
		}
	}

	return params
}
