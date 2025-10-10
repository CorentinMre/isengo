package catalog

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// parse catalog entries from HTML document
func ParseCatalogEntries(doc *goquery.Document) []CatalogEntry {
	entries := []CatalogEntry{}

	// parse the table
	doc.Find("tbody#form\\:j_idt193_data tr").Each(func(i int, row *goquery.Selection) {
		rowIndex, exists := row.Attr("data-ri")
		if !exists {
			return
		}

		cells := row.Find("td")
		if cells.Length() < 1 {
			return
		}

		var company, city, postalCode, year string

		// helper function to extract text from cell
		extractCellText := func(cell *goquery.Selection) string {
			// try to find span.preformatted first
			text := strings.TrimSpace(cell.Find("span.preformatted").Text())
			if text != "" {
				return text
			}
			// fallback: get all text but remove the ui-column-title span
			cell.Find("span.ui-column-title").Remove()
			return strings.TrimSpace(cell.Text())
		}

		// extract company (always first column)
		company = extractCellText(cells.Eq(0))

		// handle different table structures
		if cells.Length() == 6 {
			// structure with button: Entreprise, Ville, Code postal, Pays, Année, Button
			city = extractCellText(cells.Eq(1))
			postalCode = extractCellText(cells.Eq(2))
			// skip Pays (index 3)
			year = strings.ReplaceAll(extractCellText(cells.Eq(4)), "\u00a0", " ")
		} else if cells.Length() == 5 {
			// structure: Entreprise, Ville, Code postal, Pays, Année
			city = extractCellText(cells.Eq(1))
			postalCode = extractCellText(cells.Eq(2))
			// skip Pays (index 3)
			year = strings.ReplaceAll(extractCellText(cells.Eq(4)), "\u00a0", " ")
		} else if cells.Length() == 4 {
			// structure: Entreprise, Ville, Code postal, Année
			city = extractCellText(cells.Eq(1))
			postalCode = extractCellText(cells.Eq(2))
			year = strings.ReplaceAll(extractCellText(cells.Eq(3)), "\u00a0", " ")
		} else if cells.Length() >= 2 {
			// simplified structure: Entreprise, Année
			year = strings.ReplaceAll(extractCellText(cells.Eq(1)), "\u00a0", " ")
		}

		var idx int
		fmt.Sscanf(rowIndex, "%d", &idx)

		entries = append(entries, *NewCatalogEntry(company, city, postalCode, year, idx))
	})

	return entries
}

// parse catalog list from HTML document
func ParseCatalogs(doc *goquery.Document, submenuID string) []Catalog {
	catalogs := []Catalog{}

	// find all catalog menu items
	doc.Find(fmt.Sprintf("ul#%s li.ui-menuitem", submenuID)).Each(func(i int, item *goquery.Selection) {
		link := item.Find("a.ui-menuitem-link")
		name := strings.TrimSpace(link.Find("span.ui-menuitem-text").Text())

		// extract onclick parameters
		onclick, exists := link.Attr("onclick")
		if !exists {
			return
		}

		params := parseOnclickButton(onclick)
		menuID, hasMenuID := params["form:sidebar_menuid"]
		if !hasMenuID {
			return
		}

		catalogs = append(catalogs, *NewCatalog(name, submenuID, menuID))
	})

	return catalogs
}

// parse onclick button to extract parameters
func parseOnclickButton(onclick string) map[string]string {
	params := make(map[string]string)

	// find the parameter object: PrimeFaces.ab({...})
	start := strings.Index(onclick, "{")
	end := strings.LastIndex(onclick, "}")
	if start == -1 || end == -1 {
		return params
	}

	paramString := onclick[start+1 : end]

	// manual character-by-character parsing respecting single quotes
	i := 0
	for i < len(paramString) {
		// skip to opening quote
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
		i++ // skip closing '

		// skip to next opening quote for value
		for i < len(paramString) && paramString[i] != '\'' {
			i++
		}
		if i >= len(paramString) {
			break
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
		i++ // skip closing '

		params[key] = value
	}

	return params
}
