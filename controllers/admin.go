package controllers

import (
	"archive/zip"
	"att_service/database"
	"att_service/models"
	"att_service/services"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
)

var SessionStore *session.Store

func AdminSetTargetStudent(c *fiber.Ctx) error {
	var body struct {
		ID     uint   `json:"id"`
		Source string `json:"source"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid request",
		})
	}

	sess, _ := SessionStore.Get(c)
	sess.Set("targetStudentID", body.ID)
	sess.Set("source", body.Source)
	_ = sess.Save()

	return c.JSON(fiber.Map{"success": true})
}

// AdminPage возвращает главную страницу администратора (pages/admin_page.html)
func AdminPage(c *fiber.Ctx) error {
	return c.SendFile("views/pages/admin/main.html")
}

// AdminChangeUserRole меняет роль пользователя по ID
func AdminChangeUserRole(c *fiber.Ctx) error {
	// Структура запроса
	type request struct {
		Role string `json:"role"`
	}

	var body request
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Неверный формат данных"})
	}

	// Достаём сессию
	sess, _ := SessionStore.Get(c)

	studentIDRaw := sess.Get("targetStudentID")
	studentID, ok := studentIDRaw.(uint)
	if !ok || studentID == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("ID студента не найден в сессии")
	}

	// Проверяем корректность роли
	if body.Role != "student" && body.Role != "examiner" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Некорректная роль"})
	}

	// Обновляем роль пользователя в БД
	if err := database.DB.Model(&models.User{}).Where("id = ?", studentID).Update("role", body.Role).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Ошибка обновления роли"})
	}

	return c.JSON(fiber.Map{"success": true})
}

func AdminDeclineStudent(c *fiber.Ctx) error {
	// Достаём сессию
	sess, _ := SessionStore.Get(c)

	studentIDRaw := sess.Get("targetStudentID")
	studentID, ok := studentIDRaw.(uint)
	if !ok || studentID == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("ID студента не найден в сессии")
	}

	var student models.User
	if err := database.DB.First(&student, studentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "student not found"})
	}

	// Получение тела запроса
	var body struct {
		Reasons     []string `json:"reasons"`
		Explanation string   `json:"explanation"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "cannot parse body"})
	}

	var reasonLabels = map[string]string{
		"invalid_name":     "Неверно указано ФИО",
		"invalid_contacts": "Неверно указаны контакты",
		"no_documents":     "Не все документы прикреплены",
	}

	reasons := []string{}
	for _, key := range body.Reasons {
		label, ok := reasonLabels[key]
		if ok {
			reasons = append(reasons, label)
		} else {
			reasons = append(reasons, key) // fallback
		}
	}

	student.Status = "declined"
	student.DeclineReason = strings.Join(reasons, ", ")
	if body.Explanation != "" {
		student.DeclineReason += " | " + body.Explanation
	}

	if err := database.DB.Save(&student).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "update error"})
	}

	return c.JSON(fiber.Map{"success": true})
}

func AdminCreateExamPage(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).SendString("Необходима авторизация")
	}

	var admin models.User
	if err := database.DB.First(&admin, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка загрузки администратора")
	}
	if admin.Role != "admin" {
		return c.Status(fiber.StatusForbidden).SendString("Доступ запрещён")
	}

	// Загружаем экзаменаторов и студентов со статусом "approved"
	var rawExaminers []models.User
	var rawStudents []models.User
	database.DB.Where("role = ? AND status = ?", "examiner", "approved").Find(&rawExaminers)
	database.DB.Where("role = ? AND status = ?", "student", "approved").Find(&rawStudents)

	// Оставляем только студентов с последним approved заявлением
	var approvedStudents []models.User
	for _, u := range rawStudents {
		var lastApp models.Application
		if err := database.DB.Where("user_id = ?", u.ID).Order("created_at DESC").First(&lastApp).Error; err == nil && lastApp.Status == "approved" {
			approvedStudents = append(approvedStudents, u)
		}
	}

	// Подсчёт количества экзаменов со статусом scheduled и completed
	var examCount int64
	database.DB.Model(&models.Exam{}).Where("status IN ?", []string{"scheduled", "completed"}).Count(&examCount)

	// Генерация предварительного кода экзамена
	examCode := generateExamCode(examCount + 1)

	// Вспомогательная функция для поиска аватарки
	findAvatar := func(user models.User) string {
		if user.StoragePath != "" {
			if files, err := os.ReadDir(user.StoragePath); err == nil {
				for _, f := range files {
					if strings.HasPrefix(f.Name(), "avatar") {
						return "/uploads/" + filepath.Base(user.StoragePath) + "/" + f.Name()
					}
				}
			}
		}
		return "/pictures/Generic avatar.png"
	}

	type ExamUser struct {
		ID       uint
		Name     string
		Avatar   string
		Selected bool
	}

	var examiners, students []ExamUser
	for _, u := range rawExaminers {
		examiners = append(examiners, ExamUser{
			ID:     u.ID,
			Name:   fmt.Sprintf("%s %s %s", u.SurnameInIp, u.NameInIp, u.LastnameInIp),
			Avatar: findAvatar(u),
		})
	}
	for _, u := range approvedStudents {
		students = append(students, ExamUser{
			ID:     u.ID,
			Name:   fmt.Sprintf("%s %s %s", u.SurnameInIp, u.NameInIp, u.LastnameInIp),
			Avatar: findAvatar(u),
		})
	}

	return services.Render(c, "admin", "exams/create_exam.html", fiber.Map{
		"role":      admin.Role,
		"status":    admin.Status,
		"avatar":    findAvatar(admin),
		"Examiners": examiners,
		"Students":  students,
		"ExamCode":  examCode,
		"path":      c.Path(),
	})
}

func generateExamCode(index int64) string {
	return fmt.Sprintf("06-30-%d", index)
}
func AdminCreateExam(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	// Проверка прав
	var currentUser models.User
	if err := database.DB.First(&currentUser, userID).Error; err != nil || currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).SendString("Доступ запрещён")
	}

	// Получение формы
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Ошибка получения данных формы")
	}

	// Извлечение параметров
	examinersRaw := form.Value["examiners"]
	studentsRaw := form.Value["students"]
	dateStr := form.Value["date"][0]
	commissionStartStr := form.Value["commission_start"][0]
	commissionEndStr := form.Value["commission_end"][0]

	var examinersID []uint
	var studentsID []uint
	_ = json.Unmarshal([]byte(examinersRaw[0]), &examinersID)
	_ = json.Unmarshal([]byte(studentsRaw[0]), &studentsID)

	date, _ := time.Parse("2006-01-02", dateStr)
	commissionStart, _ := time.Parse("2006-01-02", commissionStartStr)
	commissionEnd, _ := time.Parse("2006-01-02", commissionEndStr)

	status := "planned"
	if val, ok := form.Value["auto_schedule"]; ok && val[0] == "true" {
		status = "scheduled"
	}

	// Загружаем экзамен из сессии, если есть
	sess, _ := SessionStore.Get(c)
	examIDRaw := sess.Get("targetExamID")
	var exam models.Exam
	if examID, ok := examIDRaw.(uint); ok && examID > 0 {
		if err := database.DB.First(&exam, examID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).SendString("Экзамен не найден")
		}
		exam.Date = date
		exam.CommissionStart = commissionStart
		exam.CommissionEnd = commissionEnd
		exam.Status = status
	} else {
		exam = models.Exam{
			Date:            date,
			CommissionStart: commissionStart,
			CommissionEnd:   commissionEnd,
			Status:          status,
		}
	}

	// Добавляем экзаменаторов
	if len(examinersID) > 0 {
		var examiners []models.User
		database.DB.Where("id IN ?", examinersID).Find(&examiners)
		exam.Examiners = examiners
	}

	// Добавляем студентов
	var filteredStudents []models.User
	for _, id := range studentsID {
		var user models.User
		if err := database.DB.First(&user, id).Error; err != nil || user.Role != "student" {
			continue
		}
		var lastApp models.Application
		if err := database.DB.Where("user_id = ?", user.ID).Order("created_at DESC").First(&lastApp).Error; err == nil && lastApp.Status == "approved" {
			filteredStudents = append(filteredStudents, user)
		}
	}
	exam.Students = filteredStudents

	// Сохраняем: если новый — Create, если редактирование — Save (обновление)
	if exam.ID > 0 {
		if err := database.DB.Session(&gorm.Session{FullSaveAssociations: true}).Save(&exam).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Ошибка обновления экзамена")
		}
	} else {
		if err := database.DB.Create(&exam).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Ошибка создания экзамена")
		}
	}

	// Очищаем экзамен из сессии
	sess.Delete("targetExamID")
	sess.Save()

	return c.JSON(fiber.Map{
		"success": true,
	})
}

func AdminCancelExam(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	// Проверка, что пользователь админ
	var admin models.User
	if err := database.DB.First(&admin, userID).Error; err != nil || admin.Role != "admin" {
		return c.Status(fiber.StatusForbidden).SendString("Доступ запрещён")
	}

	// Получаем JSON тело
	var body struct {
		ExamID uint `json:"exam_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.ExamID == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("Неверный ID экзамена")
	}

	// Находим экзамен
	var exam models.Exam
	if err := database.DB.First(&exam, body.ExamID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Экзамен не найден")
	}

	// Меняем статус
	exam.Status = "planned"
	if err := database.DB.Save(&exam).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка при сохранении")
	}

	return c.JSON(fiber.Map{"success": true})
}

func AdminShowExam(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var admin models.User
	if err := database.DB.First(&admin, userID).Error; err != nil || admin.Role != "admin" {
		return c.Status(fiber.StatusForbidden).SendString("Доступ запрещён")
	}

	sess, _ := SessionStore.Get(c)
	examIDRaw := sess.Get("targetExamID")
	examID, ok := examIDRaw.(uint)
	if !ok || examID == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("ID экзамена не найден в сессии")
	}

	// Загружаем экзамен с выбранными пользователями
	var exam models.Exam
	if err := database.DB.Preload("Examiners").Preload("Students").First(&exam, examID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Экзамен не найден")
	}

	// Получаем всех пользователей с ролями
	var allUsers []models.User
	if err := database.DB.Where("role IN ?", []string{"student", "examiner"}).Find(&allUsers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка загрузки пользователей")
	}

	// Карты для проверки, кто был выбран
	selectedStudentIDs := make(map[uint]bool)
	selectedExaminerIDs := make(map[uint]bool)
	for _, u := range exam.Students {
		selectedStudentIDs[u.ID] = true
	}
	for _, u := range exam.Examiners {
		selectedExaminerIDs[u.ID] = true
	}

	// Структура для шаблона
	type ExamUser struct {
		ID       uint
		Name     string
		Avatar   string
		Selected bool
	}

	var examiners []ExamUser
	var students []ExamUser

	for _, u := range allUsers {
		avatar := findAvatar(u.StoragePath)
		fullName := fmt.Sprintf("%s %s %s", u.SurnameInIp, u.NameInIp, u.LastnameInIp)

		if u.Role == "examiner" {
			examiners = append(examiners, ExamUser{
				ID:       u.ID,
				Name:     fullName,
				Avatar:   avatar,
				Selected: selectedExaminerIDs[u.ID],
			})
		} else if u.Role == "student" {
			// Проверяем наличие одобренного заявления
			var lastApp models.Application
			if err := database.DB.Where("user_id = ?", u.ID).
				Order("created_at DESC").
				First(&lastApp).Error; err != nil || lastApp.Status != "approved" {
				continue
			}
			students = append(students, ExamUser{
				ID:       u.ID,
				Name:     fullName,
				Avatar:   avatar,
				Selected: selectedStudentIDs[u.ID],
			})
		}
	}

	return services.Render(c, "admin", "exams/create_exam.html", fiber.Map{
		"role":             admin.Role,
		"status":           admin.Status,
		"avatar":           findAvatar(admin.StoragePath),
		"ExamCode":         fmt.Sprintf("06-30-%d", exam.ID),
		"Examiners":        examiners,
		"Students":         students,
		"exam_date":        exam.Date.Format("2006-01-02"),
		"commission_start": exam.CommissionStart.Format("2006-01-02"),
		"commission_end":   exam.CommissionEnd.Format("2006-01-02"),
		"path":             c.Path(),
	})
}

func DeclineApplication(c *fiber.Ctx) error {
	type DeclineRequest struct {
		ID          uint     `json:"id"`
		Reasons     []string `json:"reasons"`
		Explanation string   `json:"explanation"`
	}

	var req DeclineRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Некорректный формат запроса")
	}

	if req.ID == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("Отсутствует ID заявления")
	}

	// Ищем заявку
	var app models.Application
	if err := database.DB.First(&app, req.ID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Заявление не найдено")
	}

	// Обновляем статус и причину
	app.Status = "declined"
	app.DeclineReason = strings.Join(req.Reasons, ", ")
	app.DeclineExplanation = req.Explanation

	if err := database.DB.Save(&app).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Не удалось обновить заявление")
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

func AdminSetTargetExam(c *fiber.Ctx) error {
	var body struct {
		ID uint `json:"id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	sess, _ := SessionStore.Get(c)
	sess.Set("targetExamID", body.ID)
	_ = sess.Save()

	return c.JSON(fiber.Map{"success": true})
}

func AdminShowStudentProfile(c *fiber.Ctx) error {
	// Достаём сессию
	sess, _ := SessionStore.Get(c)

	studentIDRaw := sess.Get("targetStudentID")
	studentID, ok := studentIDRaw.(uint)
	if !ok || studentID == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("ID студента не найден в сессии")
	}

	source := sess.Get("source")

	// Загружаем пользователя
	var student models.User
	if err := database.DB.First(&student, studentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Пользователь не найден")
	}

	// Ищем аватарку
	avatar := ""
	if student.StoragePath != "" {
		if files, err := os.ReadDir(student.StoragePath); err == nil {
			for _, f := range files {
				if strings.HasPrefix(f.Name(), "avatar") {
					avatar = "/uploads/" + filepath.Base(student.StoragePath) + "/" + f.Name()
					break
				}
			}
		}
	}
	if avatar == "" {
		avatar = "../pictures/Generic avatar.png"
	}

	showButtons := false

	if source == "application" {
		showButtons = true
	}

	return services.Render(c, "admin", "users/show-main.html", fiber.Map{
		"id":             student.ID,
		"surname_in_ip":  student.SurnameInIp,
		"surname_in_rp":  student.SurnameInRp,
		"surname_in_dp":  student.SurnameInDp,
		"name_in_ip":     student.NameInIp,
		"name_in_rp":     student.NameInRp,
		"name_in_dp":     student.NameInDp,
		"lastname_in_ip": student.LastnameInIp,
		"lastname_in_rp": student.LastnameInRp,
		"lastname_in_dp": student.LastnameInDp,
		"sex":            student.Sex,
		"email":          student.Email,
		"mail":           student.Mail,
		"mobile_phone":   student.MobilePhone,
		"work_phone":     student.WorkPhone,
		"status":         student.Status,
		"avatar":         avatar,
		"showButtons":    showButtons,
		"decline_reason": student.DeclineReason,
		"role":           student.Role,
	})
}

func GetPastExamsPage(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var currentUser models.User
	if err := database.DB.First(&currentUser, userID).Error; err != nil || currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).SendString("Доступ запрещён")
	}

	// Получаем экзамены со статусом "completed"
	var exams []models.Exam
	if err := database.DB.
		Where("status = ?", "completed").
		Order("date DESC").
		Find(&exams).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка получения экзаменов")
	}

	type ExamItem struct {
		Date string
	}

	var list []ExamItem
	for _, e := range exams {
		list = append(list, ExamItem{
			Date: e.Date.Format("02.01.2006"),
		})
	}

	return services.Render(c, "admin", "exams/exam-list.html", fiber.Map{
		"Exams": list,
	})
}

func ExamPlanningPage(c *fiber.Ctx) error {
	var exams []models.Exam
	if err := database.DB.
		Where("status = ?", "planned").
		Order("date ASC").
		Find(&exams).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка при загрузке экзаменов")
	}

	type ExamItem struct {
		ID   uint
		Date string
	}

	var plannedExams []ExamItem
	for _, exam := range exams {
		plannedExams = append(plannedExams, ExamItem{
			ID:   exam.ID,
			Date: exam.Date.Format("02.01.2006"),
		})
	}

	return services.Render(c, "admin", "exams/exam-planning.html", fiber.Map{
		"PlannedExams": plannedExams,
	})
}

func ExamScheduledPage(c *fiber.Ctx) error {
	var exams []models.Exam
	if err := database.DB.
		Where("status = ?", "scheduled").
		Order("date ASC").
		Find(&exams).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка при загрузке назначенных экзаменов")
	}

	type ExamItem struct {
		ID   uint
		Date string
	}

	var scheduledExams []ExamItem
	for _, exam := range exams {
		scheduledExams = append(scheduledExams, ExamItem{
			ID:   exam.ID,
			Date: exam.Date.Format("02.01.2006"),
		})
	}

	return services.Render(c, "admin", "exams/exam-scheduled.html", fiber.Map{
		"ScheduledExams": scheduledExams,
	})
}

func ScheduleExam(c *fiber.Ctx) error {
	var input struct {
		ID uint `json:"id"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Некорректный формат данных")
	}

	var exam models.Exam
	if err := database.DB.First(&exam, input.ID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Экзамен не найден")
	}

	// Обновляем статус
	exam.Status = "scheduled"
	if err := database.DB.Save(&exam).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка при обновлении")
	}

	return c.JSON(fiber.Map{"success": true})
}

func AdminShowStudentDocuments(c *fiber.Ctx) error {
	// Достаём сессию
	sess, _ := SessionStore.Get(c)

	studentIDRaw := sess.Get("targetStudentID")
	studentID, ok := studentIDRaw.(uint)
	if !ok || studentID == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("ID студента не найден в сессии")
	}

	source := sess.Get("source")

	// Получаем пользователя, паспорт, диплом
	var student models.User
	if err := database.DB.First(&student, studentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Пользователь не найден")
	}
	var passport models.Passport
	_ = database.DB.Where("user_id = ?", student.ID).First(&passport)

	var edu models.EducationDocument
	_ = database.DB.Where("user_id = ?", student.ID).First(&edu)

	// Скан-изображения
	passportImages := []string{}
	snilsImages := []string{}
	diplomImages := []string{}

	if student.StoragePath != "" {
		entries, _ := os.ReadDir(student.StoragePath)
		for _, f := range entries {
			fp := "/uploads/" + filepath.Base(student.StoragePath) + "/" + f.Name()
			switch {
			case strings.Contains(f.Name(), "паспорт"):
				passportImages = append(passportImages, fp)
			case strings.Contains(f.Name(), "снилс"):
				snilsImages = append(snilsImages, fp)
			case strings.Contains(f.Name(), "диплом"):
				diplomImages = append(diplomImages, fp)
			}
		}
	}
	showButtons := false

	if source == "application" {
		showButtons = true
	}

	return services.Render(c, "admin", "users/show-documents.html", fiber.Map{
		"id":              student.ID,
		"passport_serial": passport.PassportSeries,
		"unit_code":       passport.PassportDivisionCode,
		"passport_num":    passport.PassportNumber,
		"passport_date":   passport.PassportIssueDate.Format("2006-01-02"),
		"passport_issue":  passport.PassportIssuedBy,
		"bithday_date":    passport.BirthDate.Format("2006-01-02"),
		"born_place":      passport.BirthPlace,
		"registr_address": passport.RegistrationAddress,
		"snils_num":       student.Snils,
		"diplom_num":      edu.DiplomaRegNumber,
		"status":          student.Status,
		"avatar":          findAvatar(student.StoragePath),
		"passport_images": passportImages,
		"snils_images":    snilsImages,
		"diplom_images":   diplomImages,
		"showButtons":     showButtons,
		"decline_reason":  student.DeclineReason,
		"role":            student.Role,
	})
}

func AdminConfirmStudent(c *fiber.Ctx) error {
	// Достаём сессию
	sess, _ := SessionStore.Get(c)

	studentIDRaw := sess.Get("targetStudentID")
	studentID, ok := studentIDRaw.(uint)
	if !ok || studentID == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("ID студента не найден в сессии")
	}

	// Проверка прав администратора
	adminID := c.Locals("userID").(uint)
	var admin models.User
	if err := database.DB.First(&admin, adminID).Error; err != nil || admin.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "Нет доступа",
		})
	}

	// Читаем JSON тело
	var req struct {
		Confirm bool `json:"confirm"`
	}
	if err := c.BodyParser(&req); err != nil || !req.Confirm {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Некорректные данные запроса",
		})
	}

	// Обновляем статус студента
	var student models.User
	if err := database.DB.First(&student, studentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Студент не найден",
		})
	}

	student.Status = "approved"
	if err := database.DB.Save(&student).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Ошибка при сохранении",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// Вспомогательная функция для поиска аватарки
func findAvatar(path string) string {
	if path == "" {
		return ""
	}
	files, _ := os.ReadDir(path)
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "avatar") {
			return "/uploads/" + filepath.Base(path) + "/" + f.Name()
		}
	}
	return ""
}

func AdminUserApplication(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	// Проверка прав
	var currentUser models.User
	if err := database.DB.First(&currentUser, userID).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Пользователь не найден")
	}
	if currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).SendString("Доступ запрещён")
	}

	// Получение всех пользователей
	var users []models.User
	if err := database.DB.Where("status = ?", "pending").Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка получения списка пользователей")
	}
	// Сборка списка
	type UserItem struct {
		ID       uint
		Surname  string
		Name     string
		Lastname string
		Role     string
		Avatar   string
	}

	var list []UserItem
	for _, u := range users {
		avatar := ""
		if u.StoragePath != "" {
			if files, err := os.ReadDir(u.StoragePath); err == nil {
				for _, f := range files {
					if strings.HasPrefix(f.Name(), "avatar") {
						avatar = "/uploads/" + filepath.Base(u.StoragePath) + "/" + f.Name()
						break
					}
				}
			}
		}
		if avatar == "" {
			avatar = "../../pictures/Generic avatar.png"
		}

		list = append(list, UserItem{
			ID:       u.ID,
			Surname:  u.SurnameInIp,
			Name:     u.NameInIp,
			Lastname: u.LastnameInIp,
			Role:     u.Role,
			Avatar:   avatar,
		})
	}

	// Передаём данные в шаблон
	return services.Render(c, "admin", "users/user-applications.html", fiber.Map{
		"Lists": list,
	})
}

func AdminSelectUsersByRole(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	// Проверка прав доступа
	var currentUser models.User
	if err := database.DB.First(&currentUser, userID).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Пользователь не найден"})
	}
	if currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ запрещён"})
	}

	// Читаем тело запроса
	var req struct {
		Role string `json:"role"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Некорректные данные"})
	}

	// Ролевой маппинг
	roleMap := map[string]string{
		"students":  "student",
		"examiners": "examiner",
		"admin":     "admin",
	}

	var users []models.User
	query := database.DB

	if req.Role != "all" {
		if mappedRole, ok := roleMap[req.Role]; ok {
			query = query.Where("role = ?", mappedRole)
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неизвестная роль"})
		}
	}

	if err := query.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка запроса к БД"})
	}

	// Формируем ответ
	type UserItem struct {
		ID       uint   `json:"id"`
		Surname  string `json:"surname"`
		Name     string `json:"name"`
		Lastname string `json:"lastname"`
		Role     string `json:"role"`
		Avatar   string `json:"avatar"`
	}

	var list []UserItem
	for _, u := range users {
		avatar := "/pictures/Generic avatar.png"
		if u.StoragePath != "" {
			if files, err := os.ReadDir(u.StoragePath); err == nil {
				for _, f := range files {
					if strings.HasPrefix(f.Name(), "avatar") {
						avatar = "/uploads/" + filepath.Base(u.StoragePath) + "/" + f.Name()
						break
					}
				}
			}
		}

		list = append(list, UserItem{
			ID:       u.ID,
			Surname:  u.SurnameInIp,
			Name:     u.NameInIp,
			Lastname: u.LastnameInIp,
			Role:     u.Role,
			Avatar:   avatar,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"users":   list,
	})
}

func AdminSelectApplicationUsers(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	// Проверка прав
	var currentUser models.User
	if err := database.DB.First(&currentUser, userID).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Пользователь не найден"})
	}
	if currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ запрещён"})
	}

	// Читаем тело запроса
	var req struct {
		Role string `json:"role"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Некорректные данные"})
	}

	// Загружаем пользователей с нужной ролью и статусом "pending"
	var users []models.User
	query := database.DB.Where("status = ?", "pending") // только неподтверждённые

	roleMap := map[string]string{
		"students":  "student",
		"examiners": "examiner",
		"admin":     "admin",
	}

	role := req.Role
	if role != "all" {
		if mapped, ok := roleMap[role]; ok {
			role = mapped
			query = query.Where("role = ?", role)
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неизвестная роль"})
		}
	}

	if err := query.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при получении данных"})
	}

	type UserItem struct {
		ID       uint   `json:"id"`
		Surname  string `json:"surname"`
		Name     string `json:"name"`
		Lastname string `json:"lastname"`
		Role     string `json:"role"`
		Avatar   string `json:"avatar"`
	}

	var list []UserItem
	for _, u := range users {
		avatar := "/pictures/Generic avatar.png"
		if u.StoragePath != "" {
			if files, err := os.ReadDir(u.StoragePath); err == nil {
				for _, f := range files {
					if strings.HasPrefix(f.Name(), "avatar") {
						avatar = "/uploads/" + filepath.Base(u.StoragePath) + "/" + f.Name()
						break
					}
				}
			}
		}

		list = append(list, UserItem{
			ID:       u.ID,
			Surname:  u.SurnameInIp,
			Name:     u.NameInIp,
			Lastname: u.LastnameInIp,
			Role:     u.Role,
			Avatar:   avatar,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"users":   list,
	})
}

func AdminSearchAll(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	// Проверка на роль
	var currentUser models.User
	if err := database.DB.First(&currentUser, userID).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "Пользователь не найден"})
	}
	if currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Доступ запрещён"})
	}

	// Получаем фамилию из запроса
	var req struct {
		Surname string `json:"surname"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Невалидные данные"})
	}

	// Поиск пользователей по фамилии
	var users []models.User
	if err := database.DB.Where("surname_in_ip ILIKE ?", "%"+req.Surname+"%").Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Ошибка поиска"})
	}

	// Сборка результатов
	type UserItem struct {
		ID       uint   `json:"id"`
		Surname  string `json:"surname"`
		Name     string `json:"name"`
		Lastname string `json:"lastname"`
		Role     string `json:"role"`
		Avatar   string `json:"avatar"`
	}

	var result []UserItem
	for _, u := range users {
		avatar := ""
		if u.StoragePath != "" {
			files, _ := os.ReadDir(u.StoragePath)
			for _, f := range files {
				if strings.HasPrefix(f.Name(), "avatar") {
					avatar = "/uploads/" + filepath.Base(u.StoragePath) + "/" + f.Name()
					break
				}
			}
		}
		result = append(result, UserItem{
			ID:       u.ID,
			Surname:  u.SurnameInIp,
			Name:     u.NameInIp,
			Lastname: u.LastnameInIp,
			Role:     u.Role,
			Avatar:   avatar,
		})
	}

	if len(result) == 0 {
		return c.JSON(fiber.Map{
			"success": false,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"users":   result,
	})
}

// AdminUserListPage возвращает страницу списка пользователей (pages/admin__user_list.html)
func AdminUserList(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	// Проверка, что текущий пользователь — админ
	var currentUser models.User
	if err := database.DB.First(&currentUser, userID).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Пользователь не найден")
	}
	if currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).SendString("Доступ запрещён")
	}

	// Получаем список всех пользователей
	var users []models.User
	if err := database.DB.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка загрузки пользователей")
	}

	// Собираем список с ФИО, ролью и аватаркой
	type UserItem struct {
		ID       uint   `json:"id"`
		Surname  string `json:"surname"`
		Name     string `json:"name"`
		Lastname string `json:"lastname"`
		Role     string `json:"role"`
		Avatar   string `json:"avatar"`
	}

	var list []UserItem
	for _, u := range users {
		avatar := ""
		if u.StoragePath != "" {
			files, _ := os.ReadDir(u.StoragePath)
			for _, f := range files {
				if strings.HasPrefix(f.Name(), "avatar") {
					avatar = "/uploads/" + filepath.Base(u.StoragePath) + "/" + f.Name()
					break
				}
			}
		}
		list = append(list, UserItem{
			ID:       u.ID,
			Surname:  u.SurnameInIp,
			Name:     u.NameInIp,
			Lastname: u.LastnameInIp,
			Role:     u.Role,
			Avatar:   avatar,
		})
	}

	return services.Render(c, "admin", "users/user-list.html", fiber.Map{
		"Lists": list,
	})
}

func AdminDeleteStudent(c *fiber.Ctx) error {
	// Достаём сессию
	sess, _ := SessionStore.Get(c)

	studentIDRaw := sess.Get("targetStudentID")
	userID, ok := studentIDRaw.(uint)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("ID студента не найден в сессии")
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Пользователь не найден"})
	}

	// Архивация файлов
	if user.StoragePath != "" {
		if stat, err := os.Stat(user.StoragePath); err == nil && stat.IsDir() {
			archiveDir := "./archive"
			if _, err := os.Stat(archiveDir); os.IsNotExist(err) {
				if err := os.MkdirAll(archiveDir, 0755); err != nil {
					log.Printf("Ошибка создания папки архива: %v", err)
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Ошибка архивации"})
				}
			}
			timestamp := time.Now().Format("20060102_150405")
			safeName := fmt.Sprintf("%d_%s_%s", userID, sanitizeString(user.SurnameInIp), sanitizeString(user.Email))
			zipFile := filepath.Join(archiveDir, fmt.Sprintf("%s_%s.zip", safeName, timestamp))
			if err := zipDirectory(user.StoragePath, zipFile); err != nil {
				log.Printf("Ошибка архивирования данных пользователя %d: %v", userID, err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Ошибка архивирования файлов"})
			}
			_ = os.RemoveAll(user.StoragePath)
		}
	}

	// Удаление связанных данных
	_ = database.DB.Unscoped().Where("user_id = ?", userID).Delete(&models.Passport{}).Error
	_ = database.DB.Unscoped().Where("user_id = ?", userID).Delete(&models.EducationDocument{}).Error
	_ = database.DB.Unscoped().Where("user_id = ?", userID).Delete(&models.UserDocument{}).Error

	// Удаление пользователя полностью, без soft delete
	if err := database.DB.Unscoped().Delete(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Ошибка удаления пользователя"})
	}

	sess.Delete("targetStudentID")
	sess.Delete("source")
	_ = sess.Save()

	return c.JSON(fiber.Map{"success": true, "message": "Пользователь удалён и архив создан"})
}

// zipDirectory архивирует содержимое директории в zip-файл.
func zipDirectory(srcDir, destZip string) error {
	zipFile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = zip.Deflate
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
}

func sanitizeString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "@", "_")
	s = strings.ReplaceAll(s, "/", "_")
	return s
}

func GetAdminExamApplications(c *fiber.Ctx) error {
	adminID := c.Locals("userID").(uint)
	if adminID == 0 {
		return c.Status(fiber.StatusUnauthorized).SendString("Необходима авторизация")
	}

	var applications []models.Application
	if err := database.DB.Where("status = ?", "pending").Find(&applications).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка загрузки заявок")
	}

	type ExamItem struct {
		UserName string
		AppID    uint
		Avatar   string
	}

	var exams []ExamItem
	for _, app := range applications {
		var user models.User
		if err := database.DB.First(&user, app.UserID).Error; err == nil {
			fullName := fmt.Sprintf("%s %s %s", user.SurnameInIp, user.NameInIp, user.LastnameInIp)

			avatar := ""
			if user.StoragePath != "" {
				files, _ := os.ReadDir(user.StoragePath)
				for _, file := range files {
					if strings.HasPrefix(file.Name(), "avatar") {
						avatar = "/uploads/" + filepath.Base(user.StoragePath) + "/" + file.Name()
						break
					}
				}
			}

			exams = append(exams, ExamItem{
				UserName: fullName,
				AppID:    app.ID,
				Avatar:   avatar,
			})
		}
	}

	return services.Render(c, "admin", "exam_applications/exam-applications.html", fiber.Map{
		"Exams": exams,
	})
}

func AdminShowStudentApplication(c *fiber.Ctx) error {
	adminID := c.Locals("userID").(uint)
	if adminID == 0 {
		return c.Status(fiber.StatusUnauthorized).SendString("Необходима авторизация")
	}

	appID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Некорректный ID заявления")
	}

	var application models.Application
	if err := database.DB.First(&application, appID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Заявление не найдено")
	}

	var user models.User
	if err := database.DB.First(&user, application.UserID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Пользователь не найден")
	}

	// Поиск изображений пользователя
	getImages := func(prefix string) []string {
		var images []string
		if user.StoragePath != "" {
			files, _ := os.ReadDir(user.StoragePath)
			for _, f := range files {
				if strings.Contains(f.Name(), prefix) {
					images = append(images, "/uploads/"+filepath.Base(user.StoragePath)+"/"+f.Name())
				}
			}
		}
		return images
	}

	return services.Render(c, "admin", "exam_applications/exam-application.html", fiber.Map{
		"AppID":                       appID,
		"native_language":             application.NativeLanguage,
		"citizenship":                 application.Citizenship,
		"marital_status":              application.MaritalStatus,
		"organization":                application.Organization,
		"job_position":                application.JobPosition,
		"requested_category":          application.RequestedCategory,
		"basis_for_attestation":       application.BasisForAttestation,
		"existing_category":           application.ExistingCategory,
		"existing_category_term":      application.ExistingCategoryTerm,
		"work_experience":             application.WorkExperience,
		"current_position_experience": application.CurrentPositionExperience,
		"awards_info":                 application.AwardsInfo,
		"training_info":               application.TrainingInfo,
		"memberships":                 application.Memberships,
		"consent":                     application.Consent,
		"status":                      user.Status,
		"role":                        user.Role,
		"avatar":                      findAvatar(user.StoragePath),
		"diplom_images":               getImages("диплом"),
		"diplom_jest_images":          getImages("жест"),
		"passport_images":             getImages("паспорт"),
		"tk_book_images":              getImages("трудовая"),
		"characteristic_images":       getImages("характеристика"),
	})
}

func ApproveApplication(c *fiber.Ctx) error {
	var body struct {
		ID uint `json:"id"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Неверный формат запроса")
	}

	var application models.Application
	if err := database.DB.First(&application, body.ID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Заявление не найдено")
	}

	application.Status = "approved"
	if err := database.DB.Save(&application).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка при сохранении")
	}

	return c.SendStatus(fiber.StatusOK)
}
