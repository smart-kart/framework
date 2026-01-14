package contact

import (
	"regexp"
	"strings"
)

// ContactType represents the type of contact information
type ContactType string

const (
	ContactTypeEmail ContactType = "email"
	ContactTypePhone ContactType = "phone"
)

// Regex patterns for validation
var (
	// EmailRegex is a simple regex for validating email format
	EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// PhoneRegex matches phone numbers with optional + prefix and 7-15 digits
	PhoneRegex = regexp.MustCompile(`^\+?[0-9]{7,15}$`)
)

// MaskContact masks email or phone for display purposes
func MaskContact(input string, contactType ContactType) string {
	if contactType == ContactTypeEmail {
		return maskEmail(input)
	}
	return maskPhone(input)
}

// maskEmail masks email address (e.g., "john.doe@example.com" -> "j***@example.com")
func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	localPart := parts[0]
	domain := parts[1]

	if len(localPart) <= 1 {
		return email
	}

	// Show first character + *** + @domain
	masked := string(localPart[0]) + "***@" + domain
	return masked
}

// maskPhone masks phone number (e.g., "+919876543210" -> "+91-9876***210")
func maskPhone(phone string) string {
	// Remove any spaces or hyphens
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	if len(phone) < 7 {
		return phone
	}

	// If starts with +, preserve country code
	if strings.HasPrefix(phone, "+") {
		if len(phone) <= 7 {
			return phone
		}
		// Show +CC-XXXX***XXX format
		countryCode := phone[:3]           // e.g., +91
		firstDigits := phone[3:7]          // First 4 digits after country code
		lastDigits := phone[len(phone)-3:] // Last 3 digits
		return countryCode + "-" + firstDigits + "***" + lastDigits
	}

	// Without country code: show first 4 and last 3
	if len(phone) > 7 {
		firstDigits := phone[:4]
		lastDigits := phone[len(phone)-3:]
		return firstDigits + "***" + lastDigits
	}

	return phone
}

// NormalizePhone removes spaces, hyphens, and ensures clean format
func NormalizePhone(phone string) string {
	// Remove spaces and hyphens
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.TrimSpace(phone)

	return phone
}

// NormalizeEmail converts email to lowercase and trims whitespace
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// DetectContactType determines if the input is an email or phone number
func DetectContactType(input string) (ContactType, bool) {
	input = strings.TrimSpace(input)

	// Check if it matches email pattern
	if EmailRegex.MatchString(input) {
		return ContactTypeEmail, true
	}

	// Check if it matches phone pattern
	if PhoneRegex.MatchString(input) {
		return ContactTypePhone, true
	}

	// Invalid format
	return "", false
}
