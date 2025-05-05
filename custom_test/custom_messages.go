package custom_test

import (
	"raft/utils"

	"math/rand"
	"time"
)

type MockUserReq struct {
	RefTable utils.RefTable
	*utils.UserPayload
	*utils.AdminPayload
	*utils.WalletOperationPayload
}

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
func generateCreateAccountData() MockUserReq {
	payload := utils.UserPayload{
		FirstName:                generateRandomString(5),
		LastName:                 generateRandomString(7),
		HashedPassword:           generateRandomString(8),
		Email:                    generateRandomString(8),
		DateOfBirth:              time.Date(2002, time.January, 25, 0, 0, 0, 0, time.UTC),
		IdentificationNumber:     generateRandomString(9),
		IdentificationImageFront: generateRandomString(33),
		IdentificationImageBack:  generateRandomString(32),
		PrevPW:                   "",
		NewPW:                    "",
		UserID:                   -1,
		Action:                   utils.UserCreateAccount,
	}
	req := MockUserReq{
		RefTable:               utils.RefUser,
		UserPayload:            &payload,
		AdminPayload:           nil,
		WalletOperationPayload: nil,
	}
	return req
}
func MockAPIRequest() []MockUserReq {
	results := make([]MockUserReq, 2)
	for i := range results {
		results[i] = generateCreateAccountData()
	}
	return results
}
