// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword() {
	logger.Info("starting password hashing function")

	if len(os.Args) < 3 {
		logger.Error("please provide a password: <app> hash <password>")
		return
	}

	pass := strings.Join(os.Args[2:], " ")

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(pass),
		bcrypt.DefaultCost,
	)
	if err != nil {
		logger.Error("failed to hash password",
			"err", err,
		)
		return
	}

	fmt.Println("")
	fmt.Println(string(hash))
}
