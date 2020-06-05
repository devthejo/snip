package main

import "fmt"

func Apply(cmd string) string {
	return fmt.Sprintf("sudo %v", cmd)
}
