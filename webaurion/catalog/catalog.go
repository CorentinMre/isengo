package catalog

import (
	"encoding/json"
	"fmt"
)

// Catalog represents a catalog available in WebAurion (stages, apprenticeships, etc.)
type Catalog struct {
	Name      string `json:"name"`
	SubmenuID string `json:"submenuId"`
	MenuID    string `json:"menuId"`
}

// create a new instance of Catalog
func NewCatalog(name, submenuID, menuID string) *Catalog {
	return &Catalog{
		Name:      name,
		SubmenuID: submenuID,
		MenuID:    menuID,
	}
}

// return a string representation of the Catalog
func (c *Catalog) String() string {
	return fmt.Sprintf("Catalog(name='%s', submenuId='%s', menuId='%s')", c.Name, c.SubmenuID, c.MenuID)
}

// get the value of a specific key for the Catalog
func (c *Catalog) Get(key string) (interface{}, error) {
	switch key {
	case "name":
		return c.Name, nil
	case "submenuId":
		return c.SubmenuID, nil
	case "menuId":
		return c.MenuID, nil
	default:
		return nil, fmt.Errorf("invalid key: %s, valid keys are 'name', 'submenuId', and 'menuId'", key)
	}
}

// return JSON representation of Catalog
func (c *Catalog) JSON() string {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling to JSON: %v", err)
	}
	return string(data)
}

// CatalogEntry represents an entry in a catalog (internship, apprenticeship, etc.)
type CatalogEntry struct {
	Company    string `json:"company"`
	City       string `json:"city"`
	PostalCode string `json:"postalCode"`
	Year       string `json:"year"`
	RowIndex   int    `json:"-"` // internal use only for fetching details
}

// create a new instance of CatalogEntry
func NewCatalogEntry(company, city, postalCode, year string, rowIndex int) *CatalogEntry {
	return &CatalogEntry{
		Company:    company,
		City:       city,
		PostalCode: postalCode,
		Year:       year,
		RowIndex:   rowIndex,
	}
}

// return a string representation of the CatalogEntry
func (ce *CatalogEntry) String() string {
	return fmt.Sprintf("CatalogEntry(company='%s', city='%s', postalCode='%s', year='%s')",
		ce.Company, ce.City, ce.PostalCode, ce.Year)
}

// get the value of a specific key for the CatalogEntry
func (ce *CatalogEntry) Get(key string) (interface{}, error) {
	switch key {
	case "company":
		return ce.Company, nil
	case "city":
		return ce.City, nil
	case "postalCode":
		return ce.PostalCode, nil
	case "year":
		return ce.Year, nil
	default:
		return nil, fmt.Errorf("invalid key: %s, valid keys are 'company', 'city', 'postalCode', and 'year'", key)
	}
}

// CatalogDetails represents the full details of a catalog entry
type CatalogDetails struct {
	Title       string `json:"title"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Description string `json:"description"`
	Company     string `json:"company"`
	City        string `json:"city"`
	PostalCode  string `json:"postalCode"`
	Year        string `json:"year"`
	StudentName string `json:"studentName,omitempty"`
}

// create a new instance of CatalogDetails
func NewCatalogDetails(title, startDate, endDate, description, company, city, postalCode, year, studentName string) *CatalogDetails {
	return &CatalogDetails{
		Title:       title,
		StartDate:   startDate,
		EndDate:     endDate,
		Description: description,
		Company:     company,
		City:        city,
		PostalCode:  postalCode,
		Year:        year,
		StudentName: studentName,
	}
}

// return a string representation of the CatalogDetails
func (cd *CatalogDetails) String() string {
	return fmt.Sprintf("CatalogDetails(title='%s', startDate='%s', endDate='%s', company='%s', city='%s')",
		cd.Title, cd.StartDate, cd.EndDate, cd.Company, cd.City)
}

// get the value of a specific key for the CatalogDetails
func (cd *CatalogDetails) Get(key string) (interface{}, error) {
	switch key {
	case "title":
		return cd.Title, nil
	case "startDate":
		return cd.StartDate, nil
	case "endDate":
		return cd.EndDate, nil
	case "description":
		return cd.Description, nil
	case "company":
		return cd.Company, nil
	case "city":
		return cd.City, nil
	case "postalCode":
		return cd.PostalCode, nil
	case "year":
		return cd.Year, nil
	case "studentName":
		return cd.StudentName, nil
	default:
		return nil, fmt.Errorf("invalid key: %s", key)
	}
}

// return JSON representation of CatalogDetails
func (cd *CatalogDetails) JSON() string {
	data, err := json.MarshalIndent(cd, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling to JSON: %v", err)
	}
	return string(data)
}

// CatalogReport represents a report about catalog entries
type CatalogReport struct {
	TotalEntries int            `json:"totalEntries"`
	Entries      []CatalogEntry `json:"entries"`
}

// create a new instance of CatalogReport
func NewCatalogReport(entries []CatalogEntry) *CatalogReport {
	return &CatalogReport{
		TotalEntries: len(entries),
		Entries:      entries,
	}
}

// return a string representation of the CatalogReport
func (cr *CatalogReport) String() string {
	return fmt.Sprintf("CatalogReport(totalEntries=%d, entries=%v)", cr.TotalEntries, cr.Entries)
}

// get the value of a specific key for the CatalogReport
func (cr *CatalogReport) Get(key string) (interface{}, error) {
	switch key {
	case "totalEntries":
		return cr.TotalEntries, nil
	case "entries":
		return cr.Entries, nil
	default:
		return nil, fmt.Errorf("invalid key: %s, valid keys are 'totalEntries' and 'entries'", key)
	}
}

// return JSON representation of CatalogReport
func (cr *CatalogReport) JSON() string {
	data, err := json.MarshalIndent(cr, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling to JSON: %v", err)
	}
	return string(data)
}
