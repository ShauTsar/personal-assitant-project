package scanner

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func Scanner() {
	// Укажи путь к изображению чека
	imagePath := "C:\\Users\\novikov_n\\go\\src\\personal-assitant-project\\personal-assitant-server\\scanner\\check.png"

	// Опционально: установи язык, если изображение на нестандартном языке
	language := "rus"

	// Получи текст с помощью Tesseract OCR
	text, err := scanWithTesseract(imagePath, language)
	if err != nil {
		log.Fatal(err)
	}

	// Обработай распознанный текст
	processScannedText(text)

	// Удали изображение чека
	//err = deleteCheckImage(imagePath)
	//if err != nil {
	//	log.Println("Failed to delete check image:", err)
	//}
}

func scanWithTesseract(imagePath, language string) (string, error) {
	cmd := exec.Command("tesseract", imagePath, "stdout", "-l", language)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ошибка при выполнении Tesseract OCR: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func processScannedText(text string) {
	// Раздели текст на строки
	lines := strings.Split(text, "\n")

	// Проанализируй каждую строку и извлеки данные
	for _, line := range lines {
		// Здесь ты можешь добавить логику для обработки каждой строки текста
		// Например, поиск ключевых слов и извлечение соответствующих данных
		fmt.Println(line)
	}
}

func deleteCheckImage(imagePath string) error {
	// Удали изображение чека
	err := os.Remove(imagePath)
	if err != nil {
		return fmt.Errorf("ошибка при удалении изображения чека: %v", err)
	}

	fmt.Println("Изображение чека успешно удалено.")
	return nil
}
