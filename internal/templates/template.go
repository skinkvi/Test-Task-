package templates

import (
	"bytes"
	"html/template"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// ExecuteTemplate - функция для выполнения шаблона с заданными данными.
//
// templateName - имя файла шаблона, находящегося в папке templates.
// data - данные, с которыми будет выполнен шаблон.
// Возвращает строку с результатом выполнения шаблона и ошибку, если есть.
func ExecuteTemplate(templateName string, data interface{}) (string, error) {
	templatePath := filepath.Join("templates", templateName)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		logrus.Errorf("Error parsing template file: %v", err)

		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		logrus.Errorf("Error executing template: %v", err)

		return "", err
	}

	return buf.String(), nil
}
