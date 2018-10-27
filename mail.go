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

func RsyncIssueMail(rsyncError error, rsyncErrorCode int) {
	var errText string
	if s, ok := rsyncIgnoredErrors[rsyncErrorCode]; ok == true {
		errText = s
	} else {
		errText = "<unknown>"
	}
	mail := fmt.Sprintf(`rsync finished with error: %s (%s).
This is a non-fatal error, snaprd will try again.`, rsyncError, errText)
	subject := fmt.Sprintf("snaprd rsync error (origin: %s)", config.Origin)
	// In this case we care that the mail command is not blocking the whole
	// program
	go SendMail(config.Notify, subject, mail)
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
