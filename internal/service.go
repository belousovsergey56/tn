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

func EditFile(filePath, line, col string) error {
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
			"\x1b[31mCritical Error: No text editor found in your system ($PATH).\x1b[0m\n" +
				"Please install nano, helix, or vim, or set a correct path in config.toml.",
		)
	}

	var args []string
	baseEditor := strings.ToLower(filepath.Base(editor))
	baseEditor = strings.TrimSuffix(baseEditor, ".exe")
	switch baseEditor {
	case "vim", "nvim", "vi":
		if line != "" {
			if col != "" {
				args = append(args, fmt.Sprintf("+call cursor(%s,%s)", line, col), filePath)
			} else {
				args = append(args, "+"+line, filePath)
			}
		} else {
			args = append(args, filePath)
		}

	case "hx", "helix":
		if col != "" {
			args = append(args, fmt.Sprintf("%s:%s:%s", filePath, line, col))
		} else if line != "" {
			args = append(args, fmt.Sprintf("%s:%s", filePath, line))
		} else {
			args = append(args, filePath)
		}

	case "nano":
		if line != "" {
			if col != "" {
				args = append(args, fmt.Sprintf("+%s,%s", line, col), filePath)
			} else {
				args = append(args, "+"+line, filePath)
			}
		} else {
			args = append(args, filePath)
		}

	default:
		args = append(args, filePath)
	}

	fmt.Print("\033[H\033[2J")

	cmd := exec.Command(editor, args...)
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
			EditFile(item.FilePath, "", "")
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
		EditFile(fullPath, "", "")
		return
	}
	_, err = file.Write(template)
	if err != nil {
		fmt.Printf("write tmp in new file error %v", err)
	}
	file.Close()
	EditFile(fullPath, "", "")
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

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Ошибка создания пайпа: %v\n", err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Ошибка запуска поиска: %v\n", err)
		return
	}

	scanner := bufio.NewScanner(stdout)

	const maxCapacity = 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxCapacity)

	var results []SearchLine
	const maxLines = 30000

	for scanner.Scan() {
		if len(results) >= maxLines {
			_ = cmd.Process.Kill()
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
		if len(results) < maxLines {
			fmt.Printf("Ошибка при сканировании вывода: %v\n", err)
		}
	}
	_ = cmd.Wait()

	if len(results) == 0 {
		fmt.Println("Ничего не найдено.")
		return
	}

	idx, err := fuzzyfinder.Find(
		results,
		func(i int) string {
			relPath, _ := filepath.Rel(root, results[i].FilePath)
			if results[i].Col != "" {
				return fmt.Sprintf("%s:%s:%s -> %s", relPath, results[i].Line, results[i].Col, results[i].Text)
			}
			return fmt.Sprintf("%s:%s -> %s", relPath, results[i].Line, results[i].Text)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
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
			return
		}
		log.Printf("Ошибка выбора: %v", err)
		return
	}

	selected := results[idx]
	if err := EditFile(selected.FilePath, selected.Line, selected.Col); err != nil {
		fmt.Printf("Не удалось запустить редактор: %v\n", err)
	}
}
