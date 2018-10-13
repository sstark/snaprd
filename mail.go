package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
)

func FailureMail(exitCode int, logBuffer *RingIO) {
	mail := fmt.Sprintf("snaprd exited with return value %d.\nLatest log output:\n\n%s",
		exitCode, logBuffer.GetAsText())
	subject := fmt.Sprintf("snaprd failure (origin: %s)", config.Origin)
	SendMail(config.Notify, subject, mail)
}

func NotifyMail(to, msg string) {
	SendMail(to, "snaprd notice", msg)
}

func SendMail(to, subject, msg string) {
	sendmail := exec.Command("mail", "-s", subject, to)
	stdin, err := sendmail.StdinPipe()
	if err != nil {
		log.Println(err)
		return
	}
	stdout, err := sendmail.StdoutPipe()
	if err != nil {
		log.Println(err)
		return
	}
	sendmail.Start()
	stdin.Write([]byte(msg))
	stdin.Write([]byte("\n"))
	stdin.Close()
	ioutil.ReadAll(stdout)
	sendmail.Wait()
	log.Printf("sending notification to %s done\n", to)
}
