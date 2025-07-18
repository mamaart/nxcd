package notifier

import (
	"context"
	"log"
	"os"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

type Notifier interface {
	Notify(s string)
}

type notifier struct {
	ch chan<- string
}

func FromEnv() (Notifier, error) {
	h := os.Getenv("MATRIX_HOMESERVER")
	u := os.Getenv("MATRIX_USERNAME")
	t := os.Getenv("MATRIX_PASSWORD")
	r := os.Getenv("MATRIX_ROOMID")

	if h == "" || u == "" || r == "" || t == "" {
		panic("some environment variable for matrix notifier is missing")
	}

	return New(h, u, r, t)
}

func New(homeserver, username, password, roomId string) (Notifier, error) {
	ch := make(chan string)
	cli, err := mautrix.NewClient(homeserver, "", "")
	if err != nil {
		return nil, err
	}

	if _, err := cli.Login(context.Background(), &mautrix.ReqLogin{
		Type: mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{
			Type: mautrix.IdentifierTypeUser,
			User: username,
		},
		Password:         password,
		StoreCredentials: true,
	}); err != nil {
		log.Fatal(err)
	}

	go func(cli *mautrix.Client, roomId string) {
		for s := range ch {
			log.Println(s)
			if _, err := cli.SendText(
				context.Background(),
				id.RoomID(roomId),
				s,
			); err != nil {
				log.Println(err)
			}
		}
	}(cli, roomId)
	return &notifier{ch}, nil
}

func (n *notifier) Notify(s string) {
	go func(s string) {
		n.ch <- s
	}(s)
}
