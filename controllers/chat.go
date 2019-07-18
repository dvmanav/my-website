package controllers

import (
	"github.com/Dev-ManavSethi/my-website/models"
	"github.com/Dev-ManavSethi/my-website/utils"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net/http"
	"time"
)



func Chat(w http.ResponseWriter, r *http.Request){


	if r.Method==http.MethodGet {

		IPAddress := utils.GetUserIP(r)


		UserExists := utils.CheckChatUserExists(IPAddress)

		if UserExists{
			err:= models.Templates.ExecuteTemplate(w, "chat.html", models.Chats[IPAddress])
			utils.HandleErr(err, "Error executing chat.html for " + IPAddress, "")
			if err!=nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else{


			utils.RegisterChatUser(IPAddress, "Guest")

			error := utils.BackupChats()
			if error!=nil{
				log.Println("Unable to backup chats")
				log.Println(error)
			}


			err := models.Templates.ExecuteTemplate(w, "chat.html", models.Chats[IPAddress])
			utils.HandleErr(err, "Error executing chat.html for " + IPAddress, "")
			if err!=nil {
				w.WriteHeader(http.StatusInternalServerError)
			}





		}


	}



	if r.Method == http.MethodPost{

		err := r.ParseForm()
		if err!=nil{
			//handle error
		}

		name := r.FormValue("name")
		message := r.FormValue("message")
		time := time.Now().Unix()

		IncomingMessage := models.ChatMessage{
			Name:name,
			Message:message,
			Time:time,
		}

		IPAddress := utils.GetUserIP(r)

		UserExists := utils.CheckChatUserExists(IPAddress)

		if UserExists{

		User := models.Chats[IPAddress]

		User.Chats = append(User.Chats, IncomingMessage)

		models.Chats[IPAddress]= User
			error := utils.BackupChats()
			if error!=nil{
				log.Println("Unable to Backup chats")
				log.Fatalln(error)
			}


		} else {

			utils.RegisterChatUser(IPAddress, name)
			User := models.Chats[IPAddress]

			User.Chats = append(User.Chats, IncomingMessage)

			models.GlobalMutex.Lock()
			models.Chats[IPAddress]= User
			models.GlobalMutex.Unlock()

			error := utils.BackupChats()
			if error!=nil{

				log.Println("Unable to Backup chats")
			log.Fatalln(error)
			}






		}
	}



}

func ChatWS(ws *websocket.Conn){

	var IncomingMessage models.ChatMessage

	for {
		err := websocket.JSON.Receive(ws, &IncomingMessage)
		if err==io.EOF{

		} else if err!=io.EOF && err!=nil{
			utils.HandleErr(err, "Error recieving Chat message from ip: "+IncomingMessage.IP, "Recieved chat message from ip: "+IncomingMessage.IP)
		} else {
//save to db
			IncomingMessage.Time = time.Now().Unix()
			User := models.Chats[IncomingMessage.IP]
			User.Chats = append(User.Chats, IncomingMessage)

			models.GlobalMutex.Lock()
			models.Chats[IncomingMessage.IP] = User
			models.GlobalMutex.Unlock()

			err := utils.BackupChats()
			if err != nil {

			} else {

				err2 := websocket.JSON.Send(ws, IncomingMessage)
				if err2 != nil {

				} else {

				}

			}

			//send reply
			var OutgoingMessage models.ChatMessage
			OutgoingMessage.Name = "Manav"
			OutgoingMessage.Message = "Hi! This is automated message"
			OutgoingMessage.Time = time.Now().Unix()
			OutgoingMessage.IP = "0.0.0.0"

			err2 := websocket.JSON.Send(ws, OutgoingMessage)
			if err2 != nil {

			} else {

				//save to db
				User := models.Chats[IncomingMessage.IP]
				User.Chats = append(User.Chats, OutgoingMessage)

				models.GlobalMutex.Lock()
				models.Chats[IncomingMessage.IP] = User
				models.GlobalMutex.Unlock()

				err := utils.BackupChats()
				if err != nil {

				} else {

				}
			}

		}

	}

}
