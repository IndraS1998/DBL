package custom_test

import (
	"fmt"
	"math/rand"
	"strconv"
)

var msg_str string = "RPC  message"

func GenerateMessage() string {
	randN := rand.Intn(10000)
	return fmt.Sprintf("%s %s", msg_str, strconv.Itoa(randN))
}
