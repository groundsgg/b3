package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Password: ")
	password, _ := reader.ReadString('\n')

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(strings.TrimSpace(password)),
		bcrypt.DefaultCost,
	)
	if err != nil {
		logger.Error("failed to hash password",
			"err", err,
		)
		return
	}

	fmt.Print("Hash: ")
	fmt.Println(string(hash))
}
