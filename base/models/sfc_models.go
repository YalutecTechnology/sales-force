package models

type SfcContact struct {
	ID          string `json:"Id"`
	FirstName   string `json:"FirstName"`
	LastName    string `json:"LastName"`
	Email       string `json:"Email"`
	MobilePhone string `json:"MobilePhone"`
	AccountID   string `json:"AccountID"`
	// This field will be given to us by salesforce, to know if a user is blocked
	Blocked bool `json:"Blocked"`
}

type SfcAccount struct {
	ID                string `json:"Id"`
	FirstName         string `json:"FirstName"`
	LastName          string `json:"LastName"`
	PersonContactId   string `json:"PersonContactId"`
	PersonEmail       string `json:"PersonEmail"`
	PersonMobilePhone string `json:"PersonMobilePhone"`
	RecordTypeId      string `json:"RecordTypeId"`
}
