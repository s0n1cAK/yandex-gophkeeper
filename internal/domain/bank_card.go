package domain

import "strings"

type BankCardPayload struct {
	Number string `json:"number"`
	Expiry string `json:"expiry"`
	Holder string `json:"holder,omitempty"`
	CVC    string `json:"cvc,omitempty"`
}

func (p BankCardPayload) NormalizeNumber() string {
	s := strings.TrimSpace(p.Number)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	return s
}

func (p BankCardPayload) Validate() error {
	num := p.NormalizeNumber()
	if num == "" {
		return ErrEmptyCardNumber
	}

	for _, r := range num {
		if r < '0' || r > '9' {
			return ErrInvalidCardNumberFormat
		}
	}

	if !LuhnValid(num) {
		return ErrInvalidBankCardNumber
	}

	if strings.TrimSpace(p.Expiry) == "" {
		return ErrEmptyCardExpiry
	}

	return nil
}
