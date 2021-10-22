package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"gitlab.com/wolfplus/gamespace-lobby/db"
	"gitlab.com/wolfplus/gamespace-lobby/helper"
	"gitlab.com/wolfplus/gamespace-lobby/ws"
)

//just pass to front end

const maxNumberToShow = 15

var lobbyClients = map[int]*ws.LobbySession{}
var adminInLobbyClients = map[int]*ws.LobbySession{}

var coinUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var mu sync.Mutex

func LobbyUpdateCustomerServiceMessage(uid int, data map[string]string) {
	if client, ok := lobbyClients[uid]; ok {
		client.Conn.WriteJSON(data)
	}
	for uid := range adminInLobbyClients {
		if admin, ok := adminInLobbyClients[uid]; ok {
			go func(lobbyClientConnection *websocket.Conn) {
				mu.Lock()
				defer mu.Unlock()
				// go lobby.Conn.WriteJSON(data)
				if err := lobbyClientConnection.WriteJSON(data); err != nil { // push
					log.Println(err)
					return
				}
			}(admin.Conn)
		}
	}
}

func broadcastJsonMessage(data map[string]interface{}) {
	for uid := range lobbyClients {
		if client, ok := lobbyClients[uid]; ok {
			go func(lobbyClientConnection *websocket.Conn) {
				mu.Lock()
				defer mu.Unlock()
				if err := lobbyClientConnection.WriteJSON(data); err != nil { // push
					log.Println(err)
					return
				}
			}(client.Conn)
		}
	}
}

func lobbySendRoomMessage(data map[string]interface{}, conn *websocket.Conn) {
	data["type"] = "lobbySendRoomMessage"
	// data["room"] = data["room"]
	data["message"] = helper.CheckForSensitiveWords(data["message"].(string))

	//lobby client
	broadcastJsonMessage(data)
}

func lobbySendChatMessage(data map[string]interface{}, conn *websocket.Conn) {
	data["type"] = "lobbySendChatMessage"
	data["message"] = helper.CheckForSensitiveWords(data["message"].(string))

	//lobby client
	broadcastJsonMessage(data)
}

func lobbySendClubMessage(data map[string]interface{}, conn *websocket.Conn) {
	clubId := "0"

	club_user := db.GetValidClubUserByUserId(fmt.Sprintf("%v", data["uid"]))
	if club_user != nil { // this player has a club
		clubId = club_user["club_id"]
	}
	if clubId == "0" {
		return
	}
	data["type"] = "lobbySendClubMessage"
	data["message"] = helper.CheckForSensitiveWords(data["message"].(string))

	for uid := range lobbyClients {
		if client, ok := lobbyClients[uid]; ok {
			go func(lobbyClient *ws.LobbySession, clubId string) {
				if lobbyClient.ClubID != clubId {
					return
				}
				if err := lobbyClient.Conn.WriteJSON(data); err != nil { // push
					log.Println(err)
					return
				}
			}(client, clubId)
		}
	}
}

func LobbyCoinUpdate(data map[string]interface{}) {
	// data: game_coin, uid
	if uid, ok := data["uid"]; ok {
		data["uid"] = helper.StringToInt64(fmt.Sprintf("%v", uid))

		if client, ok := lobbyClients[int(data["uid"].(int64))]; ok {
			client.Conn.WriteJSON(data)
		}
	}

	// for uid := range lobbyClients {
	// 	if client, ok := lobbyClients[uid]; ok {
	// 		if client.Conn == conn {
	// 			continue
	// 		}
	// 		go func(lobbyClientConnection *websocket.Conn) {
	// 			mu.Lock()
	// 			defer mu.Unlock()
	// 			if err := lobbyClientConnection.WriteJSON(data); err != nil { // push
	// 				log.Println(err)
	// 				return
	// 			}
	// 		}(client.Conn)
	// 	}
	// }
}

func lobbyLineOnChat() {
	data := map[string]interface{}{
		"type":       "lineOnChat",
		"lineOnChat": "Thêm Zalo CSKH 0569892415 nhận thưởng mỗi ngày!",
	}
	fmt.Println("data line:", data)
	for uid := range lobbyClients {
		fmt.Println(len(lobbyClients))
		if client, ok := lobbyClients[uid]; ok {
			go func(lobbyClientConnection *websocket.Conn) {
				mu.Lock()
				defer mu.Unlock()
				if err := lobbyClientConnection.WriteJSON(data); err != nil { // push
					log.Println(err)
					return
				}
			}(client.Conn)
		}
	}
}

func LobbyPurchaseUpdate(data map[string]interface{}) {
	// data: purchase_amount, uid
	fmt.Println("data:", data)
	if uid, ok := data["uid"]; ok {
		data["uid"] = helper.StringToInt64(fmt.Sprintf("%v", uid))

		if client, ok := lobbyClients[int(data["uid"].(int64))]; ok {
			client.Conn.WriteJSON(data)
		}
	}
}

func lobbyUpdateOnlinePlayers() {
	var numberToShow = 0
	if maxNumberToShow > len(lobbyClients) {
		numberToShow = len(lobbyClients)
	} else {
		numberToShow = maxNumberToShow
	}
	var lobbyClientsToShow = []ws.LobbySession{}
	for uid := range lobbyClients {
		if client, ok := lobbyClients[uid]; ok {
			lobbyClientsToShow = append(lobbyClientsToShow, *client)
			numberToShow--
			if numberToShow == 0 {
				break
			}
		}
	}
	data := map[string]interface{}{
		"type":               "lobbyUpdateOnlinePlayers",
		"lobbyClientsToShow": lobbyClientsToShow,
		"totalOnlinePlayers": len(lobbyClients),
	}
	for uid := range lobbyClients {
		if client, ok := lobbyClients[uid]; ok {
			go func(lobbyClientConnection *websocket.Conn) {
				mu.Lock()
				defer mu.Unlock()
				if err := lobbyClientConnection.WriteJSON(data); err != nil { // push
					log.Println(err)
					return
				}
			}(client.Conn)
		}
	}
}

func initLobbySession(data map[string]interface{}, conn *websocket.Conn) {
	claims, isValid := helper.VerifyJWTString(data["jwt"].(string))
	uidStr := fmt.Sprintf("%v", claims[helper.TOKENUIDKEY])
	errorResponse := map[string]interface{}{
		"code":    "err",
		"message": "",
	}
	if !isValid {
		errorResponse["message"] = "Client connection closed due to invalid token."
		conn.WriteJSON(errorResponse)
		conn.Close()
		return
	}
	uid, err := strconv.Atoi(uidStr)
	if err != nil {
		log.Println(err)
		errorResponse["message"] = "Client connection closed and not addded to session list."
		conn.WriteJSON(errorResponse)
		conn.Close()
		return
	}
	if _, ok := lobbyClients[uid]; ok { // user exists
		errorResponse["message"] = "Client connection closed due to already existing user."
		toCloseConn := lobbyClients[uid].Conn // kick out this connection
		toCloseConn.WriteJSON(errorResponse)
		toCloseConn.Close()
		// return // no need to return cuz u need to bind this guy
	}
	clubId := "0"
	club_user := db.GetValidClubUserByUserId(uidStr)
	if club_user != nil { // this player has a club
		clubId = club_user["club_id"]
	}
	levelAsString := "0"
	levelAndExp := db.GetExperienceAndLevel(uidStr)
	if levelAndExp != nil {
		levelAsString = levelAndExp["level"]
	}
	level, err := strconv.Atoi(levelAsString)
	if err != nil {
		log.Println(err)
	}
	// log.Println(level)
	newPlayer := &ws.LobbySession{
		Uid:      uidStr,
		Username: data["username"].(string),
		FaceUri:  data["faceUri"].(string),
		ClubID:   clubId,
		Level:    level,
		Conn:     conn,
	}

	lobbyClients[uid] = newPlayer
	go func(data map[string]interface{}, conn *websocket.Conn) {
		if db.CheckAdmin(uidStr) {
			adminInLobbyClients[uid] = newPlayer
		}
	}(data, conn)

	//Update online players
	lobbyUpdateOnlinePlayers()

	lobbyLineOnChat()
}

func coinWsEndpoint(w http.ResponseWriter, r *http.Request) {
	conn, err := coinUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Client Successfully Connected")

	coinReader(conn) // open and maintains a connection to the client
}

func coinReader(conn *websocket.Conn) {
	for {
		var socketRequest *ws.SocketRequest = &ws.SocketRequest{}
		err := conn.ReadJSON(socketRequest)
		if err != nil { // disconnected somehow
			for uid := range lobbyClients {
				if lobby, ok := lobbyClients[uid]; ok {
					if lobby.Conn == conn {
						coinRemoveClient(uid)
						lobbyUpdateOnlinePlayers()
						break
					}
				}
			}
			break
		}

		switch socketRequest.Type {

		case "sendRoomMessage":
			lobbySendRoomMessage(socketRequest.Data, conn)

		case "sendClubMessage":
			lobbySendClubMessage(socketRequest.Data, conn)

		case "sendChatMessage":
			lobbySendChatMessage(socketRequest.Data, conn)

		case "coinUpdate":
			LobbyCoinUpdate(socketRequest.Data)

		case "amountUpdate":
			LobbyPurchaseUpdate(socketRequest.Data)

		case "init":
			initLobbySession(socketRequest.Data, conn)
			break
		}
	}
}

func coinRemoveClient(index int) {
	mu.Lock()
	defer mu.Unlock()
	delete(lobbyClients, index)
}

func MakeCoinWsService() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/coinws/net", coinWsEndpoint)

	return router
}
