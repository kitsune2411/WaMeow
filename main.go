package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "github.com/mattn/go-sqlite3"

	"github.com/mdp/qrterminal/v3"

	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var client *whatsmeow.Client

func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		if !v.Info.IsFromMe {
			fmt.Println("GetConversation : ", v.Message.GetConversation())
			fmt.Println("Sender : ", v.Info.Sender)
			fmt.Println("Sender Number : ", v.Info.Sender.User)
			fmt.Println("IsGroup : ", v.Info.IsGroup)
			fmt.Println("MessageSource : ", v.Info.MessageSource)
			fmt.Println("ID : ", v.Info.ID)
			fmt.Println("PushName : ", v.Info.PushName)
			fmt.Println("BroadcastListOwner : ", v.Info.BroadcastListOwner)
			fmt.Println("Category : ", v.Info.Category)
			fmt.Println("Chat : ", v.Info.Chat)
			fmt.Println("DeviceSentMeta : ", v.Info.DeviceSentMeta)
			fmt.Println("IsFromMe : ", v.Info.IsFromMe)
			fmt.Println("MediaType : ", v.Info.MediaType)
			fmt.Println("Multicast : ", v.Info.Multicast)
			fmt.Println("Info.Chat.Server : ", v.Info.Chat.Server)
			fmt.Println("livelocation : ", v.Message.LiveLocationMessage)
			if v.Info.Chat.Server == "g.us" {
				groupInfo, err := client.GetGroupInfo(v.Info.Chat)
				fmt.Println("error GetGroupInfo : ", err)
				fmt.Println("Nama Group : ", groupInfo.GroupName.Name)
			}

			fmt.Println("Message : ", *v.RawMessage.ExtendedTextMessage.Text)

			textwa := *v.RawMessage.ExtendedTextMessage.Text

			targetJID := types.JID{
				Server: v.Info.Sender.Server,
				User:   v.Info.Sender.User,
			}

			if !v.Info.IsGroup {
				if strings.Contains(textwa, "ini") {
					client.SendMessage(context.Background(), targetJID, &waProto.Message{
						// Conversation: proto.String("Hello, World!"),
						Conversation: proto.String("halo " + v.Info.PushName + ", apa ini ?"),
					})
				} else {

					client.SendMessage(context.Background(), targetJID, &waProto.Message{
						// Conversation: proto.String("Hello, World!"),
						Conversation: proto.String("halo " + v.Info.PushName + ", ngapain kamu bilang " + textwa + " ?"),
					})
				}
			}

			if v.Info.IsGroup {
				if v.Info.Chat.Server == "g.us" {
					groupInfo, err := client.GetGroupInfo(v.Info.Chat)
					fmt.Println("error GetGroupInfo : ", err)
					fmt.Println("Nama Group : ", groupInfo.GroupName.Name)

					groupJID := types.JID{
						Server: groupInfo.JID.Server,
						User:   groupInfo.JID.User,
					}

					namGrup := groupInfo.GroupName.Name

					if namGrup == "Test" {
						client.SendMessage(context.Background(), groupJID, &waProto.Message{
							// Conversation: proto.String("Hello, World!"),
							Conversation: proto.String("halo " + v.Info.PushName + ", apa ini ?"),
						})
					}
				}
			}

		}
	}
}

func main() {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite
	container, err := sqlstore.New("sqlite3", "file:gowa.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client = whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				fmt.Println("QR code:", evt.Code)
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}
