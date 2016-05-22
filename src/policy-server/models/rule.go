package models

import "errors"

type Rule struct {
	Source      string `json:"group1"`
	Destination string `json:"group2"`
}

func (r Rule) Equals(otherRule Rule) bool {
	return r.Source == otherRule.Source &&
		r.Destination == otherRule.Destination
}

func (r Rule) Validate() error {
	ok := (r.Source != "" && r.Destination != "")
	if !ok {
		return errors.New("missing required field(s)")
	}
	return nil
}
