package main

import "github.com/ssh-connection-manager/kernel/v2/pkg/connect"

// main for success connect start make command: up-ssh-port
func main() {
	connection := &connect.Connect{
		Alias:     "test",
		Login:     "root",
		Password:  "password",
		Address:   "localhost",
		Type:      connect.TypeSSH,
		CreatedAt: "",
		UpdatedAt: "",
		SshOptions: &connect.SshOptions{
			Port:       2222,
			PrivateKey: "",
		},
	}

	sshConnect := connect.NewSshConnector()
	session, err := sshConnect.NewSession(connection)
	if err != nil {
		panic(err)
	}

	err = sshConnect.Connect(session)
	if err != nil {
		panic(err)
	}
}
