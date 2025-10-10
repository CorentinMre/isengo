package catalog

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// load all available catalogs from WebAurion
func LoadCatalogs(w WebAurionClient, payload string) ([]Catalog, error) {
	// first, we need to load the "Divers" submenu to get the catalogs
	submenuPayload := payload + "&javax.faces.partial.ajax=true&javax.faces.source=webscolaapp.Sidebar.ID_SUBMENU&javax.faces.partial.execute=@all&webscolaapp.Sidebar.ID_SUBMENU=webscolaapp.Sidebar.ID_SUBMENU&webscolaapp.Sidebar.ID_SUBMENU_menuid=6&form:sidebar_expandedMenuId=6_0"

	req, err := http.NewRequest("POST", w.GetBaseURL()+"/webAurion/faces/MesDonneesAccueil.xhtml", strings.NewReader(submenuPayload))
	if err != nil {
		return nil, fmt.Errorf("error creating submenu request: %v", err)
	}

	w.SetRequestHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Faces-Request", "partial/ajax")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := w.GetClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("error loading submenu: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading submenu response: %v", err)
	}

	// parse the XML response to extract HTML from CDATA
	responseStr := string(bodyBytes)
	cdataStart := strings.Index(responseStr, "<![CDATA[")
	if cdataStart == -1 {
		return nil, fmt.Errorf("CDATA not found in submenu response")
	}

	cdataEnd := strings.Index(responseStr[cdataStart:], "]]>")
	if cdataEnd == -1 {
		return nil, fmt.Errorf("CDATA end not found in submenu response")
	}
	cdataEnd += cdataStart

	htmlContent := responseStr[cdataStart+9 : cdataEnd]

	// parse the HTML to extract catalog information
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("error parsing submenu HTML: %v", err)
	}

	// extract submenuID from the HTML
	submenuID := ""
	doc.Find("ul.ui-menu-list").Each(func(i int, ul *goquery.Selection) {
		if id, exists := ul.Attr("id"); exists {
			submenuID = id
		}
	})

	if submenuID == "" {
		return nil, fmt.Errorf("submenu ID not found in response")
	}

	return ParseCatalogs(doc, submenuID), nil
}
