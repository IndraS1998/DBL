package custom_test

import (
	"raft/utils"

	"math/rand"
	"time"
)

// GenerateRandomString generates a random string of the specified length containing only letters
func generateRandomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator

	result := make([]byte, length)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}
func generateCreateAccountData() utils.UserPayload {
	payload := utils.UserPayload{
		FirstName:                generateRandomString(5),
		LastName:                 generateRandomString(7),
		HashedPassword:           generateRandomString(8),
		Email:                    generateRandomString(8),
		DateOfBirth:              time.Date(2002, time.January, 25, 0, 0, 0, 0, time.UTC),
		IdentificationNumber:     generateRandomString(9),
		IdentificationImageFront: generateRandomString(33),
		IdentificationImageBack:  generateRandomString(32),
		PrevPW:                   generateRandomString(8),
		NewPW:                    generateRandomString(8),
		UserID:                   -1,
		Action:                   utils.UserCreateAccount,
	}
	return payload
}
func MockAPIRequest() []utils.Payload {
	results := make([]utils.Payload, 2)
	for i := range results {
		results[i] = generateCreateAccountData()
	}
	return results
}
