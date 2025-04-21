package controllers

import (
	"archive/zip"
	"att_service/database"
	"att_service/models"
	"att_service/services"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func ChangeUserPhoto(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	// Получаем пользователя
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Пользователь не найден"})
	}

	// Получаем файл
	file, err := c.FormFile("user_photo")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Файл не загружен"})
	}

	// Формируем новое имя папки
	newFolderName := fmt.Sprintf("%s_%s_%s_%s",
		user.JestID,
		sanitizeString(user.SurnameInIp),
		sanitizeString(user.NameInIp),
		sanitizeString(user.LastnameInIp),
	)
	newPath := filepath.Join("./uploads", newFolderName)

	removeOldAvatars := func(dir string) {
		files, _ := os.ReadDir(dir)
		for _, f := range files {
			if strings.HasPrefix(f.Name(), "avatar.") || strings.HasPrefix(f.Name(), "avatar_") {
				_ = os.Remove(filepath.Join(dir, f.Name())) // удаляем старые аватарки
			}
		}
	}

	// Если имя актуально — сохраняем прямо туда
	if user.StoragePath == newPath {
		if err := os.MkdirAll(newPath, 0755); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось создать директорию"})
		}

		removeOldAvatars(newPath) // ← вызов удаления

		savePath := filepath.Join(newPath, "avatar"+filepath.Ext(file.Filename))
		if err := c.SaveFile(file, savePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось сохранить файл"})
		}

		return c.JSON(fiber.Map{"success": true})
	}

	// Имя устарело — переносим содержимое (кроме старого avatar*)
	filesToMove := []string{}
	if entries, err := os.ReadDir(user.StoragePath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && !strings.HasPrefix(entry.Name(), "avatar") {
				filesToMove = append(filesToMove, entry.Name())
			}
		}
	}

	// Создаем новую папку
	if err := os.MkdirAll(newPath, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось создать новую папку"})
	}

	// Переносим файлы
	for _, name := range filesToMove {
		oldFilePath := filepath.Join(user.StoragePath, name)
		newFilePath := filepath.Join(newPath, name)
		os.Rename(oldFilePath, newFilePath)
	}

	// Удаляем старую папку
	os.RemoveAll(user.StoragePath)

	// Сохраняем новую аватарку
	avatarPath := filepath.Join(newPath, "avatar"+filepath.Ext(file.Filename))
	if err := c.SaveFile(file, avatarPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось сохранить аватарку"})
	}

	// Обновляем путь в БД
	user.StoragePath = newPath
	database.DB.Save(&user)

	return c.JSON(fiber.Map{"success": true})
}

// GetUserProfile возвращает страницу профиля студента (pages/student_page.html)
func GetUserProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusUnauthorized).SendString("Необходима авторизация")
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка загрузки пользователя")
	}

	avatarPath := ""
	if user.StoragePath != "" {
		files, _ := os.ReadDir(user.StoragePath)
		for _, file := range files {
			if strings.Contains(file.Name(), "avatar") {
				avatarPath = "/uploads/" + filepath.Base(user.StoragePath) + "/" + file.Name()
				break
			}
		}
	}

	return services.Render(c, user.Role, "main.html", fiber.Map{
		"name_in_ip":     user.NameInIp,
		"name_in_rp":     user.NameInRp,
		"name_in_dp":     user.NameInDp,
		"surname_in_ip":  user.SurnameInIp,
		"surname_in_rp":  user.SurnameInRp,
		"surname_in_dp":  user.SurnameInDp,
		"lastname_in_ip": user.LastnameInIp,
		"lastname_in_rp": user.LastnameInRp,
		"lastname_in_dp": user.LastnameInDp,
		"email":          user.Email,
		"mail":           user.Mail,
		"work_phone":     user.WorkPhone,
		"mobile_phone":   user.MobilePhone,
		"sex":            user.Sex,
		"status":         user.Status,
		"role":           user.Role,
		"avatar":         avatarPath,
		"path":           c.Path(),
		"decline_reason": user.DeclineReason,
	})
}

func GetUserDocuments(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusUnauthorized).SendString("Необходима авторизация")
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка загрузки пользователя")
	}

	avatarPath := ""
	if user.StoragePath != "" {
		files, _ := os.ReadDir(user.StoragePath)
		for _, file := range files {
			if strings.Contains(file.Name(), "avatar") {
				avatarPath = "/uploads/" + filepath.Base(user.StoragePath) + "/" + file.Name()
				break
			}
		}
	}

	// Если папка не задана — возвращаем пустую форму
	if user.StoragePath == "" {
		return services.Render(c, user.Role, "documents.html", fiber.Map{
			"status":          user.Status,
			"passport_serial": "",
			"passport_num":    "",
			"unit_code":       "",
			"passport_issue":  "",
			"passport_date":   "",
			"bithday_date":    "",
			"born_place":      "",
			"registr_address": "",
			"snils_num":       "",
			"diplom_num":      "",
			"avatar":          avatarPath,
			"path":            c.Path(),

			"passport_images": []string{},
			"snils_images":    []string{},
			"diplom_images":   []string{},
		})
	}

	// Загружаем данные из БД
	var passport models.Passport
	database.DB.Where("user_id = ?", userID).First(&passport)

	var eduDoc models.EducationDocument
	database.DB.Where("user_id = ?", userID).First(&eduDoc)

	// Ищем изображения в папке
	passportImages := findImagesInDir(user.StoragePath, "паспорт")
	snilsImages := findImagesInDir(user.StoragePath, "снилс")
	diplomImages := findImagesInDir(user.StoragePath, "диплом")

	return services.Render(c, user.Role, "documents.html", fiber.Map{
		"status":          user.Status,
		"passport_serial": passport.PassportSeries,
		"passport_num":    passport.PassportNumber,
		"unit_code":       passport.PassportDivisionCode,
		"passport_issue":  passport.PassportIssuedBy,
		"passport_date":   formatDate(&passport.PassportIssueDate),
		"bithday_date":    formatDate(&passport.BirthDate),
		"born_place":      passport.BirthPlace,
		"registr_address": passport.RegistrationAddress,
		"role":            user.Role,

		"snils_num":  user.Snils,
		"diplom_num": eduDoc.DiplomaRegNumber,

		"passport_images": passportImages,
		"snils_images":    snilsImages,
		"diplom_images":   diplomImages,
		"avatar":          avatarPath,
		"path":            c.Path(),
		"decline_reason":  user.DeclineReason,
	})
}

func GetDeclineReasons(c *fiber.Ctx) error {
	// Получаем userID из авторизации (middleware)
	userID := c.Locals("userID")

	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "Unauthorized"})
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "User not found"})
	}

	reasons := make(map[string]bool)
	explanation := ""

	parts := strings.Split(user.DeclineReason, " | ")

	if len(parts) > 0 {
		reasonsPart := parts[0]

		if strings.Contains(reasonsPart, "ФИО") {
			reasons["invalid_name"] = true
		}
		if strings.Contains(reasonsPart, "контакт") {
			reasons["invalid_contacts"] = true
		}
		if strings.Contains(reasonsPart, "документ") {
			reasons["no_documents"] = true
		}
	}

	if len(parts) > 1 {
		explanation = strings.TrimSpace(parts[1])
	}

	return c.JSON(fiber.Map{
		"success": true,
		"list": fiber.Map{
			"invalid_name":     reasons["invalid_name"],
			"invalid_contacts": reasons["invalid_contacts"],
			"no_documents":     reasons["no_documents"],
			"explanation":      explanation,
		},
	})
}

func parseDate(dateStr string) time.Time {
	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{} // Возврат нулевой даты
	}
	return parsed
}

func CheckUserDataCorrectness(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "unauthorized"})
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "user not found"})
	}

	missing := []string{}

	// Личные данные
	if user.SurnameInIp == "" {
		missing = append(missing, "Фамилия (Именит. падеж)")
	}
	if user.SurnameInRp == "" {
		missing = append(missing, "Фамилия (Родит. падеж)")
	}
	if user.SurnameInDp == "" {
		missing = append(missing, "Фамилия (Дател. падеж)")
	}
	if user.NameInIp == "" {
		missing = append(missing, "Имя (Именит. падеж)")
	}
	if user.NameInRp == "" {
		missing = append(missing, "Имя (Родит. падеж)")
	}
	if user.NameInDp == "" {
		missing = append(missing, "Имя (Дател. падеж)")
	}
	if user.LastnameInIp == "" {
		missing = append(missing, "Отчество (Именит. падеж)")
	}
	if user.LastnameInRp == "" {
		missing = append(missing, "Отчество (Родит. падеж)")
	}
	if user.LastnameInDp == "" {
		missing = append(missing, "Отчество (Дател. падеж)")
	}
	if user.Email == "" {
		missing = append(missing, "Электронная почта")
	}
	if user.MobilePhone == "" {
		missing = append(missing, "Мобильный телефон")
	}
	if user.Mail == "" {
		missing = append(missing, "Почтовый адрес")
	}
	if user.Sex == "" {
		missing = append(missing, "Пол")
	}

	// Паспорт
	var passport models.Passport
	if err := database.DB.Where("user_id = ?", userID).First(&passport).Error; err == nil {
		if passport.PassportSeries == "" {
			missing = append(missing, "Серия паспорта")
		}
		if passport.PassportNumber == "" {
			missing = append(missing, "Номер паспорта")
		}
		if passport.PassportDivisionCode == "" {
			missing = append(missing, "Код подразделения")
		}
		if passport.PassportIssuedBy == "" {
			missing = append(missing, "Кем выдан паспорт")
		}
		if passport.PassportIssueDate.IsZero() {
			missing = append(missing, "Дата выдачи паспорта")
		}
		if passport.BirthDate.IsZero() {
			missing = append(missing, "Дата рождения")
		}
		if passport.BirthPlace == "" {
			missing = append(missing, "Место рождения")
		}
		if passport.RegistrationAddress == "" {
			missing = append(missing, "Адрес регистрации")
		}
	} else {
		missing = append(missing, "Данные паспорта отсутствуют")
	}

	// СНИЛС
	if user.Snils == "" {
		missing = append(missing, "СНИЛС")
	}

	// Диплом
	var edu models.EducationDocument
	if err := database.DB.Where("user_id = ?", userID).First(&edu).Error; err == nil {
		if edu.DiplomaRegNumber == "" {
			missing = append(missing, "Номер диплома")
		}
	} else {
		missing = append(missing, "Данные диплома отсутствуют")
	}

	// Проверка сканов документов (в БД и папке пользователя)
	var docs []models.UserDocument
	database.DB.Where("user_id = ?", userID).Find(&docs)

	countByType := map[string]int{}
	for _, doc := range docs {
		countByType[doc.DocumentType]++
	}

	if countByType["паспорт"] == 0 {
		missing = append(missing, "Сканы паспорта")
	}
	if countByType["снилс"] == 0 {
		missing = append(missing, "Скан СНИЛС")
	}
	if countByType["диплом"] == 0 {
		missing = append(missing, "Сканы диплома")
	}

	// Проверка наличия аватарки
	hasAvatar := false
	if user.StoragePath != "" {
		files, _ := os.ReadDir(user.StoragePath)
		for _, file := range files {
			if strings.HasPrefix(file.Name(), "avatar.") {
				hasAvatar = true
				break
			}
		}
	}
	if !hasAvatar {
		missing = append(missing, "Фотография профиля (аватарка)")
	}

	if len(missing) > 0 {
		return c.JSON(fiber.Map{
			"success": false,
			"list":    missing,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

func ApproveUserBySelf(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "unauthorized"})
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "user not found"})
	}

	if user.Status == "approved" {
		return c.JSON(fiber.Map{"success": true, "message": "Уже подтверждено"})
	}

	// Обновляем статус
	user.Status = "pending"
	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "save error"})
	}

	return c.JSON(fiber.Map{"success": true})
}

func UploadUserDocuments(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Требуется авторизация"})
	}
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Пользователь не найден"})
	}

	// Считываем текстовые поля формы
	formFields := map[string]string{
		"passport_serial": c.FormValue("passport_serial"),
		"unit_code":       c.FormValue("unit_code"),
		"passport_num":    c.FormValue("passport_num"),
		"passport_date":   c.FormValue("passport_date"),
		"passport_issue":  c.FormValue("passport_issue"),
		"bithday_date":    c.FormValue("bithday_date"),
		"born_place":      c.FormValue("born_place"),
		"registr_address": c.FormValue("registr_address"),
		"snils_num":       c.FormValue("snils_num"),
		"diplom_num":      c.FormValue("diplom_num"),
	}

	// Обновляем путь хранения файлов, если он изменился
	oldPath := user.StoragePath
	newFolder := fmt.Sprintf("%s_%s_%s_%s",
		user.JestID,
		sanitizeString(user.SurnameInIp),
		sanitizeString(user.NameInIp),
		sanitizeString(user.LastnameInIp),
	)
	newPath := filepath.Join("./uploads", newFolder)
	if oldPath != "" && oldPath != newPath {
		if err := os.MkdirAll(newPath, 0755); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка создания новой папки"})
		}
		// Переносим все файлы из старой папки в новую
		if entries, err := os.ReadDir(oldPath); err == nil {
			for _, f := range entries {
				src := filepath.Join(oldPath, f.Name())
				dst := filepath.Join(newPath, f.Name())
				os.Rename(src, dst)
			}
		}
		os.RemoveAll(oldPath)
		user.StoragePath = newPath
		database.DB.Save(&user)
	}

	// --- Обновление данных паспорта ---
	var passport models.Passport
	if err := database.DB.Where("user_id = ?", userID).First(&passport).Error; err != nil {
		if formFields["passport_serial"] != "" || formFields["passport_num"] != "" ||
			formFields["unit_code"] != "" || formFields["passport_issue"] != "" ||
			formFields["passport_date"] != "" || formFields["bithday_date"] != "" ||
			formFields["born_place"] != "" || formFields["registr_address"] != "" {
			passport = models.Passport{UserID: userID}
			passport.PassportSeries = formFields["passport_serial"]
			passport.PassportNumber = formFields["passport_num"]
			passport.PassportDivisionCode = formFields["unit_code"]
			passport.PassportIssuedBy = formFields["passport_issue"]
			passport.PassportIssueDate = parseDate(formFields["passport_date"])
			passport.BirthDate = parseDate(formFields["bithday_date"])
			passport.BirthPlace = formFields["born_place"]
			passport.RegistrationAddress = formFields["registr_address"]
			database.DB.Create(&passport)
		}
	} else {
		updated := false
		if formFields["passport_serial"] != "" && formFields["passport_serial"] != passport.PassportSeries {
			passport.PassportSeries = formFields["passport_serial"]
			updated = true
		}
		if formFields["passport_num"] != "" && formFields["passport_num"] != passport.PassportNumber {
			passport.PassportNumber = formFields["passport_num"]
			updated = true
		}
		if formFields["unit_code"] != "" && formFields["unit_code"] != passport.PassportDivisionCode {
			passport.PassportDivisionCode = formFields["unit_code"]
			updated = true
		}
		if formFields["passport_issue"] != "" && formFields["passport_issue"] != passport.PassportIssuedBy {
			passport.PassportIssuedBy = formFields["passport_issue"]
			updated = true
		}
		if formFields["passport_date"] != "" {
			newDate := parseDate(formFields["passport_date"])
			if !newDate.IsZero() && !newDate.Equal(passport.PassportIssueDate) {
				passport.PassportIssueDate = newDate
				updated = true
			}
		}
		if formFields["bithday_date"] != "" {
			newDate := parseDate(formFields["bithday_date"])
			if !newDate.IsZero() && !newDate.Equal(passport.BirthDate) {
				passport.BirthDate = newDate
				updated = true
			}
		}
		if formFields["born_place"] != "" && formFields["born_place"] != passport.BirthPlace {
			passport.BirthPlace = formFields["born_place"]
			updated = true
		}
		if formFields["registr_address"] != "" && formFields["registr_address"] != passport.RegistrationAddress {
			passport.RegistrationAddress = formFields["registr_address"]
			updated = true
		}
		if updated {
			database.DB.Save(&passport)
		}
	}

	// --- Обновление SNILS ---
	if formFields["snils_num"] != "" && formFields["snils_num"] != user.Snils {
		user.Snils = formFields["snils_num"]
		database.DB.Save(&user)
	}

	// --- Обновление образовательного документа ---
	var edu models.EducationDocument
	if err := database.DB.Where("user_id = ?", userID).First(&edu).Error; err != nil {
		if formFields["diplom_num"] != "" {
			edu = models.EducationDocument{UserID: userID, DiplomaRegNumber: formFields["diplom_num"]}
			database.DB.Create(&edu)
		}
	} else {
		if formFields["diplom_num"] != "" && formFields["diplom_num"] != edu.DiplomaRegNumber {
			edu.DiplomaRegNumber = formFields["diplom_num"]
			database.DB.Save(&edu)
		}
	}

	// Обработка загруженных файлов
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(fiber.Map{"success": true})
	}

	removeFilesByType := func(dir, prefix string) {
		files, _ := os.ReadDir(dir)
		for _, f := range files {
			if strings.Contains(f.Name(), prefix) {
				_ = os.Remove(filepath.Join(dir, f.Name()))
			}
		}
	}

	// --- Мультизагрузка паспортных сканов (поле "passport_all") ---
	if files, exists := form.File["passport_all"]; exists && len(files) > 0 {
		// Удаляем старые файлы типа "паспорт"
		var oldDocs []models.UserDocument
		database.DB.Where("user_id = ? AND document_type = ?", userID, "паспорт").Find(&oldDocs)
		for _, doc := range oldDocs {
			os.Remove(doc.FilePath)
		}
		database.DB.Where("user_id = ? AND document_type = ?", userID, "паспорт").Delete(&models.UserDocument{})

		removeFilesByType(user.StoragePath, "паспорт")

		// Загружаем новые файлы, нумеруя их начиная с 1
		for i, file := range files {
			ext := filepath.Ext(file.Filename)
			base := generateFolderName(user.JestID, "паспорт")
			unique := fmt.Sprintf("%s_%d%s", base, i+1, ext)
			savePath := filepath.Join(user.StoragePath, unique)
			if err := c.SaveFile(file, savePath); err != nil {
				log.Printf("Ошибка при сохранении файла %s: %v", file.Filename, err)
				continue
			}
			doc := models.UserDocument{
				UserID:       userID,
				DocumentName: file.Filename,
				DocumentType: "паспорт",
				FilePath:     savePath,
			}
			database.DB.Create(&doc)
		}
	}

	// --- Мультизагрузка файлов для снилса (поле "snils_img") ---
	if files, exists := form.File["snils_img"]; exists && len(files) > 0 {
		var oldDocs []models.UserDocument
		database.DB.Where("user_id = ? AND document_type = ?", userID, "снилс").Find(&oldDocs)
		for _, doc := range oldDocs {
			os.Remove(doc.FilePath)
		}
		database.DB.Where("user_id = ? AND document_type = ?", userID, "снилс").Delete(&models.UserDocument{})

		removeFilesByType(user.StoragePath, "снилс")

		for i, file := range files {
			ext := filepath.Ext(file.Filename)
			base := generateFolderName(user.JestID, "снилс")
			unique := fmt.Sprintf("%s_%d%s", base, i+1, ext)
			savePath := filepath.Join(user.StoragePath, unique)
			if err := c.SaveFile(file, savePath); err != nil {
				log.Printf("Ошибка при сохранении файла %s: %v", file.Filename, err)
				continue
			}
			doc := models.UserDocument{
				UserID:       userID,
				DocumentName: file.Filename,
				DocumentType: "снилс",
				FilePath:     savePath,
			}
			database.DB.Create(&doc)
		}
	}

	// --- Мультизагрузка файлов для диплома (поле "diplom_img") ---
	if files, exists := form.File["diplom_img"]; exists && len(files) > 0 {
		var oldDocs []models.UserDocument
		database.DB.Where("user_id = ? AND document_type = ?", userID, "диплом").Find(&oldDocs)
		for _, doc := range oldDocs {
			os.Remove(doc.FilePath)
		}
		database.DB.Where("user_id = ? AND document_type = ?", userID, "диплом").Delete(&models.UserDocument{})

		removeFilesByType(user.StoragePath, "диплом")

		for i, file := range files {
			ext := filepath.Ext(file.Filename)
			base := generateFolderName(user.JestID, "диплом")
			unique := fmt.Sprintf("%s_%d%s", base, i+1, ext)
			savePath := filepath.Join(user.StoragePath, unique)
			if err := c.SaveFile(file, savePath); err != nil {
				log.Printf("Ошибка при сохранении файла %s: %v", file.Filename, err)
				continue
			}
			doc := models.UserDocument{
				UserID:       userID,
				DocumentName: file.Filename,
				DocumentType: "диплом",
				FilePath:     savePath,
			}
			database.DB.Create(&doc)
		}
	}

	// --- Мультизагрузка дополнительных документов (поля с префиксом "new_doc_img_") ---
	for field, files := range form.File {
		if strings.HasPrefix(field, "new_doc_img_") && len(files) > 0 {
			docNum := strings.TrimPrefix(field, "new_doc_img_")
			docNameArr := form.Value["new_doc_"+docNum]
			docType := "документ"
			if len(docNameArr) > 0 {
				docType = docNameArr[0]
			}
			var oldDocs []models.UserDocument
			database.DB.Where("user_id = ? AND document_type = ?", userID, docType).Find(&oldDocs)
			for _, doc := range oldDocs {
				os.Remove(doc.FilePath)
			}
			database.DB.Where("user_id = ? AND document_type = ?", userID, docType).Delete(&models.UserDocument{})

			removeFilesByType(user.StoragePath, docType)

			for i, file := range files {
				ext := filepath.Ext(file.Filename)
				base := generateFolderName(user.JestID, docType)
				unique := fmt.Sprintf("%s_%d%s", base, i+1, ext)
				savePath := filepath.Join(user.StoragePath, unique)
				if err := c.SaveFile(file, savePath); err != nil {
					log.Printf("Ошибка при сохранении файла %s: %v", file.Filename, err)
					continue
				}
				doc := models.UserDocument{
					UserID:       userID,
					DocumentName: file.Filename,
					DocumentType: docType,
					FilePath:     savePath,
				}
				database.DB.Create(&doc)
			}
		}
	}

	return c.JSON(fiber.Map{"success": true})
}

func findImagesInDir(dir, docType string) []string {
	var result []string
	allowedExt := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return result
	}

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if !file.IsDir() && strings.Contains(file.Name(), docType) && allowedExt[ext] {
			result = append(result, "/uploads/"+filepath.Base(dir)+"/"+file.Name())
		}
	}
	return result
}

func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func SaveMainPageData(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var input struct {
		SurnameInIp  string `json:"surname_in_ip"`
		SurnameInRp  string `json:"surname_in_rp"`
		SurnameInDp  string `json:"surname_in_dp"`
		NameInIp     string `json:"name_in_ip"`
		NameInRp     string `json:"name_in_rp"`
		NameInDp     string `json:"name_in_dp"`
		LastnameInIp string `json:"lastname_in_ip"`
		LastnameInRp string `json:"lastname_in_rp"`
		LastnameInDp string `json:"lastname_in_dp"`
		Email        string `json:"email"`
		Mail         string `json:"mail"`
		WorkPhone    string `json:"work_phone"`
		MobilePhone  string `json:"mobile_phone"`
		Sex          string `json:"sex"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный формат данных",
		})
	}

	// Загружаем текущего пользователя
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Пользователь не найден"})
	}

	// Обновляем только переданные поля (если они не пустые)
	if input.SurnameInIp != "" {
		user.SurnameInIp = input.SurnameInIp
	}
	if input.SurnameInRp != "" {
		user.SurnameInRp = input.SurnameInRp
	}
	if input.SurnameInDp != "" {
		user.SurnameInDp = input.SurnameInDp
	}
	if input.NameInIp != "" {
		user.NameInIp = input.NameInIp
	}
	if input.NameInRp != "" {
		user.NameInRp = input.NameInRp
	}
	if input.NameInDp != "" {
		user.NameInDp = input.NameInDp
	}
	if input.LastnameInIp != "" {
		user.LastnameInIp = input.LastnameInIp
	}
	if input.LastnameInRp != "" {
		user.LastnameInRp = input.LastnameInRp
	}
	if input.LastnameInDp != "" {
		user.LastnameInDp = input.LastnameInDp
	}
	if input.Email != "" {
		user.Email = input.Email
	}
	if input.Mail != "" {
		user.Mail = input.Mail
	}
	if input.WorkPhone != "" {
		user.WorkPhone = input.WorkPhone
	}
	if input.MobilePhone != "" {
		user.MobilePhone = input.MobilePhone
	}
	if input.Sex != "" {
		user.Sex = input.Sex
	}

	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка сохранения данных"})
	}

	return c.JSON(fiber.Map{"success": true})
}

func generateApplicationNumber() string {
	var count int64
	database.DB.Model(&models.Application{}).Count(&count)
	return fmt.Sprintf("06-10-%d", count+1)
}

func SaveUserApplication(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusUnauthorized).SendString("Необходима авторизация")
	}

	// Чтение формы (multipart/form-data)
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Ошибка чтения формы"})
	}

	// Чтение текстовых полей
	values := form.Value
	get := func(key string) string {
		if val, ok := values[key]; ok && len(val) > 0 {
			return val[0]
		}
		return ""
	}

	// Обработка согласия
	consent := get("consent")
	if consent != "true" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Необходимо согласие на обработку персональных данных"})
	}

	// Сохраняем заявку
	app := models.Application{
		UserID:                    userID,
		NativeLanguage:            get("native_language"),
		Citizenship:               get("citizenship"),
		MaritalStatus:             get("marital_status"),
		Organization:              get("organization"),
		JobPosition:               get("job_position"),
		RequestedCategory:         get("requested_category"),
		BasisForAttestation:       get("basis_for_attestation"),
		ExistingCategory:          get("existing_category"),
		ExistingCategoryTerm:      get("existing_category_term"),
		WorkExperience:            get("work_experience"),
		CurrentPositionExperience: get("current_position_experience"),
		AwardsInfo:                get("awards_info"),
		TrainingInfo:              get("training_info"),
		Memberships:               get("memberships"),
		Consent:                   true,
		ApplicationNumber:         generateApplicationNumber(),
		Status:                    "pending",
	}

	if err := database.DB.Create(&app).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка сохранения заявки"})
	}

	// Получаем пользователя и его папку
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка получения пользователя"})
	}
	if user.StoragePath == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "У пользователя не указан путь к папке"})
	}

	// Документные типы и поля формы
	docFields := map[string]string{
		"diplom_img":      "диплом",
		"diplom_jest_img": "диплом_жестовый",
		"passport_all":    "паспорт",
		"tk_book":         "трудовая",
		"characteristic":  "характеристика",
	}

	// Удаление старых документов
	for _, docType := range docFields {
		database.DB.Where("user_id = ? AND document_type = ?", userID, docType).Delete(&models.UserDocument{})
	}

	// Сохранение новых файлов
	for field, docType := range docFields {
		files := form.File[field]
		for i, file := range files {
			ext := filepath.Ext(file.Filename)
			unique := fmt.Sprintf("%s_%s_%d%s", user.JestID, sanitizeString(docType), i+1, ext)
			savePath := filepath.Join(user.StoragePath, unique)
			if err := c.SaveFile(file, savePath); err != nil {
				log.Printf("Ошибка сохранения файла %s: %v", file.Filename, err)
				continue
			}
			doc := models.UserDocument{
				UserID:       userID,
				DocumentName: file.Filename,
				DocumentType: docType,
				FilePath:     savePath,
			}
			database.DB.Create(&doc)
		}
	}
	var passport models.Passport
	database.DB.Where("user_id = ?", user.ID).First(&passport)

	var edu models.EducationDocument
	database.DB.Where("user_id = ?", user.ID).First(&edu)

	// В конце SaveUserApplication:
	if err := GenerateApplicationODT(user, app, passport, edu); err != nil {
		log.Println("Ошибка генерации ODT:", err)
	}

	return c.JSON(fiber.Map{"success": true})
}
func GetUserCreateApplicationPage(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusUnauthorized).SendString("Необходима авторизация")
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка загрузки пользователя")
	}

	// --- Аватарка ---
	avatarPath := ""
	if user.StoragePath != "" {
		if files, err := os.ReadDir(user.StoragePath); err == nil {
			for _, f := range files {
				if strings.HasPrefix(f.Name(), "avatar") {
					avatarPath = "/uploads/" + filepath.Base(user.StoragePath) + "/" + f.Name()
					break
				}
			}
		}
	}

	// --- Последняя заявка ---
	var application models.Application
	database.DB.Where("user_id = ?", userID).Order("created_at DESC").First(&application)

	// --- Документы ---
	diplomImages := findImagesInDir(user.StoragePath, "диплом")
	diplomJestImages := findImagesInDir(user.StoragePath, "диплом_жестовый")
	passportImages := findImagesInDir(user.StoragePath, "паспорт")
	tkBookImages := findImagesInDir(user.StoragePath, "трудовая")
	characteristicImages := findImagesInDir(user.StoragePath, "характеристика")

	return services.Render(c, "student", "create_application.html", fiber.Map{
		"status":                      user.Status,
		"role":                        user.Role,
		"avatar":                      avatarPath,
		"path":                        c.Path(),
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
		"diplom_images":               diplomImages,
		"diplom_jest_images":          diplomJestImages,
		"passport_images":             passportImages,
		"tk_book_images":              tkBookImages,
		"characteristic_images":       characteristicImages,
	})
}

func getDocumentPaths(userID uint, docType string) []string {
	var docs []models.UserDocument
	database.DB.Where("user_id = ? AND document_type = ?", userID, docType).Find(&docs)
	paths := []string{}
	for _, doc := range docs {
		if strings.HasPrefix(doc.FilePath, "./uploads/") {
			paths = append(paths, "/uploads/"+filepath.Base(filepath.Dir(doc.FilePath))+"/"+filepath.Base(doc.FilePath))
		}
	}
	return paths
}
func GetUserApplicationPage(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка загрузки пользователя")
	}

	avatarPath := ""
	if user.StoragePath != "" {
		files, _ := os.ReadDir(user.StoragePath)
		for _, file := range files {
			if strings.Contains(file.Name(), "avatar") {
				avatarPath = "/uploads/" + filepath.Base(user.StoragePath) + "/" + file.Name()
				break
			}
		}
	}

	// Загружаем заявки с причиной отказа
	var applications []models.Application
	if err := database.DB.Preload("Decline").Where("user_id = ?", userID).Find(&applications).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка загрузки заявок")
	}

	// Структура для отображения в шаблоне
	type ApplicationItem struct {
		ID      uint
		Number  string
		Status  string
		Decline *models.ApplicationDecline
	}
	var list []ApplicationItem
	for _, app := range applications {
		list = append(list, ApplicationItem{
			ID:      app.ID,
			Number:  app.ApplicationNumber,
			Status:  strings.ToLower(app.Status),
			Decline: app.Decline,
		})
	}

	return services.Render(c, "student", "applications.html", fiber.Map{
		"status":       user.Status,
		"role":         user.Role,
		"avatar":       avatarPath,
		"id":           userID,
		"path":         c.Path(),
		"applications": list,
	})
}

func GetApplicationDeclineReason(c *fiber.Ctx) error {
	type RequestBody struct {
		ID uint `json:"id"`
	}

	var req RequestBody
	if err := c.BodyParser(&req); err != nil || req.ID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Некорректные или отсутствующие данные в запросе",
		})
	}

	// Загружаем заявку с отказом
	var app models.Application
	if err := database.DB.Preload("Decline").
		First(&app, req.ID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Заявка не найдена",
		})
	}

	if app.Decline == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Причины отказа не найдены",
		})
	}

	// Разбираем причины
	reasonMap := map[string]bool{}
	for _, r := range strings.Split(app.Decline.Reasons, ",") {
		reason := strings.TrimSpace(r)
		if reason != "" {
			reasonMap[reason] = true
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"list": fiber.Map{
			"invalid_name":     reasonMap["invalid_name"],
			"invalid_contacts": reasonMap["invalid_contacts"],
			"no_documents":     reasonMap["no_documents"],
			"explanation":      app.Decline.Explanation,
		},
	})
}

func GetExamWaitingPage(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	userRole := c.Locals("userRole").(string)

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(500).SendString("Ошибка загрузки пользователя")
	}

	examID := c.Params("exam_id") // предполагаем, что в URL будет /exam/waiting/:exam_id

	return services.Render(c, "admin", "exam_procedure/exam_waiting-page.html", fiber.Map{
		"role":    userRole,
		"id":      user.ID,
		"name":    fmt.Sprintf("%s %s", user.SurnameInIp, user.NameInIp),
		"exam_id": examID,
	})
}

func GetExamStudentPage(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	userRole := c.Locals("userRole").(string)

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(500).SendString("Ошибка загрузки пользователя")
	}

	examID := c.Params("exam_id") // например, из URL /exam/student/:exam_id/:student_id
	studentID := c.Params("student_id")

	return services.Render(c, "admin", "exam_procedure/exam_person-page.html", fiber.Map{
		"role":       userRole,
		"id":         user.ID,
		"name":       fmt.Sprintf("%s %s", user.SurnameInIp, user.NameInIp),
		"exam_id":    examID,
		"student_id": studentID, // можно использовать в шаблоне или JS
	})
}

func GenerateApplicationODT(user models.User, app models.Application, passport models.Passport, edu models.EducationDocument) error {
	if user.StoragePath == "" {
		return fmt.Errorf("не указан путь хранения пользователя")
	}

	templatePath := "./templates_odt/Application_FIXED.ott"
	workDir := filepath.Join(user.StoragePath, "odt_temp")
	outputODT := filepath.Join(user.StoragePath, fmt.Sprintf("%s_заявление.odt", user.JestID))

	os.RemoveAll(workDir)
	os.MkdirAll(workDir, os.ModePerm)

	// 1. Распаковать шаблон
	if err := unzip(templatePath, workDir); err != nil {
		return fmt.Errorf("ошибка распаковки шаблона: %v", err)
	}

	// 2. Копирование изображений
	imageMap := map[string]string{
		"passport_1":        "паспорт_разворот",
		"passport_2":        "паспорт_прописка",
		"diplom_image":      "диплом",
		"diplom_jest_image": "диплом_жестовый",
		"book":              "трудовая",
		"characteristics":   "характеристика",
	}
	picturesDir := filepath.Join(workDir, "Pictures")
	os.MkdirAll(picturesDir, os.ModePerm)
	replacedImages := map[string]string{}

	for placeholder, prefix := range imageMap {
		matches, _ := filepath.Glob(filepath.Join(user.StoragePath, prefix+"*"))
		if len(matches) > 0 {
			ext := filepath.Ext(matches[0])
			destName := placeholder + ext
			destPath := filepath.Join(picturesDir, destName)
			if err := copyFile(matches[0], destPath); err == nil {
				replacedImages[placeholder] = destName
			}
		}
	}

	// 3. Заменить текст и изображения в content.xml
	contentPath := filepath.Join(workDir, "content.xml")
	contentBytes, err := os.ReadFile(contentPath)
	if err != nil {
		return fmt.Errorf("не удалось прочитать content.xml: %v", err)
	}
	updatedContent := string(contentBytes)

	// 3.1. Заменим пути к изображениям
	for key, fname := range replacedImages {
		updatedContent = strings.ReplaceAll(updatedContent, "Pictures/{{"+key+"}}", "Pictures/"+fname)
	}

	// 3.2. Подстановка текстов
	textMap := map[string]string{
		"{{surname}}":                     user.SurnameInIp,
		"{{first_name}}":                  user.NameInIp,
		"{{patronymic}}":                  user.LastnameInIp,
		"{{birth_date}}":                  passport.BirthDate.Format("02.01.2006"),
		"{{birth_place}}":                 passport.BirthPlace,
		"{{gender}}":                      map[string]string{"male": "мужской", "female": "женский"}[user.Sex],
		"{{native_language}}":             app.NativeLanguage,
		"{{citizenship}}":                 app.Citizenship,
		"{{marital_status}}":              app.MaritalStatus,
		"{{organization}}":                app.Organization,
		"{{job_position}}":                app.JobPosition,
		"{{requested_category}}":          app.RequestedCategory,
		"{{basis_for_attestation}}":       app.BasisForAttestation,
		"{{existing_category}}":           app.ExistingCategory,
		"{{existing_category_term}}":      app.ExistingCategoryTerm,
		"{{work_experience}}":             app.WorkExperience,
		"{{current_position_experience}}": app.CurrentPositionExperience,
		"{{awards_info}}":                 app.AwardsInfo,
		"{{training_info}}":               app.TrainingInfo,
		"{{memberships}}":                 app.Memberships,
		"{{email}}":                       user.Email,
		"{{mobile_phone}}":                user.MobilePhone,
		"{{work_phone}}":                  user.WorkPhone,
		"{{home_phone}}":                  "", // нет отдельного поля
		"{{home_address}}":                passport.RegistrationAddress,
		"{{passport_series}}":             passport.PassportSeries,
		"{{passport_number}}":             passport.PassportNumber,
		"{{passport_issued_by}}":          passport.PassportIssuedBy,
		"{{passport_issue_date}}":         passport.PassportIssueDate.Format("02.01.2006"),
		"{{passport_division_code}}":      passport.PassportDivisionCode,
		"{{education}}":                   edu.DiplomaRegNumber,
		"{{education_document_scan}}":     "см. вложение",
		"{{application_number}}":          fmt.Sprintf("№%d", app.ID),
		"{{application_type}}":            "Первичная", // или другой параметр при необходимости
		"{{declared_specialization}}":     "",          // заполни при наличии поля
		"{{date}}":                        time.Now().Format("02.01.2006"),
		"{{consent}}":                     "Да",
	}

	for key, value := range textMap {
		updatedContent = strings.ReplaceAll(updatedContent, key, value)
	}

	if err := os.WriteFile(contentPath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("ошибка записи content.xml: %v", err)
	}

	// 4. Собираем обратно в odt
	if err := zipFolder(workDir, outputODT); err != nil {
		return fmt.Errorf("ошибка сборки odt: %v", err)
	}

	os.RemoveAll(workDir)
	return nil
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		fp := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fp, os.ModePerm)
			continue
		}
		os.MkdirAll(filepath.Dir(fp), os.ModePerm)
		dst, err := os.Create(fp)
		if err != nil {
			return err
		}
		defer dst.Close()
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		_, err = io.Copy(dst, rc)
		if err != nil {
			return err
		}
	}
	return nil
}

func zipFolder(sourceDir, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	w := zip.NewWriter(zipFile)
	defer w.Close()
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == sourceDir {
			return nil
		}
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			_, err := w.Create(relPath + "/")
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		fw, err := w.Create(relPath)
		if err != nil {
			return err
		}
		_, err = io.Copy(fw, f)
		return err
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// func GetUserApplicationPage(c *fiber.Ctx) error {
// 	userID, ok := c.Locals("userID").(uint)
// 	if !ok || userID == 0 {
// 		return c.Status(fiber.StatusUnauthorized).SendString("Необходима авторизация")
// 	}

// 	var application models.Application
// 	// Проверяем есть ли уже заявление у пользователя
// 	if err := database.DB.Where("user_id = ?", userID).First(&application).Error; err != nil {
// 		// Если нет — создаём пустое (опционально)
// 		application = models.Application{}
// 	}

// 	return services.Render(c, "student", "create_application.html", fiber.Map{
// 		"native_language":             application.NativeLanguage,
// 		"citizenship":                 application.Citizenship,
// 		"marital_status":              application.MaritalStatus,
// 		"organization":                application.Organization,
// 		"job_position":                application.JobPosition,
// 		"requested_category":          application.RequestedCategory,
// 		"basis_for_attestation":       application.BasisForAttestation,
// 		"existing_category":           application.ExistingCategory,
// 		"existing_category_term":      application.ExistingCategoryTerm,
// 		"work_experience":             application.WorkExperience,
// 		"current_position_experience": application.CurrentPositionExperience,
// 		"awards_info":                 application.AwardsInfo,
// 		"training_info":               application.TrainingInfo,
// 		"memberships":                 application.Memberships,
// 		"consent":                     application.Consent,
// 		"path":                        c.Path(),
// 	})
// }

// Вспомогательные функции.
func generateFolderName(jestID, docType string) string {
	safeDocType := sanitizeString(docType)
	return fmt.Sprintf("%s_%s", jestID, safeDocType)
}
