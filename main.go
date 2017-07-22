package main

import "github.com/aws/aws-sdk-go/aws/session"
import "fmt"

func main() {
	sess := session.Must(session.NewSession())

	if sess != nil {
		fmt.Println("yay")
	}
}
