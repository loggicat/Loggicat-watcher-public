package util

import (
	"fmt"

	ct "github.com/daviddengcn/go-colortext"
	log "github.com/sirupsen/logrus"
)

func PrintRed(s string) {
	ct.Foreground(ct.Red, false)
	fmt.Println("[-] ", s)
	ct.ResetColor()
	log.Error(s)
}

func PrintRedFatal(s string) {
	ct.Foreground(ct.Red, false)
	fmt.Println("[-] ", s)
	ct.ResetColor()
	log.Fatal(s)
}

func PrintGreen(s string) {
	ct.Foreground(ct.Green, false)
	fmt.Println("[+] ", s)
	ct.ResetColor()
	log.Info(s)
}
