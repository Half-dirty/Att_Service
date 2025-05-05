// Новый ws.go с поддержкой открытия студента для экзаменаторов

package services

import (
	"att_service/database"
	"att_service/models"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

type Client struct {
	UserID uint
	Name   string
	Role   string
	Conn   *websocket.Conn
}

type ExaminerInfo struct {
	ID     uint
	Name   string
	Avatar string
}

type ExamRoom struct {
	ChairmanID   uint
	Chairman     *Client
	Examiners    map[uint]*Client
	ExaminerList []ExaminerInfo
	Connected    map[uint]bool
	Progress     map[uint]int // studentID -> number of votes
	Quorum       int
	Mutex        sync.Mutex
}

var examRooms = make(map[int]*ExamRoom)
var globalMutex sync.Mutex

func WebSocketHandler(c *websocket.Conn) {
	defer c.Close()

	var room *ExamRoom
	var client *Client

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}

		var incoming struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}

		if err := json.Unmarshal(msg, &incoming); err != nil {
			log.Println("Ошибка парсинга сообщения:", err)
			continue
		}

		switch incoming.Type {
		case "init_user":
			var data struct {
				ExamID int    `json:"exam_id"`
				UserID uint   `json:"user_id"`
				Name   string `json:"name"`
				Role   string `json:"role"`
			}
			if err := json.Unmarshal(incoming.Data, &data); err != nil {
				log.Println("Ошибка парсинга init_user:", err)
				continue
			}
			r, cl := handleInitUser(c, data.ExamID, data.UserID, data.Name)
			r.Mutex.Lock()
			r.Connected[cl.UserID] = true
			r.Mutex.Unlock()
			room = r
			client = cl
			go startPingPong(room)
			broadcastExaminerList(room)
			broadcastChairmanStatus(room)

		case "request_examiner_list":
			var data struct {
				ExamID int `json:"exam_id"`
			}
			if err := json.Unmarshal(incoming.Data, &data); err != nil {
				log.Println("Ошибка парсинга request_examiner_list:", err)
				continue
			}
			handleRequestExaminerList(c, data.ExamID)

		case "start_exam":
			var data struct {
				ExamID int `json:"exam_id"`
			}
			if err := json.Unmarshal(incoming.Data, &data); err != nil {
				log.Println("Ошибка парсинга start_exam:", err)
				continue
			}
			handleStartExam(c, data.ExamID)

		case "open_student":
			var data struct {
				ExamID    int  `json:"exam_id"`
				StudentID uint `json:"student_id"`
			}
			if err := json.Unmarshal(incoming.Data, &data); err != nil {
				log.Println("Ошибка парсинга open_student:", err)
				continue
			}
			handleOpenStudent(c, data.ExamID, data.StudentID)

		case "progress_update":
			var data struct {
				ExamID          int  `json:"exam_id"`
				StudentID       uint `json:"student_id"`
				CurrentProgress int  `json:"current_progress"`
			}
			if err := json.Unmarshal(incoming.Data, &data); err != nil {
				log.Println("Ошибка парсинга progress_update:", err)
				continue
			}
			handleProgressUpdate(c, data.ExamID, data.StudentID, data.CurrentProgress)

		case "save_grade":
			var data struct {
				ExamID          uint   `json:"exam_id"`
				StudentID       uint   `json:"student_id"`
				Scores          []int  `json:"scores"`          // баллы (массив)
				Qualification   string `json:"qualification"`   // квалификация
				Recommendations string `json:"recommendations"` // рекомендации
				Specialization  string `json:"specialization"`  // специализация
				Abstained       bool   `json:"abstained"`       // флаг воздержания
			}
			if err := json.Unmarshal(incoming.Data, &data); err != nil {
				log.Println("Ошибка разбора save_grade:", err)
				continue
			}

			if room == nil || client == nil {
				log.Println("Нет комнаты или клиента при сохранении оценки")
				continue
			}

			// Создаем основную запись в ExamGrade
			grade := models.ExamGrade{
				ExamID:         data.ExamID,
				ExaminerID:     client.UserID,
				StudentID:      data.StudentID,
				Qualification:  data.Qualification,
				Specialization: data.Specialization,
				Recommendation: data.Recommendations,
				Abstained:      data.Abstained,
			}

			if err := database.DB.Create(&grade).Error; err != nil {
				log.Println("Ошибка сохранения основной записи оценки:", err)
				continue
			}
			// === После сохранения ExamGrade ===

			// Считаем количество оценок для этого студента
			var gradesCount int64
			database.DB.Model(&models.ExamGrade{}).
				Where("exam_id = ? AND student_id = ?", data.ExamID, data.StudentID).
				Count(&gradesCount)

			// Узнаём общее количество экзаменаторов
			var exam models.Exam
			if err := database.DB.First(&exam, data.ExamID).Error; err != nil {
				log.Println("Ошибка загрузки экзамена:", err)
				continue
			}

			// Считаем экзаменаторов + председателя
			var totalExaminers int64
			database.DB.Table("exam_examiners").Where("exam_id = ?", data.ExamID).Count(&totalExaminers)
			totalExaminers += 1 // председатель тоже оценивает

			// Высчитываем процент
			progressPercent := int(float64(gradesCount) / float64(totalExaminers) * 100)

			// Завершен ли студент?
			completed := gradesCount >= totalExaminers

			// Шлём всем обновление прогресса
			if room != nil {
				room.Mutex.Lock()
				defer room.Mutex.Unlock()

				for _, examiner := range room.Examiners {
					if examiner.Conn != nil {
						examiner.Conn.WriteJSON(map[string]interface{}{
							"type": "progress_update",
							"data": map[string]interface{}{
								"studentId": data.StudentID,
								"progress":  progressPercent,
								"completed": completed,
							},
						})
					}
				}
				if room.Chairman != nil && room.Chairman.Conn != nil {
					room.Chairman.Conn.WriteJSON(map[string]interface{}{
						"type": "progress_update",
						"data": map[string]interface{}{
							"studentId": data.StudentID,
							"progress":  progressPercent,
							"completed": completed,
						},
					})
				}
			}
			// Если НЕ воздержался — создаём критерии
			if !data.Abstained && len(data.Scores) > 0 {
				for i, score := range data.Scores {
					criterion := models.ExamGradeCriterion{
						GradeID:     grade.ID,
						CriterionID: i + 1, // 🔥 Критерии идут с 1
						Score:       optionalInt(score),
					}
					if err := database.DB.Create(&criterion).Error; err != nil {
						log.Println("Ошибка сохранения критерия:", err)
						continue
					}
				}
			}

			// Перенаправляем обратно
			var redirectURL string
			if client.Role == "chairman" {
				redirectURL = fmt.Sprintf("/user/exam/start/%d", data.ExamID)
			} else {
				redirectURL = fmt.Sprintf("/user/exam/waiting/%d", data.ExamID)
			}

			if client.Conn != nil {
				client.Conn.WriteJSON(map[string]interface{}{
					"type": "redirect",
					"data": map[string]interface{}{
						"url": redirectURL,
					},
				})
			}

		case "ping":
			// ничего не делать на пинг
			continue

		default:
			log.Println("Неизвестный тип сообщения:", incoming.Type)
		}
	}

	if room != nil && client != nil {
		room.Mutex.Lock()
		delete(room.Connected, client.UserID)
		room.Mutex.Unlock()
		broadcastExaminerList(room)
		broadcastChairmanStatus(room)
	}
}

func optionalInt(v int) *int {
	return &v
}

func handleInitUser(c *websocket.Conn, examID int, userID uint, userName string) (*ExamRoom, *Client) {
	var exam models.Exam
	if err := database.DB.First(&exam, examID).Error; err != nil {
		log.Println("Ошибка загрузки экзамена:", err)
		return nil, nil
	}

	client := &Client{
		UserID: userID,
		Name:   userName,
		Conn:   c,
	}

	globalMutex.Lock()
	room, exists := examRooms[examID]
	if !exists {
		room = &ExamRoom{
			Examiners:    make(map[uint]*Client),
			Connected:    make(map[uint]bool),
			Progress:     make(map[uint]int),
			Quorum:       2,
			ExaminerList: []ExaminerInfo{},
		}
		examRooms[examID] = room
	}
	globalMutex.Unlock()

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	if userID == exam.ChairmanID {
		client.Role = "chairman"
		room.ChairmanID = userID
		room.Chairman = client

		// Загружаем список экзаменаторов экзамена
		var examinerIDs []uint
		database.DB.Table("exam_examiners").Where("exam_id = ?", examID).Pluck("user_id", &examinerIDs)

		room.ExaminerList = []ExaminerInfo{}
		for _, eid := range examinerIDs {
			if eid == exam.ChairmanID {
				continue
			}
			var user models.User
			if err := database.DB.First(&user, eid).Error; err == nil {
				room.ExaminerList = append(room.ExaminerList, ExaminerInfo{
					ID:     user.ID,
					Name:   user.SurnameInIp + " " + user.NameInIp + " " + user.LastnameInIp,
					Avatar: findAvatarPath(user.StoragePath),
				})
			}
		}
	} else {
		client.Role = "examiner"
		room.Examiners[userID] = client
	}

	return room, client
}

func handleRequestExaminerList(c *websocket.Conn, examID int) {
	globalMutex.Lock()
	room, exists := examRooms[examID]
	globalMutex.Unlock()

	if !exists || room == nil {
		return
	}
	broadcastExaminerList(room)
}

func handleStartExam(c *websocket.Conn, examID int) {
	globalMutex.Lock()
	room, exists := examRooms[examID]
	globalMutex.Unlock()

	if !exists || room == nil {
		log.Println("Комната не найдена:", examID)
		return
	}

	// Проверяем, что председатель всё ещё в комнате
	if room.Chairman == nil || room.Chairman.Conn == nil {
		log.Println("Председатель не найден или не в сети:", examID)
		return
	}

	// Отправляем статус председателя всем участникам (в том числе экзаменаторам)
	for _, examiner := range room.Examiners {
		if examiner.Conn != nil {
			examiner.Conn.WriteJSON(map[string]interface{}{
				"type": "chairman_status",
				"data": "present", // Убедимся, что председатель в сети
			})
		}
	}

	// Отправляем статус председателя самому председателю (если его соединение всё ещё активно)
	if room.Chairman.Conn != nil {
		room.Chairman.Conn.WriteJSON(map[string]interface{}{
			"type": "chairman_status",
			"data": "present", // Убедимся, что председатель в сети
		})
	}

	// Редирект на страницу управления экзаменом
	startData := map[string]interface{}{
		"type": "redirect",
		"data": map[string]interface{}{
			"url": "/user/exam/start/" + strconv.Itoa(examID), // Переход на страницу управления экзаменом
		},
	}

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	// Отправляем редирект на страницу управления экзаменом председателю
	if room.Chairman != nil && room.Chairman.Conn != nil {
		room.Chairman.Conn.WriteJSON(startData)
	}
}

func handleOpenStudent(c *websocket.Conn, examID int, studentID uint) {
	globalMutex.Lock()
	room, exists := examRooms[examID]
	globalMutex.Unlock()

	if !exists || room == nil {
		return
	}

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	openData := map[string]interface{}{
		"type": "open_student",
		"data": map[string]interface{}{
			"url": fmt.Sprintf("/user/exam/student/%d/%d", examID, studentID), // 🔥 Исправлено
			"id":  studentID,
		},
	}

	for _, examiner := range room.Examiners {
		if examiner.Conn != nil && room.Connected[examiner.UserID] {
			examiner.Conn.WriteJSON(openData)
		}
	}

	if room.Chairman != nil && room.Chairman.Conn != nil {
		room.Chairman.Conn.WriteJSON(openData)
	}
}

func handleProgressUpdate(c *websocket.Conn, examID int, studentID uint, currentProgress int) {
	globalMutex.Lock()
	room, exists := examRooms[examID]
	globalMutex.Unlock()

	if !exists || room == nil {
		return
	}

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	room.Progress[studentID] = currentProgress

	progressData := map[string]interface{}{
		"type": "progress_update",
		"data": map[string]interface{}{
			"studentId":       studentID,
			"currentProgress": currentProgress,
		},
	}

	if room.Chairman != nil && room.Chairman.Conn != nil {
		room.Chairman.Conn.WriteJSON(progressData)
	}
}

func broadcastExaminerList(room *ExamRoom) {
	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	list := []map[string]interface{}{}
	for _, examiner := range room.ExaminerList {
		status := "offline"
		if room.Connected[examiner.ID] {
			status = "online"
		}
		list = append(list, map[string]interface{}{
			"id":     examiner.ID,
			"name":   examiner.Name,
			"avatar": examiner.Avatar,
			"status": status,
		})
	}

	if room.Chairman != nil && room.Chairman.Conn != nil {
		room.Chairman.Conn.WriteJSON(map[string]interface{}{
			"type": "examiner_list",
			"data": list,
		})
	}
}

func broadcastChairmanStatus(room *ExamRoom) {
	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	// Статус председателя для всех экзаменаторов
	for _, examiner := range room.Examiners {
		if examiner.Conn != nil {
			status := "absent"
			// Проверяем, подключён ли председатель
			if room.Chairman != nil && room.Connected[room.ChairmanID] {
				status = "present"
			}
			examiner.Conn.WriteJSON(map[string]interface{}{
				"type": "chairman_status",
				"data": status,
			})
		}
	}
}

func startPingPong(room *ExamRoom) {
	for {
		time.Sleep(30 * time.Second)

		room.Mutex.Lock()
		for _, examiner := range room.Examiners {
			if examiner.Conn != nil && room.Connected[examiner.UserID] {
				examiner.Conn.WriteJSON(map[string]interface{}{
					"type": "ping",
				})
			}
		}
		if room.Chairman != nil && room.Chairman.Conn != nil && room.Connected[room.ChairmanID] {
			room.Chairman.Conn.WriteJSON(map[string]interface{}{
				"type": "ping",
			})
		}
		room.Mutex.Unlock()
	}
}

func findAvatarPath(storagePath string) string {
	if storagePath == "" {
		return ""
	}
	files, err := os.ReadDir(storagePath)
	if err != nil {
		return ""
	}
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "avatar") {
			return "/uploads/" + filepath.Base(storagePath) + "/" + f.Name()
		}
	}
	return ""
}







package services

import (
	"att_service/database"
	"att_service/models"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

type Client struct {
	UserID uint
	Name   string
	Role   string
	Conn   *websocket.Conn
}

type ExaminerInfo struct {
	ID     uint
	Name   string
	Avatar string
}

type ExamRoom struct {
	ChairmanID   uint
	Chairman     *Client
	Examiners    map[uint]*Client
	ExaminerList []ExaminerInfo
	Connected    map[uint]bool
	Progress     map[uint]int // studentID -> number of votes
	Quorum       int
	ChairmanStatus string      // Статус председателя (present / absent)
	Mutex        sync.Mutex
}

var examRooms = make(map[int]*ExamRoom)
var globalMutex sync.Mutex

// WebSocket обработчик
func WebSocketHandler(c *websocket.Conn) {
	defer c.Close()

	var room *ExamRoom
	var client *Client

	// Основной цикл чтения сообщений
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}

		var incoming struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}

		// Парсим сообщение
		if err := json.Unmarshal(msg, &incoming); err != nil {
			log.Println("Ошибка парсинга сообщения:", err)
			continue
		}

		switch incoming.Type {
		// Инициализация пользователя (председатель или экзаменатор)
		case "init_user":
			var data struct {
				ExamID int    `json:"exam_id"`
				UserID uint   `json:"user_id"`
				Name   string `json:"name"`
				Role   string `json:"role"`
			}
			if err := json.Unmarshal(incoming.Data, &data); err != nil {
				log.Println("Ошибка парсинга init_user:", err)
				continue
			}

			// Инициализация пользователя и комнаты
			r, cl := handleInitUser(c, data.ExamID, data.UserID, data.Name)
			r.Mutex.Lock()
			r.Connected[cl.UserID] = true
			r.Mutex.Unlock()
			room = r
			client = cl

			// Отправляем сообщение о статусе председателя всем участникам
			go startPingPong(room)
			broadcastExaminerList(room)
			broadcastChairmanStatus(room)

		// Запрос списка экзаменаторов
		case "request_examiner_list":
			var data struct {
				ExamID int `json:"exam_id"`
			}
			if err := json.Unmarshal(incoming.Data, &data); err != nil {
				log.Println("Ошибка парсинга request_examiner_list:", err)
				continue
			}
			handleRequestExaminerList(c, data.ExamID)

		// Старт экзамена
		case "start_exam":
			var data struct {
				ExamID int `json:"exam_id"`
			}
			if err := json.Unmarshal(incoming.Data, &data); err != nil {
				log.Println("Ошибка парсинга start_exam:", err)
				continue
			}
			handleStartExam(c, data.ExamID)

		// Открытие студента для экзаменаторов
		case "open_student":
			var data struct {
				ExamID    int  `json:"exam_id"`
				StudentID uint `json:"student_id"`
			}
			if err := json.Unmarshal(incoming.Data, &data); err != nil {
				log.Println("Ошибка парсинга open_student:", err)
				continue
			}
			handleOpenStudent(c, data.ExamID, data.StudentID)

		// Обновление прогресса студента
		case "progress_update":
			var data struct {
				ExamID          int  `json:"exam_id"`
				StudentID       uint `json:"student_id"`
				CurrentProgress int  `json:"current_progress"`
			}
			if err := json.Unmarshal(incoming.Data, &data); err != nil {
				log.Println("Ошибка парсинга progress_update:", err)
				continue
			}
			handleProgressUpdate(c, data.ExamID, data.StudentID, data.CurrentProgress)

		// Сохранение оценки
		case "save_grade":
			var data struct {
				ExamID          uint   `json:"exam_id"`
				StudentID       uint   `json:"student_id"`
				Scores          []int  `json:"scores"`          // баллы (массив)
				Qualification   string `json:"qualification"`   // квалификация
				Recommendations string `json:"recommendations"` // рекомендации
				Specialization  string `json:"specialization"`  // специализация
				Abstained       bool   `json:"abstained"`       // флаг воздержания
			}
			if err := json.Unmarshal(incoming.Data, &data); err != nil {
				log.Println("Ошибка разбора save_grade:", err)
				continue
			}

			if room == nil || client == nil {
				log.Println("Нет комнаты или клиента при сохранении оценки")
				continue
			}

			// Создаем основную запись в ExamGrade
			grade := models.ExamGrade{
				ExamID:         data.ExamID,
				ExaminerID:     client.UserID,
				StudentID:      data.StudentID,
				Qualification:  data.Qualification,
				Specialization: data.Specialization,
				Recommendation: data.Recommendations,
				Abstained:      data.Abstained,
			}

			if err := database.DB.Create(&grade).Error; err != nil {
				log.Println("Ошибка сохранения основной записи оценки:", err)
				continue
			}

			// Считаем количество оценок для этого студента
			var gradesCount int64
			database.DB.Model(&models.ExamGrade{}).
				Where("exam_id = ? AND student_id = ?", data.ExamID, data.StudentID).
				Count(&gradesCount)

			// Узнаём общее количество экзаменаторов
			var exam models.Exam
			if err := database.DB.First(&exam, data.ExamID).Error; err != nil {
				log.Println("Ошибка загрузки экзамена:", err)
				continue
			}

			// Считаем экзаменаторов + председателя
			var totalExaminers int64
			database.DB.Table("exam_examiners").Where("exam_id = ?", data.ExamID).Count(&totalExaminers)
			totalExaminers += 1 // председатель тоже оценивает

			// Высчитываем процент
			progressPercent := int(float64(gradesCount) / float64(totalExaminers) * 100)

			// Завершен ли студент?
			completed := gradesCount >= totalExaminers

			// Шлём всем обновление прогресса
			if room != nil {
				room.Mutex.Lock()
				defer room.Mutex.Unlock()

				for _, examiner := range room.Examiners {
					if examiner.Conn != nil {
						examiner.Conn.WriteJSON(map[string]interface{}{
							"type": "progress_update",
							"data": map[string]interface{}{
								"studentId": data.StudentID,
								"progress":  progressPercent,
								"completed": completed,
							},
						})
					}
				}
				if room.Chairman != nil && room.Chairman.Conn != nil {
					room.Chairman.Conn.WriteJSON(map[string]interface{}{
						"type": "progress_update",
						"data": map[string]interface{}{
							"studentId": data.StudentID,
							"progress":  progressPercent,
							"completed": completed,
						},
					})
				}
			}
			// Если НЕ воздержался — создаём критерии
			if !data.Abstained && len(data.Scores) > 0 {
				for i, score := range data.Scores {
					criterion := models.ExamGradeCriterion{
						GradeID:     grade.ID,
						CriterionID: i + 1, // 🔥 Критерии идут с 1
						Score:       optionalInt(score),
					}
					if err := database.DB.Create(&criterion).Error; err != nil {
						log.Println("Ошибка сохранения критерия:", err)
						continue
					}
				}
			}

			// Перенаправляем обратно
			var redirectURL string
			if client.Role == "chairman" {
				redirectURL = fmt.Sprintf("/user/exam/start/%d", data.ExamID)
			} else {
				redirectURL = fmt.Sprintf("/user/exam/waiting/%d", data.ExamID)
			}

			if client.Conn != nil {
				client.Conn.WriteJSON(map[string]interface{}{
					"type": "redirect",
					"data": map[string]interface{}{
						"url": redirectURL,
					},
				})
			}

		case "ping":
			// ничего не делать на пинг
			continue

		default:
			log.Println("Неизвестный тип сообщения:", incoming.Type)
		}
	}

	if room != nil && client != nil {
		room.Mutex.Lock()
		delete(room.Connected, client.UserID)
		room.Mutex.Unlock()
		broadcastExaminerList(room)
		broadcastChairmanStatus(room)
	}
}

func optionalInt(v int) *int {
	return &v
}

func handleInitUser(c *websocket.Conn, examID int, userID uint, userName string) (*ExamRoom, *Client) {
	var exam models.Exam
	if err := database.DB.First(&exam, examID).Error; err != nil {
		log.Println("Ошибка загрузки экзамена:", err)
		return nil, nil
	}

	client := &Client{
		UserID: userID,
		Name:   userName,
		Conn:   c,
	}

	globalMutex.Lock()
	room, exists := examRooms[examID]
	if !exists {
		room = &ExamRoom{
			Examiners:    make(map[uint]*Client),
			Connected:    make(map[uint]bool),
			Progress:     make(map[uint]int),
			Quorum:       2,
			ExaminerList: []ExaminerInfo{},
		}
		examRooms[examID] = room
	}
	globalMutex.Unlock()

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	if userID == exam.ChairmanID {
		client.Role = "chairman"
		room.ChairmanID = userID
		room.Chairman = client

		// Загружаем список экзаменаторов экзамена
		var examinerIDs []uint
		database.DB.Table("exam_examiners").Where("exam_id = ?", examID).Pluck("user_id", &examinerIDs)

		room.ExaminerList = []ExaminerInfo{}
		for _, eid := range examinerIDs {
			if eid == exam.ChairmanID {
				continue
			}
			var user models.User
			if err := database.DB.First(&user, eid).Error; err == nil {
				room.ExaminerList = append(room.ExaminerList, ExaminerInfo{
					ID:     user.ID,
					Name:   user.SurnameInIp + " " + user.NameInIp + " " + user.LastnameInIp,
					Avatar: findAvatarPath(user.StoragePath),
				})
			}
		}
	} else {
		client.Role = "examiner"
		room.Examiners[userID] = client
	}

	return room, client
}

func handleRequestExaminerList(c *websocket.Conn, examID int) {
	globalMutex.Lock()
	room, exists := examRooms[examID]
	globalMutex.Unlock()

	if !exists || room == nil {
		return
	}
	broadcastExaminerList(room)
}

func handleStartExam(c *websocket.Conn, examID int) {
	globalMutex.Lock()
	room, exists := examRooms[examID]
	globalMutex.Unlock()

	if !exists || room == nil {
		log.Println("Комната не найдена:", examID)
		return
	}

	// Проверяем, что председатель всё ещё в комнате
	if room.Chairman == nil || room.Chairman.Conn == nil {
		log.Println("Председатель не найден или не в сети:", examID)
		return
	}

	// Отправляем статус председателя всем участникам (в том числе экзаменаторам)
	for _, examiner := range room.Examiners {
		if examiner.Conn != nil {
			examiner.Conn.WriteJSON(map[string]interface{}{
				"type": "chairman_status",
				"data": "present", // Убедимся, что председатель в сети
			})
		}
	}

	// Отправляем статус председателя самому председателю (если его соединение всё ещё активно)
	if room.Chairman.Conn != nil {
		room.Chairman.Conn.WriteJSON(map[string]interface{}{
			"type": "chairman_status",
			"data": "present", // Убедимся, что председатель в сети
		})
	}

	// Редирект на страницу управления экзаменом
	startData := map[string]interface{}{
		"type": "redirect",
		"data": map[string]interface{}{
			"url": "/user/exam/start/" + strconv.Itoa(examID), // Переход на страницу управления экзаменом
		},
	}

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	// Отправляем редирект на страницу управления экзаменом председателю
	if room.Chairman != nil && room.Chairman.Conn != nil {
		room.Chairman.Conn.WriteJSON(startData)
	}
}

func handleOpenStudent(c *websocket.Conn, examID int, studentID uint) {
	globalMutex.Lock()
	room, exists := examRooms[examID]
	globalMutex.Unlock()

	if !exists || room == nil {
		return
	}

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	openData := map[string]interface{}{
		"type": "open_student",
		"data": map[string]interface{}{
			"url": fmt.Sprintf("/user/exam/student/%d/%d", examID, studentID), // 🔥 Исправлено
			"id":  studentID,
		},
	}

	for _, examiner := range room.Examiners {
		if examiner.Conn != nil && room.Connected[examiner.UserID] {
			examiner.Conn.WriteJSON(openData)
		}
	}

	if room.Chairman != nil && room.Chairman.Conn != nil {
		room.Chairman.Conn.WriteJSON(openData)
	}
}

func handleProgressUpdate(c *websocket.Conn, examID int, studentID uint, currentProgress int) {
	globalMutex.Lock()
	room, exists := examRooms[examID]
	globalMutex.Unlock()

	if !exists || room == nil {
		return
	}

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	room.Progress[studentID] = currentProgress

	progressData := map[string]interface{}{
		"type": "progress_update",
		"data": map[string]interface{}{
			"studentId":       studentID,
			"currentProgress": currentProgress,
		},
	}

	if room.Chairman != nil && room.Chairman.Conn != nil {
		room.Chairman.Conn.WriteJSON(progressData)
	}
}

func broadcastExaminerList(room *ExamRoom) {
	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	list := []map[string]interface{}{}
// Формируем список экзаменаторов с их статусами
	for _, examiner := range room.ExaminerList {
		status := "offline"
		if room.Connected[examiner.ID] {
			status = "online"
		}
		list = append(list, map[string]interface{}{
			"id":     examiner.ID,
			"name":   examiner.Name,
			"avatar": examiner.Avatar,
			"status": status,
		})
	}

	if room.Chairman != nil && room.Chairman.Conn != nil {
		room.Chairman.Conn.WriteJSON(map[string]interface{}{
			"type": "examiner_list",
			"data": list,
		})
	}
}

func broadcastChairmanStatus(room *ExamRoom) {
	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	// Статус председателя для всех экзаменаторов
	for _, examiner := range room.Examiners {
		if examiner.Conn != nil {
			status := "absent"
			// Проверяем, подключён ли председатель
			if room.Chairman != nil && room.Connected[room.ChairmanID] {
				status = "present"
			}
			examiner.Conn.WriteJSON(map[string]interface{}{
				"type": "chairman_status",
				"data": status,
			})
		}
	}
}

// Функция отправки пинга для поддержания соединения
func startPingPong(room *ExamRoom) {
	for {
		time.Sleep(30 * time.Second)

		room.Mutex.Lock()
		for _, examiner := range room.Examiners {
			if examiner.Conn != nil && room.Connected[examiner.UserID] {
				examiner.Conn.WriteJSON(map[string]interface{}{
					"type": "ping",
				})
			}
		}
		if room.Chairman != nil && room.Chairman.Conn != nil && room.Connected[room.ChairmanID] {
			room.Chairman.Conn.WriteJSON(map[string]interface{}{"type": "ping"})
		}
		room.Mutex.Unlock()
	}
}

func findAvatarPath(storagePath string) string {
	if storagePath == "" {
		return ""
	}
	files, err := os.ReadDir(storagePath)
	if err != nil {
		return ""
	}
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "avatar") {
			return "/uploads/" + filepath.Base(storagePath) + "/" + f.Name()
		}
	}
	return ""
}