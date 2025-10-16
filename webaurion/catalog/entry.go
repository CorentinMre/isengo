package catalog

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// WebAurionClient interface to avoid circular dependency
type WebAurionClient interface {
	GetBaseURL() string
	GetClient() *http.Client
	SetRequestHeaders(req *http.Request)
	GetViewState(reader io.Reader, isInitial bool) (string, error)
	GetPayload() string
}

// retrieve all entries from a catalog (handles pagination automatically)
func GetCatalogEntries(w WebAurionClient, catalogIndex int, catalogs []Catalog, doRequest func(string, ...string) ([]byte, error)) (*CatalogReport, error) {
	// get the payload for the catalog
	payload, err := GetCatalogPayload(w, catalogIndex, catalogs)
	if err != nil {
		return nil, err
	}

	// make the POST request to access the catalog
	data, err := doRequest(payload)
	if err != nil {
		return nil, fmt.Errorf("error loading catalog: %v", err)
	}

	// parse HTML to extract entries from the first page
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("error parsing catalog HTML: %v", err)
	}

	entries := ParseCatalogEntries(doc)

	// check if there are more pages
	hasMorePages := doc.Find("a.ui-paginator-next:not(.ui-state-disabled)").Length() > 0

	if hasMorePages {
		// get ViewState and idInit for pagination requests
		viewState, _ := w.GetViewState(strings.NewReader(string(data)), false)
		idInit := ""
		if input := doc.Find("input[name='form:idInit']"); input.Length() > 0 {
			idInit, _ = input.Attr("value")
		}

		// fetch all subsequent pages
		first := 20 // first element of the next page
		pageNum := 2
		for hasMorePages {
			moreEntries, more, err := getCatalogPage(w, first, viewState, idInit)
			if err != nil {
				break
			}

			if len(moreEntries) == 0 {
				break
			}

			entries = append(entries, moreEntries...)
			hasMorePages = more
			first += 20
			pageNum++

			// safety: limit to 100 pages max
			if pageNum > 100 {
				break
			}
		}
	}

	return NewCatalogReport(entries), nil
}

// fetch a specific page of the catalog (AJAX pagination)
func getCatalogPage(w WebAurionClient, first int, viewState, idInit string) ([]CatalogEntry, bool, error) {
	// extract necessary parameters
	payloadParts := strings.Split(w.GetPayload(), "&")
	var largeurDivCenter, jIdt267Input string

	for _, part := range payloadParts {
		if strings.HasPrefix(part, "form:largeurDivCenter=") {
			largeurDivCenter = strings.TrimPrefix(part, "form:largeurDivCenter=")
		} else if strings.HasPrefix(part, "form:j_idt820_input=") {
			jIdt267Input = strings.TrimPrefix(part, "form:j_idt820_input=")
		}
	}

	// build AJAX payload for pagination
	payload := fmt.Sprintf("javax.faces.partial.ajax=true&javax.faces.source=form:j_idt193&javax.faces.partial.execute=form:j_idt193&javax.faces.partial.render=form:j_idt193&form:j_idt193=form:j_idt193&form:j_idt193_pagination=true&form:j_idt193_first=%d&form:j_idt193_rows=20&form:j_idt193_skipChildren=true&form:j_idt193_encodeFeature=true&form=form&form:largeurDivCenter=%s&form:idInit=%s&form:messagesRubriqueInaccessible=&form:search-texte=&form:search-texte-avancer=&form:input-expression-exacte=&form:input-un-des-mots=&form:input-aucun-des-mots=&form:input-nombre-debut=&form:input-nombre-fin=&form:calendarDebut_input=&form:calendarFin_input=&form:j_idt193_reflowDD=0_0&form:j_idt193:j_idt198:filter=&form:j_idt193:j_idt200:filter=&form:j_idt193:j_idt202:filter=&form:j_idt193:j_idt204:filter=&form:j_idt267_focus=&form:j_idt267_input=%s&javax.faces.ViewState=%s",
		first, largeurDivCenter, idInit, jIdt267Input, url.QueryEscape(viewState))

	// make the AJAX request
	req, err := http.NewRequest("POST", w.GetBaseURL()+"/webAurion/faces/ChoixEvenementDUnFormulaire.xhtml", strings.NewReader(payload))
	if err != nil {
		return nil, false, fmt.Errorf("error creating pagination request: %v", err)
	}

	w.SetRequestHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Faces-Request", "partial/ajax")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Accept", "application/xml, text/xml, */*; q=0.01")

	resp, err := w.GetClient().Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("error getting page: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("error reading page body: %v", err)
	}

	// extract HTML from CDATA containing form:j_idt193
	responseStr := string(bodyBytes)

	// find the section <update id="form:j_idt193">
	tableUpdateStart := strings.Index(responseStr, `<update id="form:j_idt193">`)
	if tableUpdateStart == -1 {
		return nil, false, fmt.Errorf("table update section not found in pagination response")
	}

	// extract CDATA from this section
	cdataStart := strings.Index(responseStr[tableUpdateStart:], "<![CDATA[")
	if cdataStart == -1 {
		return nil, false, fmt.Errorf("CDATA not found in table update section")
	}
	cdataStart += tableUpdateStart

	cdataEnd := strings.Index(responseStr[cdataStart:], "]]>")
	if cdataEnd == -1 {
		return nil, false, fmt.Errorf("CDATA end not found")
	}
	cdataEnd += cdataStart

	htmlContent := responseStr[cdataStart+9 : cdataEnd]

	// parse the new entries - need to wrap in table/tbody since these are fragments
	wrappedHTML := "<table><tbody>" + htmlContent + "</tbody></table>"
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(wrappedHTML))
	if err != nil {
		return nil, false, fmt.Errorf("error parsing page HTML: %v", err)
	}

	entries := []CatalogEntry{}
	doc.Find("tr").Each(func(i int, row *goquery.Selection) {
		rowIndex, exists := row.Attr("data-ri")
		if !exists {
			return
		}

		cells := row.Find("td")
		if cells.Length() < 4 {
			return
		}

		company := strings.TrimSpace(cells.Eq(0).Find("span.preformatted").Text())
		city := strings.TrimSpace(cells.Eq(1).Find("span.preformatted").Text())
		postalCode := strings.TrimSpace(cells.Eq(2).Find("span.preformatted").Text())
		year := strings.ReplaceAll(strings.TrimSpace(cells.Eq(3).Find("span.preformatted").Text()), "\u00a0", " ")

		var idx int
		fmt.Sscanf(rowIndex, "%d", &idx)
		entries = append(entries, *NewCatalogEntry(company, city, postalCode, year, idx))
	})

	// check if there are more pages (consider there are if we got 20 entries)
	hasMore := len(entries) == 20

	return entries, hasMore, nil
}

// retrieve the details of a catalog entry
func GetCatalogEntryDetails(w WebAurionClient, entry CatalogEntry) (*CatalogDetails, error) {
	// get current ViewState
	req, err := http.NewRequest("GET", w.GetBaseURL()+"/webAurion/faces/ChoixEvenementDUnFormulaire.xhtml", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	w.SetRequestHeaders(req)

	resp, err := w.GetClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting page: %v", err)
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

	viewState, err := w.GetViewState(strings.NewReader(string(bodyBytes)), false)
	if err != nil {
		return nil, fmt.Errorf("error getting ViewState: %v", err)
	}

	// extract necessary parameters from payload
	payloadParts := strings.Split(w.GetPayload(), "&")
	var largeurDivCenter, idInit, jIdt267Input string

	for _, part := range payloadParts {
		if strings.HasPrefix(part, "form:largeurDivCenter=") {
			largeurDivCenter = strings.TrimPrefix(part, "form:largeurDivCenter=")
		} else if strings.HasPrefix(part, "form:idInit=") {
			idInit = strings.TrimPrefix(part, "form:idInit=")
		} else if strings.HasPrefix(part, "form:j_idt820_input=") {
			jIdt267Input = strings.TrimPrefix(part, "form:j_idt820_input=")
		}
	}

	// find idInit from current page
	idInitInput := doc.Find("input[name='form:idInit']")
	if idInitInput.Length() > 0 {
		if val, exists := idInitInput.Attr("value"); exists {
			idInit = val
		}
	}

	// build payload for "Consulter" button
	payload := fmt.Sprintf("form=form&form:largeurDivCenter=%s&form:idInit=%s&form:messagesRubriqueInaccessible=&form:search-texte=&form:search-texte-avancer=&form:input-expression-exacte=&form:input-un-des-mots=&form:input-aucun-des-mots=&form:input-nombre-debut=&form:input-nombre-fin=&form:calendarDebut_input=&form:calendarFin_input=&form:j_idt193_reflowDD=0_0&form:j_idt193:j_idt198:filter=&form:j_idt193:j_idt200:filter=&form:j_idt193:j_idt202:filter=&form:j_idt193:j_idt204:filter=&form:j_idt193:%d:j_idt215=&form:j_idt265_focus=&form:j_idt265_input=%s&javax.faces.ViewState=%s",
		largeurDivCenter, idInit, entry.RowIndex, jIdt267Input, url.QueryEscape(viewState))

	// make the POST request
	req2, err := http.NewRequest("POST", w.GetBaseURL()+"/webAurion/faces/ChoixEvenementDUnFormulaire.xhtml", strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("error creating detail request: %v", err)
	}

	w.SetRequestHeaders(req2)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp2, err := w.GetClient().Do(req2)
	if err != nil {
		return nil, fmt.Errorf("error getting details: %v", err)
	}
	defer resp2.Body.Close()

	bodyBytes2, err := io.ReadAll(resp2.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading details body: %v", err)
	}

	// parse the details
	doc2, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyBytes2)))
	if err != nil {
		return nil, fmt.Errorf("error parsing details HTML: %v", err)
	}

	details := &CatalogDetails{
		Company:    entry.Company,
		City:       entry.City,
		PostalCode: entry.PostalCode,
		Year:       entry.Year,
	}

	// extract information
	doc2.Find("div.ligne").Each(func(i int, ligne *goquery.Selection) {
		label := strings.TrimSpace(ligne.Find("label span.ev_libelle").Text())

		// get value from colonne2
		colonne2 := ligne.Find("div.colonne2")

		// try different selectors for value
		value := strings.TrimSpace(colonne2.Find("span.composant-type-string").Text())
		if value == "" {
			value = strings.TrimSpace(colonne2.Find("span.composant-type-text").Text())
		}
		if value == "" {
			// fallback: get first span text (for dates)
			value = strings.TrimSpace(colonne2.Find("span").First().Text())
		}

		switch label {
		case "Titre de la mission", "Titre du stage", "Titre de l'apprentissage":
			details.Title = value
		case "Début de l'apprentissage", "Début du stage", "Date de début":
			details.StartDate = value
		case "Fin de l'apprentissage", "Fin du stage", "Date de fin":
			details.EndDate = value
		case "Description de l'activité prévue", "Description de l'activité":
			details.Description = value
		case "NOM Prénom", "Nom Prénom", "Étudiant":
			details.StudentName = value
		}
	})

	return details, nil
}

// get the payload for accessing a catalog by index
func GetCatalogPayload(w WebAurionClient, catalogIndex int, catalogs []Catalog) (string, error) {
	if catalogIndex < 0 || catalogIndex >= len(catalogs) {
		return "", fmt.Errorf("catalog index out of range")
	}

	catalog := catalogs[catalogIndex]
	return BuildCatalogPayload(w.GetPayload(), catalog.MenuID), nil
}

// get the payload for accessing a catalog by name
func GetCatalogPayloadByName(w WebAurionClient, catalogName string, catalogs []Catalog) (string, error) {
	for _, catalog := range catalogs {
		if catalog.Name == catalogName {
			return BuildCatalogPayload(w.GetPayload(), catalog.MenuID), nil
		}
	}
	return "", fmt.Errorf("catalog not found")
}

// build the payload for a catalog menu item
func BuildCatalogPayload(basePayload, menuID string) string {
	return basePayload + "&form:sidebar=form:sidebar&form:sidebar_menuid=" + menuID
}
