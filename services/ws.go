package services

import (
	"encoding/json"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type Client struct {
	UserID uint
	Name   string
	Role   string
	Conn   *websocket.Conn
}

type ExamRoom struct {
	ChairmanID uint
	Quorum     int
	Chairman   *Client
	Examiners  map[uint]*Client
	Connected  map[uint]bool
	Progress   map[uint]int
	Mutex      sync.Mutex
}

var examRooms = make(map[string]*ExamRoom)
var globalMutex sync.Mutex

func WebSocketHandler(c *websocket.Conn) {
	defer c.Close()

	var (
		room   *ExamRoom
		client *Client
	)

	var initDone bool

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}

		// === 🔒 ВСТАВИТЬ ИНИЦИАЛИЗАЦИЮ КЛИЕНТА ПОСЛЕ СООБЩЕНИЯ type: init_user ===
		if !initDone {
			var init struct {
				Type string `json:"type"`
				Data struct {
					ExamID string `json:"exam_id"`
					UserID uint   `json:"user_id"`
					Name   string `json:"name"`
					Role   string `json:"role"`
				} `json:"data"`
			}

			if err := json.Unmarshal(msg, &init); err != nil || init.Type != "init_user" {
				continue // пропускаем всё, что не init
			}

			examID := init.Data.ExamID
			userID := init.Data.UserID
			role := init.Data.Role
			name := init.Data.Name

			client = &Client{
				UserID: userID,
				Name:   name,
				Role:   role,
				Conn:   c,
			}

			globalMutex.Lock()
			room, exists := examRooms[examID]
			if !exists {
				room = &ExamRoom{
					Examiners: make(map[uint]*Client),
					Connected: make(map[uint]bool),
					Progress:  make(map[uint]int),
					Quorum:    2, // можно позже получать из БД
				}
				examRooms[examID] = room
			}
			globalMutex.Unlock()

			room.Mutex.Lock()
			if role == "chairman" {
				room.ChairmanID = client.UserID
				room.Chairman = client
			} else {
				room.Examiners[client.UserID] = client
			}
			room.Connected[client.UserID] = true
			room.Mutex.Unlock()

			notifyChairmanAboutConnected(room)

			initDone = true
			continue
		}

		// === ✅ ОБЫЧНАЯ ОБРАБОТКА СООБЩЕНИЙ ===
		handleWSMessage(room, msg, client)
	}

	// === 📤 УДАЛЕНИЕ ПОЛЬЗОВАТЕЛЯ ПОСЛЕ ОТКЛЮЧЕНИЯ ===
	if room == nil || client == nil {
		return
	}

	room.Mutex.Lock()
	delete(room.Connected, client.UserID)
	delete(room.Examiners, client.UserID)
	room.Mutex.Unlock()

	notifyChairmanAboutConnected(room)

}

func notifyChairmanAboutConnected(room *ExamRoom) {
	if room.Chairman == nil {
		return
	}
	list := []string{}
	for _, examiner := range room.Examiners {
		if room.Connected[examiner.UserID] {
			list = append(list, examiner.Name)
		}
	}
	room.Chairman.Conn.WriteJSON(map[string]interface{}{
		"type": "connected_list",
		"data": list,
	})
}

func handleWSMessage(room *ExamRoom, msg []byte, sender *Client) {
	var data map[string]interface{}
	if err := json.Unmarshal(msg, &data); err != nil {
		return
	}

	switch data["type"] {
	case "start_exam":
		room.Mutex.Lock()
		if len(room.Connected) < room.Quorum {
			sender.Conn.WriteJSON(map[string]string{
				"type": "error", "data": "Недостаточно участников для кворума",
			})
			room.Mutex.Unlock()
			return
		}
		for _, examiner := range room.Examiners {
			examiner.Conn.WriteJSON(map[string]string{
				"type": "start_exam", "data": "ok",
			})
		}
		room.Mutex.Unlock()

	case "select_student":
		studentID := data["data"].(map[string]interface{})["studentId"].(string)
		for _, examiner := range room.Examiners {
			examiner.Conn.WriteJSON(map[string]interface{}{
				"type": "open_student",
				"data": map[string]string{"studentId": studentID},
			})
		}

	case "submit_grade":
		sid := uint(data["data"].(map[string]interface{})["studentId"].(float64))
		room.Mutex.Lock()
		room.Progress[sid]++
		progress := room.Progress[sid]
		room.Mutex.Unlock()

		if room.Chairman != nil {
			room.Chairman.Conn.WriteJSON(map[string]interface{}{
				"type": "progress_update",
				"data": map[string]interface{}{
					"studentId":       sid,
					"currentProgress": progress,
				},
			})
		}
	}
}
