package internal

import (
	"bufio"
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

func scanVault(root string) []Note {
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

func RemoveFile(filePath string) error {
	return os.Remove(filePath)
}

func EditFile(filePath string) error {
	var editor string

	if GlobalConfig.Editor != "" {
		if _, err := exec.LookPath(GlobalConfig.Editor); err == nil {
			editor = GlobalConfig.Editor
		} else {
			fmt.Printf("\x1b[33mWarning: Editor '%s' from config.toml not found. Trying system env...\x1b[0m\n", GlobalConfig.Editor)
		}
	}

	if editor == "" && os.Getenv("VISUAL") != "" {
		visualEnv := os.Getenv("VISUAL")
		if _, err := exec.LookPath(visualEnv); err == nil {
			editor = visualEnv
		}
	}

	if editor == "" && os.Getenv("EDITOR") != "" {
		editorEnv := os.Getenv("EDITOR")
		if _, err := exec.LookPath(editorEnv); err == nil {
			editor = editorEnv
		}
	}

	if editor == "" {
		fallbackEditors := []string{"nano", "notepad", "hx", "nvim", "vim", "vi"}
		for _, fallback := range fallbackEditors {
			if _, err := exec.LookPath(fallback); err == nil {
				editor = fallback
				break
			}
		}
	}

	if editor == "" {
		return fmt.Errorf(
			"\x1b[31m❌ Critical Error: No text editor found in your system ($PATH).\x1b[0m\n" +
				"Please install nano, helix, or vim, or set a correct path in config.toml.",
		)
	}

	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	return cmd.Run()
}

func GetFile() string {
	noteList := scanVault(GlobalConfig.MainVault)
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
	notes := scanVault(GlobalConfig.MainVault)
	for _, item := range notes {
		if item.FileName == fileName {
			EditFile(item.FilePath)
			return
		}
	}
	fullPath := filepath.Join(GlobalConfig.MainVault, fileName)
	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		fmt.Printf("create error %v\n", err)
		return
	}
	template, err := os.ReadFile(GlobalConfig.TemplateNote)
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
	filePath := filepath.Join(GlobalConfig.PathToInlineNote, fileName)
	err := os.WriteFile(filePath, []byte(inlineText), 0600)
	if err != nil {
		return fmt.Errorf("failed to write inline note: %w", err)
	}
	fmt.Printf("\033[32mNote saved successfully: %s\033[0m\n", fileName)
	return nil
}

// InteractiveSearchInternal ищет по всему vault через rg/grep,
// но фильтрует внутри встроенного go-fuzzyfinder (без внешнего fzf).
func InteractiveSearchInternal() {
	root := GlobalConfig.MainVault
	var cmd *exec.Cmd

	if _, err := exec.LookPath("rg"); err == nil {
		cmd = exec.Command("rg", "--line-number", "--column", "--no-heading", "--smart-case", "", root)
	} else if _, err := exec.LookPath("grep"); err == nil {
		cmd = exec.Command("grep", "-rn", ".", root)
	} else {
		fmt.Println("Ошибка: в системе не найдены ни 'rg', ни 'grep'.")
		return
	}

	// 1. Получаем доступ к потоку вывода команды вместо выгрузки всего в Output()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Ошибка создания пайпа: %v\n", err)
		return
	}

	// Стартуем команду в фоне
	if err := cmd.Start(); err != nil {
		fmt.Printf("Ошибка запуска поиска: %v\n", err)
		return
	}

	// 2. Настраиваем сканер для построчного чтения
	scanner := bufio.NewScanner(stdout)

	// Задаем лимит на буфер строки (на случай, если в заметках есть гигантские строки-минифайлы)
	const maxCapacity = 1024 * 1024 // 1 МБ на строку
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxCapacity)

	var results []SearchLine
	const maxLines = 30000 // ЛИМИТ СТРОК ДЛЯ ОЗУ

	for scanner.Scan() {
		// Если нагребли достаточно строк — жестко тормозим процесс поиска
		if len(results) >= maxLines {
			_ = cmd.Process.Kill() // Убиваем rg/grep, чтобы они не тратили CPU в фоне
			break
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 4)
		if len(parts) < 3 {
			continue
		}

		res := SearchLine{FilePath: parts[0], Line: parts[1]}
		if len(parts) == 4 {
			res.Col = parts[2]
			res.Text = strings.TrimSpace(parts[3])
		} else {
			res.Text = strings.TrimSpace(parts[2])
		}

		results = append(results, res)
	}

	if err := scanner.Err(); err != nil {
		// Важно: если мы сами убили процесс через Kill(), сканер может ругнуться на "read |0: file already closed".
		// Если мы вышли по лимиту строк, нам на эту ошибку плевать, данные-то мы уже собрали.
		if len(results) < maxLines {
			fmt.Printf("Ошибка при сканировании вывода: %v\n", err)
		}
	}
	// Дожидаемся окончательной очистки ресурсов процесса
	_ = cmd.Wait()

	if len(results) == 0 {
		fmt.Println("Ничего не найдено.")
		return
	}

	// 3. Запускаем go-fuzzyfinder
	idx, err := fuzzyfinder.Find(
		results,
		func(i int) string {
			relPath, _ := filepath.Rel(root, results[i].FilePath)
			// Формируем красивую строку для поиска
			if results[i].Col != "" {
				return fmt.Sprintf("%s:%s:%s -> %s", relPath, results[i].Line, results[i].Col, results[i].Text)
			}
			return fmt.Sprintf("%s:%s -> %s", relPath, results[i].Line, results[i].Text)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			// Читаем файл для превью
			contentBytes, err := os.ReadFile(results[i].FilePath)
			if err != nil {
				return fmt.Sprintf("Ошибка чтения файла: %v", err)
			}

			return fmt.Sprintf("=== %s ===\nСтрока: %s, Колонка: %s\n\n%s",
				filepath.Base(results[i].FilePath),
				results[i].Line,
				results[i].Col,
				string(contentBytes),
			)
		}),
	)

	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			return // Пользователь нажал Esc
		}
		log.Printf("Ошибка выбора: %v", err)
		return
	}

	// 4. Формируем таргет в формате file:line:col для Helix / Vim / Neovim
	selected := results[idx]
	var target string
	if selected.Col != "" {
		target = fmt.Sprintf("%s:%s:%s", selected.FilePath, selected.Line, selected.Col)
	} else {
		target = fmt.Sprintf("%s:%s", selected.FilePath, selected.Line)
	}

	// 5. Открываем редактор
	editorCmd := exec.Command(GlobalConfig.Editor, target)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		fmt.Printf("Не удалось запустить редактор: %v\n", err)
	}
}
