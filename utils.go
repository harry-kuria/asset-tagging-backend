package main

import (
	"crypto/rand"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// checkPassword compares a password with its hash
func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateRandomPassword generates a random password
func generateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)
	
	for i := range password {
		randomByte := make([]byte, 1)
		_, err := rand.Read(randomByte)
		if err != nil {
			return "", err
		}
		password[i] = charset[randomByte[0]%byte(len(charset))]
	}
	
	return string(password), nil
}

// numberToWords converts a number to words
func numberToWords(num int) string {
	if num == 0 {
		return "zero"
	}

	units := []string{"", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}
	teens := []string{"ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen"}
	tens := []string{"", "", "twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety"}

	if num < 10 {
		return units[num]
	} else if num < 20 {
		return teens[num-10]
	} else if num < 100 {
		if num%10 == 0 {
			return tens[num/10]
		}
		return tens[num/10] + " " + units[num%10]
	} else if num < 1000 {
		if num%100 == 0 {
			return units[num/100] + " hundred"
		}
		return units[num/100] + " hundred and " + numberToWords(num%100)
	} else if num < 1000000 {
		if num%1000 == 0 {
			return numberToWords(num/1000) + " thousand"
		}
		return numberToWords(num/1000) + " thousand " + numberToWords(num%1000)
	}

	return fmt.Sprintf("%d", num)
}

// formatCurrency formats a float as currency
func formatCurrency(amount float64) string {
	return fmt.Sprintf("₱%.2f", amount)
} 