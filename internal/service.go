package internal

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ktr0731/go-fuzzyfinder"
)

func ScanValute(root string) []Note {
	var notes []Note
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if strings.HasSuffix(d.Name(), "md") {
				note := Note{FileName: d.Name(), FilePath: path}
				notes = append(notes, note)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("scan error %s\n", err)
	}
	return notes
}

func EditFile(filePath string) {
	cmd := exec.Command(config.Editor, filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("command ended with error, %v\n", err)
		return
	}
}

func GetFile() string {
	noteList := ScanValute(config.MainVault)
	idx, err := fuzzyfinder.Find(
		noteList,
		func(i int) string {
			return noteList[i].FileName
		},

		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			contentBytes, err := os.ReadFile(noteList[i].FilePath)
			if err != nil {
				return fmt.Sprintf("Ошибка чтения файла: %v", err)
			}
			fileContent := string(contentBytes)
			return fmt.Sprintf("=== %s ===\nПуть: %s\n\n%s",
				noteList[i].FileName,
				noteList[i].FilePath,
				fileContent,
			)
		}))
	if err != nil {
		log.Fatal(err)
	}
	return noteList[idx].FilePath
}

func HandlerFile(fileName string) {
	notes := ScanValute(config.MainVault)
	for _, item := range notes {
		if item.FileName == fileName {
			EditFile(item.FilePath)
			return
		}
	}
	fullPath := filepath.Join(config.MainVault, fileName)
	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		fmt.Printf("create error %v\n", err)
		return
	}
	template, err := os.ReadFile(config.TemplateNote)
	if err != nil {
		file.Close()
		EditFile(fullPath)
		return
	}
	_, err = file.Write(template)
	if err != nil {
		fmt.Printf("write tmp in new file error %v", err)
	}
	file.Close()
	EditFile(fullPath)
}

func InlineNote(inlineText string) error {
	now := time.Now()
	fileName := now.Format("2006-01-02 15_04_05.md")
	filePath := filepath.Join(config.Path_to_inline_note, fileName)
	err := os.WriteFile(filePath, []byte(inlineText), 0600)
	if err != nil {
		return fmt.Errorf("failed to write inline note: %w", err)
	}
	fmt.Printf("\033[32mNote saved successfully: %s\033[0m\n", fileName)
	return nil
}
