// Клиентское приложение (TUI)
package client

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	api "github.com/alexsey-popov/gkeep/proto"
	pbFile "github.com/alexsey-popov/gkeep/proto/file"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"google.golang.org/grpc/status"
)

// initFilesListView создает экран со списком файлов
func (a *App) initFilesListView() {
	a.filesList = tview.NewList().
		ShowSecondaryText(true)

	openBtn := tview.NewButton("Открыть").
		SetSelectedFunc(func() {
			idx := a.filesList.GetCurrentItem()
			if idx >= 0 && idx < len(a.currentFiles) {
				a.openFileView(a.currentFiles[idx])
			}
		})

	uploadBtn := tview.NewButton("Загрузить").
		SetSelectedFunc(func() {
			a.clearUploadFileFields()
			a.pages.SwitchToPage("file_upload")
		})

	a.prevFilePageBtn = tview.NewButton("< Пред.").
		SetSelectedFunc(a.prevFilePage)

	a.nextFilePageBtn = tview.NewButton("След. >").
		SetSelectedFunc(a.nextFilePage)

	refreshBtn := tview.NewButton("Обновить").
		SetSelectedFunc(func() {
			a.loadFiles()
		})

	backBtn := tview.NewButton("Назад").
		SetSelectedFunc(func() {
			a.pages.SwitchToPage("data_types")
		})

	btnFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(openBtn, 10, 1, true).
		AddItem(nil, 1, 0, false).
		AddItem(uploadBtn, 12, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(a.prevFilePageBtn, 10, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(a.nextFilePageBtn, 10, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(refreshBtn, 11, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(backBtn, 8, 1, false)

	a.filesFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.filesList, 0, 1, true).
		AddItem(btnFlex, 3, 0, false)

	a.filesFlex.SetBorder(true).
		SetTitle(a.getFilesListTitle()).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkCyan)
}

// initFileDetailView создает экран просмотра метаданных файла
func (a *App) initFileDetailView() {
	a.fileDetailsText = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWrap(true)

	downloadBtn := tview.NewButton("Скачать").
		SetSelectedFunc(func() {
			if a.currentFile != nil {
				a.openFileDownload(a.currentFile)
			}
		})

	deleteBtn := tview.NewButton("Удалить").
		SetSelectedFunc(func() {
			if a.currentFile != nil {
				a.showDeleteFileConfirmModal()
			}
		})

	backBtn := tview.NewButton("Назад").
		SetSelectedFunc(func() {
			a.pages.SwitchToPage("file_list")
		})

	btnFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(downloadBtn, 14, 1, true).
		AddItem(nil, 2, 0, false).
		AddItem(deleteBtn, 14, 1, false).
		AddItem(nil, 2, 0, false).
		AddItem(backBtn, 12, 1, false).
		AddItem(nil, 0, 1, false)

	a.fileViewFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.fileDetailsText, 0, 1, false).
		AddItem(btnFlex, 3, 0, true)

	a.fileViewFlex.SetBorder(true).
		SetTitle(" [::b]Просмотр файла[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkCyan)
}

// initFileUploadForm создает форму загрузки файла
func (a *App) initFileUploadForm() {
	a.uploadFileNameField = tview.NewInputField().
		SetLabel("Название: ").
		SetFieldWidth(35)

	a.uploadFilePathField = tview.NewInputField().
		SetLabel("Путь к файлу: ").
		SetFieldWidth(35)

	a.fileUploadForm = tview.NewForm().
		AddFormItem(a.uploadFileNameField).
		AddFormItem(a.uploadFilePathField).
		AddButton("Загрузить", a.handleFileUpload).
		AddButton("Назад", func() {
			a.clearUploadFileFields()
			a.pages.SwitchToPage("file_list")
		})

	a.fileUploadForm.SetBorder(true).
		SetTitle(" [::b]Загрузка файла[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkCyan)
}

// initFileDownloadForm создает форму скачивания файла
func (a *App) initFileDownloadForm() {
	a.downloadSavePathField = tview.NewInputField().
		SetLabel("Сохранить в: ").
		SetFieldWidth(35)

	a.fileDownloadForm = tview.NewForm().
		AddFormItem(a.downloadSavePathField).
		AddButton("Скачать", a.handleFileDownload).
		AddButton("Назад", func() {
			a.clearDownloadFileFields()
			a.pages.SwitchToPage("file_view")
		})

	a.fileDownloadForm.SetBorder(true).
		SetTitle(" [::b]Скачивание файла[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkCyan)
}

// openFileList загружает список файлов и переключает на экран списка
func (a *App) openFileList() {
	a.currentFilePage = 1
	a.loadFiles()
	a.pages.SwitchToPage("file_list")
}

// loadFiles выполняет запрос к серверу для получения списка файлов пользователя
func (a *App) loadFiles() {
	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	if a.currentFilePage < 1 {
		a.currentFilePage = 1
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &api.ListRequest{}
	req.SetPage(strconv.Itoa(a.currentFilePage))

	resp, err := a.usersClient.FileList(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка получения списка файлов:\n%s", st.Message()), "data_types")
		return
	}

	a.currentFiles = resp.GetFiles()
	a.filesList.Clear()
	a.filesFlex.SetTitle(a.getFilesListTitle())

	if len(a.currentFiles) == 0 {
		if a.currentFilePage > 1 {
			a.filesList.AddItem("[gray]На этой странице нет записей[-]", "Нажмите кнопку '< Пред.' для возврата", 0, nil)
		} else {
			a.filesList.AddItem("[gray]Нет сохраненных файлов[-]", "Нажмите кнопку 'Загрузить' для добавления файла", 0, nil)
		}
		return
	}

	for i, item := range a.currentFiles {
		itemCopy := item
		shortcut := rune(0)
		if i < 9 {
			shortcut = rune('1' + i)
		} else if i == 9 {
			shortcut = '0'
		}
		secondary := fmt.Sprintf("%s (%s)", itemCopy.GetFilename(), formatFileSize(itemCopy.GetSize()))
		a.filesList.AddItem(
			itemCopy.GetName(),
			secondary,
			shortcut,
			func() {
				a.openFileView(itemCopy)
			},
		)
	}
}

// openFileView открывает экран подробного просмотра метаданных файла
func (a *App) openFileView(f *pbFile.File) {
	if f == nil {
		return
	}
	a.currentFile = f

	info := fmt.Sprintf(
		"\n  [yellow]Название:[white]          %s\n\n"+
			"  [yellow]Имя файла:[white]         %s\n\n"+
			"  [yellow]Размер:[white]            %s\n\n"+
			"  [yellow]Контрольная сумма:[white] %s\n",
		f.GetName(),
		f.GetFilename(),
		formatFileSize(f.GetSize()),
		f.GetChecksum(),
	)

	a.fileDetailsText.SetText(info)
	a.pages.SwitchToPage("file_view")
}

// openFileDownload открывает форму скачивания файла с предустановленным именем
func (a *App) openFileDownload(f *pbFile.File) {
	if f == nil {
		return
	}
	a.currentFile = f
	defaultPath := f.GetFilename()
	if defaultPath == "" {
		defaultPath = "downloaded_file"
	}
	a.downloadSavePathField.SetText(defaultPath)
	a.pages.SwitchToPage("file_download")
}

// handleFileUpload выполняет потоковую загрузку файла по чанкам на сервер
func (a *App) handleFileUpload() {
	name := strings.TrimSpace(a.uploadFileNameField.GetText())
	filePath := strings.TrimSpace(a.uploadFilePathField.GetText())

	if name == "" {
		a.showErrorModal("Название записи не может быть пустым", "file_upload")
		return
	}
	if filePath == "" {
		a.showErrorModal("Путь к файлу не может быть пустым", "file_upload")
		return
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		a.showErrorModal(fmt.Sprintf("Не удалось открыть файл: %v", err), "file_upload")
		return
	}
	if fileInfo.IsDir() {
		a.showErrorModal("Указанный путь является директорией, а не файлом", "file_upload")
		return
	}
	if fileInfo.Size() > 100*1024*1024 {
		a.showErrorModal("Размер файла превышает максимально допустимый лимит (100 МБ)", "file_upload")
		return
	}

	fileHandle, err := os.Open(filePath)
	if err != nil {
		a.showErrorModal(fmt.Sprintf("Ошибка чтения файла: %v", err), "file_upload")
		return
	}
	defer fileHandle.Close()

	// Вычисляем SHA-256 хеш файла
	hasher := sha256.New()
	if _, err := io.Copy(hasher, fileHandle); err != nil {
		a.showErrorModal(fmt.Sprintf("Ошибка вычисления контрольной суммы: %v", err), "file_upload")
		return
	}
	checksum := hex.EncodeToString(hasher.Sum(nil))

	// Возвращаем указатель чтения в начало файла
	if _, err := fileHandle.Seek(0, io.SeekStart); err != nil {
		a.showErrorModal(fmt.Sprintf("Ошибка перемотки файла: %v", err), "file_upload")
		return
	}

	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	stream, err := a.usersClient.FileUpload(ctx)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка создания потока загрузки:\n%s", st.Message()), "file_upload")
		return
	}

	// 1. Отправляем метаданные первым сообщением
	baseFileName := filepath.Base(filePath)
	fileSize := fileInfo.Size()

	metaReq := &pbFile.UploadFileRequest{}
	metaReq.SetMetadata(pbFile.FileMetadata_builder{
		Name:     &name,
		Filename: &baseFileName,
		Size:     &fileSize,
		Checksum: &checksum,
	}.Build())

	if err := stream.Send(metaReq); err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка отправки метаданных файла:\n%s", st.Message()), "file_upload")
		return
	}

	// 2. Отправляем бинарные чанки (по 64 КБ)
	buf := make([]byte, 64*1024)
	for {
		n, err := fileHandle.Read(buf)
		if n > 0 {
			chunkReq := &pbFile.UploadFileRequest{}
			chunkReq.SetChunk(buf[:n])
			if sendErr := stream.Send(chunkReq); sendErr != nil {
				// Если поток был закрыт сервером, вызываем CloseAndRecv(), чтобы получить реальный gRPC-статус ошибки
				_, recvErr := stream.CloseAndRecv()
				if recvErr != nil {
					st := status.Convert(recvErr)
					a.showErrorModal(fmt.Sprintf("Ошибка загрузки файла на сервере:\n%s", st.Message()), "file_upload")
				} else {
					st := status.Convert(sendErr)
					a.showErrorModal(fmt.Sprintf("Ошибка отправки чанка файла:\n%s", st.Message()), "file_upload")
				}
				return
			}
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			a.showErrorModal(fmt.Sprintf("Ошибка чтения локального файла:\n%v", err), "file_upload")
			return
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка завершения загрузки файла:\n%s", st.Message()), "file_upload")
		return
	}

	if resp.GetError() != "" {
		a.showErrorModal(fmt.Sprintf("Сервер вернул ошибку:\n%s", resp.GetError()), "file_upload")
		return
	}

	a.clearUploadFileFields()
	a.currentFilePage = 1
	a.loadFiles()
	a.pages.SwitchToPage("file_list")
}

// handleFileDownload выполняет потоковое скачивание файла по чанкам с сервера
func (a *App) handleFileDownload() {
	if a.currentFile == nil {
		a.pages.SwitchToPage("file_list")
		return
	}

	savePath := strings.TrimSpace(a.downloadSavePathField.GetText())
	if savePath == "" {
		a.showErrorModal("Укажите путь для сохранения файла", "file_download")
		return
	}

	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	// Создаем родительскую директорию, если она указана и не существует
	if dir := filepath.Dir(savePath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			a.showErrorModal(fmt.Sprintf("Ошибка создания директории для сохранения:\n%v", err), "file_download")
			return
		}
	}

	outFile, err := os.Create(savePath)
	if err != nil {
		a.showErrorModal(fmt.Sprintf("Ошибка создания целевого файла:\n%v", err), "file_download")
		return
	}

	var downloadSuccess bool
	defer func() {
		outFile.Close()
		if !downloadSuccess {
			_ = os.Remove(savePath)
		}
	}()

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	fileID := a.currentFile.GetId()
	stream, err := a.usersClient.FileDownload(ctx, pbFile.DownloadFileRequest_builder{
		Id: &fileID,
	}.Build())
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка вызова скачивания файла:\n%s", st.Message()), "file_download")
		return
	}

	// Читаем первое сообщение с метаданными
	firstMsg, err := stream.Recv()
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка получения метаданных скачиваемого файла:\n%s", st.Message()), "file_download")
		return
	}

	meta := firstMsg.GetMetadata()
	if meta == nil {
		a.showErrorModal("Сервер не вернул метаданные файла", "file_download")
		return
	}

	hasher := sha256.New()
	var totalBytes int64

	// Читаем поток чанков
	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			st := status.Convert(err)
			a.showErrorModal(fmt.Sprintf("Обрыв соединения при скачивании чанков:\n%s", st.Message()), "file_download")
			return
		}

		chunk := msg.GetChunk()
		if len(chunk) > 0 {
			if _, err := outFile.Write(chunk); err != nil {
				a.showErrorModal(fmt.Sprintf("Ошибка записи чанка на диск:\n%v", err), "file_download")
				return
			}
			hasher.Write(chunk)
			totalBytes += int64(len(chunk))
		}
	}

	// Проверяем контрольную сумму при скачивании
	calcChecksum := hex.EncodeToString(hasher.Sum(nil))
	if meta.GetChecksum() != "" && meta.GetChecksum() != calcChecksum {
		a.showErrorModal("Несовпадение контрольной суммы скачанного файла", "file_download")
		return
	}

	downloadSuccess = true
	a.clearDownloadFileFields()
	a.showInfoModal(fmt.Sprintf("Файл успешно скачан и сохранен:\n%s (%s)", savePath, formatFileSize(totalBytes)), "file_view")
}

// showDeleteFileConfirmModal показывает окно подтверждения удаления файла
func (a *App) showDeleteFileConfirmModal() {
	if a.currentFile == nil {
		return
	}

	name := a.currentFile.GetName()
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Вы уверены, что хотите удалить файл '%s'?", name)).
		AddButtons([]string{"Удалить", "Отмена"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("confirm_delete_file_modal")
			if buttonLabel == "Удалить" {
				a.handleFileDelete()
			} else {
				a.pages.SwitchToPage("file_view")
			}
		})

	a.pages.AddPage("confirm_delete_file_modal", modal, true, true)
}

// handleFileDelete отправляет запрос на удаление файла
func (a *App) handleFileDelete() {
	if a.currentFile == nil {
		return
	}

	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &api.IDRequest{}
	req.SetId(a.currentFile.GetId())

	_, err := a.usersClient.FileDelete(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка удаления файла:\n%s", st.Message()), "file_view")
		return
	}

	a.currentFile = nil
	a.loadFiles()
	if len(a.currentFiles) == 0 && a.currentFilePage > 1 {
		a.currentFilePage--
		a.loadFiles()
	}
	a.pages.SwitchToPage("file_list")
}

// prevFilePage переключает на предыдущую страницу списка файлов
func (a *App) prevFilePage() {
	if a.currentFilePage > 1 {
		a.currentFilePage--
		a.loadFiles()
	}
}

// nextFilePage переключает на следующую страницу списка файлов
func (a *App) nextFilePage() {
	if len(a.currentFiles) > 0 {
		a.currentFilePage++
		a.loadFiles()
	}
}

// GetCurrentFilePage возвращает номер текущей страницы списка файлов
func (a *App) GetCurrentFilePage() int {
	return a.currentFilePage
}

// SetCurrentFilePage устанавливает номер текущей страницы списка файлов
func (a *App) SetCurrentFilePage(p int) {
	if p < 1 {
		p = 1
	}
	a.currentFilePage = p
}

// getTextsListTitle возвращает заголовок для списка записей
func (a *App) getFilesListTitle() string {
	if a.currentFilePage <= 1 {
		return " [::b]Список файлов (Стр. 1)[::-] "
	}
	return fmt.Sprintf(" [::b]Список файлов (Стр. %d)[::-] ", a.currentFilePage)
}

// clearUploadFileFields очищает поля формы загрузки файла
func (a *App) clearUploadFileFields() {
	a.uploadFileNameField.SetText("")
	a.uploadFilePathField.SetText("")
}

// clearDownloadFileFields очищает поля формы скачивания файла
func (a *App) clearDownloadFileFields() {
	a.downloadSavePathField.SetText("")
}

// formatFileSize форматирует размер файла в удобочитаемый вид
func formatFileSize(bytes int64) string {
	const (
		_KB = 1024
		_MB = _KB * 1024
		_GB = _MB * 1024
	)
	switch {
	case bytes >= _GB:
		return fmt.Sprintf("%.2f GB (%d байт)", float64(bytes)/float64(_GB), bytes)
	case bytes >= _MB:
		return fmt.Sprintf("%.2f MB (%d байт)", float64(bytes)/float64(_MB), bytes)
	case bytes >= _KB:
		return fmt.Sprintf("%.2f KB (%d байт)", float64(bytes)/float64(_KB), bytes)
	default:
		return fmt.Sprintf("%d байт", bytes)
	}
}
