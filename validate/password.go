package validate

import (
	"fmt"
	"regexp"
	"unicode"
)

// PasswordPolicy defines password requirements
type PasswordPolicy struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
	MaxLength      int
}

// DefaultPasswordPolicy returns the default password policy
func DefaultPasswordPolicy() *PasswordPolicy {
	return &PasswordPolicy{
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: true,
		MaxLength:      128,
	}
}

// PasswordValidator validates passwords against policy
type PasswordValidator struct {
	policy *PasswordPolicy
}

// NewPasswordValidator creates a new password validator
func NewPasswordValidator(policy *PasswordPolicy) *PasswordValidator {
	if policy == nil {
		policy = DefaultPasswordPolicy()
	}
	return &PasswordValidator{policy: policy}
}

// Validate validates a password against the policy
func (pv *PasswordValidator) Validate(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Check length
	if len(password) < pv.policy.MinLength {
		return fmt.Errorf("password must be at least %d characters long", pv.policy.MinLength)
	}

	if pv.policy.MaxLength > 0 && len(password) > pv.policy.MaxLength {
		return fmt.Errorf("password must not exceed %d characters", pv.policy.MaxLength)
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	// Check character requirements
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if pv.policy.RequireUpper && !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if pv.policy.RequireLower && !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if pv.policy.RequireNumber && !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	if pv.policy.RequireSpecial && !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	// Check for common weak passwords
	if isCommonPassword(password) {
		return fmt.Errorf("password is too common, please choose a stronger password")
	}

	return nil
}

// GetPolicyDescription returns a human-readable description of the policy
func (pv *PasswordValidator) GetPolicyDescription() string {
	desc := fmt.Sprintf("Password must be at least %d characters", pv.policy.MinLength)

	requirements := []string{}
	if pv.policy.RequireUpper {
		requirements = append(requirements, "one uppercase letter")
	}
	if pv.policy.RequireLower {
		requirements = append(requirements, "one lowercase letter")
	}
	if pv.policy.RequireNumber {
		requirements = append(requirements, "one number")
	}
	if pv.policy.RequireSpecial {
		requirements = append(requirements, "one special character")
	}

	if len(requirements) > 0 {
		desc += " and contain "
		for i, req := range requirements {
			if i == len(requirements)-1 && i > 0 {
				desc += " and "
			} else if i > 0 {
				desc += ", "
			}
			desc += req
		}
	}

	if pv.policy.MaxLength > 0 {
		desc += fmt.Sprintf(" (maximum %d characters)", pv.policy.MaxLength)
	}

	return desc
}

// isCommonPassword checks if password is in the list of common weak passwords
func isCommonPassword(password string) bool {
	// List of most common passwords to block
	commonPasswords := []string{
		"password", "password123", "12345678", "qwerty", "123456789",
		"12345", "1234", "111111", "1234567", "dragon",
		"123123", "baseball", "iloveyou", "trustno1", "1234567890",
		"sunshine", "master", "123456789", "welcome", "shadow",
		"ashley", "football", "jesus", "michael", "ninja",
		"mustang", "password1", "admin", "administrator", "root",
	}

	passwordLower := regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(password, "")
	for _, common := range commonPasswords {
		if passwordLower == common {
			return true
		}
	}

	return false
}

// PasswordStrength returns a score from 0-4 indicating password strength
func (pv *PasswordValidator) PasswordStrength(password string) int {
	score := 0

	// Length score
	if len(password) >= 8 {
		score++
	}
	if len(password) >= 12 {
		score++
	}

	// Character variety score
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	charTypes := 0
	if hasUpper {
		charTypes++
	}
	if hasLower {
		charTypes++
	}
	if hasNumber {
		charTypes++
	}
	if hasSpecial {
		charTypes++
	}

	if charTypes >= 3 {
		score++
	}
	if charTypes == 4 {
		score++
	}

	// Check for patterns and common passwords
	if isCommonPassword(password) {
		score = 0
	}

	return score
}
