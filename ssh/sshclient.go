package ssh

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	//"github.com/cloud-pi/kraken/pkg/scp"
	"github.com/moby/term"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	ssh.ClientConfig
	serverConnection *ssh.Client

	ServerAddress string // ip:port
	Password      string // password string
	KeyFile       string
}

func (client *Client) checkError(err error) error {
	if _, ok := err.(*ssh.ExitMissingError); !ok {
		return err
	}

	return nil
}

func (client *Client) signKey(KeyName string) (ssh.Signer, error) {
	pathToPrivateKey := filepath.FromSlash(fmt.Sprintf("%s", KeyName))

	key, err := ioutil.ReadFile(pathToPrivateKey)
	if err != nil {
		return nil, err
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	return signer, nil
}

func (client *Client) monWindowsChange(session *ssh.Session, fd uintptr) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGWINCH)
	defer signal.Stop(sigs)

	// resize the tty if any signals received
	for range sigs {
		session.SendRequest("window-change", false, client.termSize(fd))
	}
}

func (client *Client) termSize(fd uintptr) []byte {
	size := make([]byte, 16)

	winsize, err := term.GetWinsize(fd)
	if err != nil {
		binary.BigEndian.PutUint32(size, uint32(80))
		binary.BigEndian.PutUint32(size[4:], uint32(24))
		return size
	}

	binary.BigEndian.PutUint32(size, uint32(winsize.Width))
	binary.BigEndian.PutUint32(size[4:], uint32(winsize.Height))

	return size
}

func (client *Client) Connect() error {
	signer, err := client.signKey(client.KeyFile)
	if err != nil {
		return err
	}

	client.Auth = []ssh.AuthMethod{
		//ssh.Password(client.Password),
		ssh.PublicKeys(signer),
	}
	client.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	/*
		addrs := strings.Split(client.ServerAddress, "::")
		proto := addrs[0]
		addr := addrs[1]
	*/

	c, err := ssh.Dial("tcp", client.ServerAddress, &client.ClientConfig)
	if err != nil {
		return err
	}

	client.serverConnection = c
	return nil
}

func (client *Client) Close() {
	if client.serverConnection != nil {
		client.serverConnection.Close()
	}
}

func (client *Client) newSession(c *ssh.Client) (*ssh.Session, error) {
	session, err := c.NewSession()
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (client *Client) sendCommand(c *ssh.Client, cmd string) (string, error) {
	// Create a session. It is one session per command.
	session, err := client.newSession(c)
	if err != nil {
		return "", err
	}

	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	// you can also pass what gets input to the stdin, allowing you to pipe
	// content from client to server
	//      session.Stdin = bytes.NewBufferString("My input")

	// Finally, run the command
	//_, err = session.Output(cmd)

	if err := client.checkError(session.Run(cmd)); err != nil {
		return "", err
	}

	return b.String(), nil
}

func (client *Client) SendCommand(cmd string) (string, error) {
	return client.sendCommand(client.serverConnection, cmd)
}

// https://github.com/dtylman/scp
// https://github.com/EugenMayer/go-sshclient/blob/master/scpwrapper/to_remote.go
func (client *Client) ScpWithRateLimit(dst, src string, rateKbps int32) (int64, time.Duration, error) {
	session, err := client.newSession(client.serverConnection)
	if err != nil {
		return 0, 0, err
	}
	defer session.Close()

	f, err := os.Open(src)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	s, err := f.Stat()
	if err != nil {
		return 0, 0, err
	}

	size := s.Size()
	mode := s.Mode()
	fileName := path.Base(src)

	w, err := session.StdinPipe()
	if err != nil {
		return 0, 0, err
	}
	defer w.Close()

	cmd := "scp -t " + dst

	if err := session.Start(cmd); err != nil {

		w.Close()
		return 0, 0, err
	}

	errors := make(chan error)

	go func() {
		errors <- session.Wait()
	}()

	start := time.Now()

	var reader io.Reader
	reader = f
	/*
		// rate limiter
		if rateKbps != 0 {
			perSize := int64(1 * 1024) // 1K buf
			bucket := ratelimit.NewBucketWithRate(float64(rateKbps*1024), perSize)
			reader = ratelimit.Reader(f, bucket)
		}
	*/

	fmt.Fprintf(w, "C%#o %d %s\n", mode, size, fileName)
	io.Copy(w, reader)
	w.Write([]byte{0x00})
	w.Close()

	err = <-errors
	if err1 := client.checkError(err); err1 != nil {
		return 0, 0, err1
	}

	dur := time.Since(start)

	return size, dur, nil
}

func (client *Client) Upload(localFileName, remoteFileName string) error {
	/*
		// povsister scp
		scpClient, err := scp.NewClientFromExistingSSH(client.serverConnection, &scp.ClientOption{})
		if err != nil {
			return err
		}
		defer scpClient.Close()

		if stat, err1 := os.Stat(localFileName); err != nil {
			return err1
		} else if stat.IsDir() {
			err = scpClient.CopyDirToRemote(localFileName, remoteFileName, &scp.DirTransferOption{
				PreserveProp: true,
			})
		} else {
			err = scpClient.CopyFileToRemote(localFileName, remoteFileName, &scp.FileTransferOption{
				PreserveProp: true,
			})
		}

		return err
	*/

	return nil
}

func checkDestDir(remoteBase, localDirName string) (string, error) {
	stat, err := os.Stat(localDirName)
	if stat != nil {
		if !stat.IsDir() {
			return "", fmt.Errorf("scp: %s is not a directory", localDirName)
		}

		// append the base of remote
		localDirName = filepath.Join(localDirName, remoteBase)
	} else {
		parent := filepath.Dir(localDirName)
		stat, err = os.Stat(parent)
		if err != nil || !stat.IsDir() {
			return "", fmt.Errorf("scp: %s is not a directory", localDirName)
		}
	}

	if _, err1 := os.Stat(localDirName); err1 != nil {
		if err1 = os.Mkdir(localDirName, os.FileMode(os.ModePerm)); err1 != nil {
			return "", fmt.Errorf("scp: %s", err1)
		}
	}

	return localDirName, nil
}

func (client *Client) Download(remoteFileName, localFileName string) error {
	/*
			script := `
		#!/bin/bash
		NAME=%s
		if [ -d $NAME ]; then
			echo -n "directory"
		elif [ -f $NAME ]; then
			echo -n "file"
		else
			echo -n "not exist"
		fi
		`
			out, err := client.SendCommand(fmt.Sprintf(script, remoteFileName))
			if err != nil {
				return fmt.Errorf("scp: %v", err)
			}

			scpClient, err := scp.NewClientFromExistingSSH(client.serverConnection, &scp.ClientOption{})
			if err != nil {
				return err
			}
			defer scpClient.Close()

			switch out {
			case "directory":
				base := path.Base(remoteFileName)
				localDir, err := checkDestDir(base, localFileName)
				if err != nil {
					return err
				}

				err = scpClient.CopyDirFromRemote(remoteFileName, localDir, &scp.DirTransferOption{
					PreserveProp: true,
				})

			case "file":
				err = scpClient.CopyFileFromRemote(remoteFileName, localFileName, &scp.FileTransferOption{
					PreserveProp: true,
				})
			case "not exist":
				return fmt.Errorf("scp: %s no such file or directory on the remote server", remoteFileName)
			default:
				return fmt.Errorf("scp: unknown file type for %s", remoteFileName)
			}

			return err
	*/

	return nil
}

func (client *Client) Shell() error {
	session, err := client.newSession(client.serverConnection)
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO: 1,
	}

	fd := os.Stdin.Fd()

	termName := os.Getenv("TERM")
	if len(termName) < 1 {
		termName = "xterm-256color"
		//termName := "xterm"
	}
	termWidth, termHeight := 80, 24

	if term.IsTerminal(fd) {
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			return err
		}

		defer term.RestoreTerminal(fd, oldState)

		winsize, err := term.GetWinsize(fd)
		if err == nil {
			termWidth = int(winsize.Width)
			termHeight = int(winsize.Height)
		}
	}

	if err := session.RequestPty(termName, termHeight, termWidth, modes); err != nil {
		return err
	}

	if err := session.Shell(); err != nil {
		return err
	}

	// monitor for sigwinch
	go client.monWindowsChange(session, os.Stdout.Fd())

	err = session.Wait()

	return client.checkError(err)
}
