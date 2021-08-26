package models

type SfcContact struct {
	Id          string `json:"Id"`
	FirstName   string `json:"FirstName"`
	LastName    string `json:"LastName"`
	Email       string `json:"Email"`
	MobilePhone string `json:"MobilePhone"`
	// This field will be given to us by salesforce, to know if a user is blocked
	Blocked bool `json:"-"`
}
