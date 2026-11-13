// Клиентское приложение (TUI)
package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	api "github.com/alexsey-popov/gkeep/proto"
	pbText "github.com/alexsey-popov/gkeep/proto/text"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"google.golang.org/grpc/status"
)

// initTextsListView создает экран со списком текстовых записей
func (a *App) initTextsListView() {
	a.textsList = tview.NewList().
		ShowSecondaryText(true)

	openBtn := tview.NewButton("Открыть").
		SetSelectedFunc(func() {
			idx := a.textsList.GetCurrentItem()
			if idx >= 0 && idx < len(a.currentTexts) {
				a.openTextView(a.currentTexts[idx])
			}
		})

	addBtn := tview.NewButton("Добавить").
		SetSelectedFunc(func() {
			a.clearCreateTextFields()
			a.pages.SwitchToPage("text_create")
		})

	a.prevTextPageBtn = tview.NewButton("< Пред.").
		SetSelectedFunc(a.prevTextPage)

	a.nextTextPageBtn = tview.NewButton("След. >").
		SetSelectedFunc(a.nextTextPage)

	refreshBtn := tview.NewButton("Обновить").
		SetSelectedFunc(a.handleRefreshTexts)

	backBtn := tview.NewButton("Назад").
		SetSelectedFunc(func() {
			a.pages.SwitchToPage("data_types")
		})

	btnFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(openBtn, 10, 1, true).
		AddItem(nil, 1, 0, false).
		AddItem(addBtn, 11, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(a.prevTextPageBtn, 10, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(a.nextTextPageBtn, 10, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(refreshBtn, 11, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(backBtn, 8, 1, false)

	a.textsFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.textsList, 0, 1, true).
		AddItem(btnFlex, 3, 0, false)

	a.textsFlex.SetBorder(true).
		SetTitle(a.getTextsListTitle()).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorCadetBlue)
}

// initTextDetailView создает экран просмотра отдельной текстовой записи
func (a *App) initTextDetailView() {
	a.textDetailsText = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWrap(true)

	editBtn := tview.NewButton("Изменить").
		SetSelectedFunc(func() {
			if a.currentText != nil {
				a.openTextEdit(a.currentText)
			}
		})

	deleteBtn := tview.NewButton("Удалить").
		SetSelectedFunc(func() {
			if a.currentText != nil {
				a.showDeleteTextConfirmModal()
			}
		})

	backBtn := tview.NewButton("Назад").
		SetSelectedFunc(func() {
			a.pages.SwitchToPage("text_list")
		})

	btnFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(editBtn, 14, 1, true).
		AddItem(nil, 2, 0, false).
		AddItem(deleteBtn, 14, 1, false).
		AddItem(nil, 2, 0, false).
		AddItem(backBtn, 12, 1, false).
		AddItem(nil, 0, 1, false)

	a.textViewFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.textDetailsText, 0, 1, false).
		AddItem(btnFlex, 3, 0, true)

	a.textViewFlex.SetBorder(true).
		SetTitle(" [::b]Просмотр текста[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkGoldenrod)
}

// initTextCreateForm создает форму добавления текстовой записи
func (a *App) initTextCreateForm() {
	a.createTextNameField = tview.NewInputField().
		SetLabel("Название: ").
		SetFieldWidth(32)

	a.createTextValField = tview.NewInputField().
		SetLabel("Текст: ").
		SetFieldWidth(32)

	a.textCreateForm = tview.NewForm().
		AddFormItem(a.createTextNameField).
		AddFormItem(a.createTextValField).
		AddButton("Сохранить", a.handleTextCreate).
		AddButton("Назад", func() {
			a.clearCreateTextFields()
			a.pages.SwitchToPage("text_list")
		})

	a.textCreateForm.SetBorder(true).
		SetTitle(" [::b]Добавление текста[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorGreen)
}

// initTextEditForm создает форму редактирования текстовой записи
func (a *App) initTextEditForm() {
	a.editTextIDField = tview.NewInputField().
		SetLabel("ID: ").
		SetFieldWidth(32)
	a.editTextIDField.SetDisabled(true)

	a.editTextNameField = tview.NewInputField().
		SetLabel("Название: ").
		SetFieldWidth(32)

	a.editTextValField = tview.NewInputField().
		SetLabel("Текст: ").
		SetFieldWidth(32)

	a.textEditForm = tview.NewForm().
		AddFormItem(a.editTextIDField).
		AddFormItem(a.editTextNameField).
		AddFormItem(a.editTextValField).
		AddButton("Сохранить", a.handleTextUpdate).
		AddButton("Назад", func() {
			a.pages.SwitchToPage("text_view")
		})

	a.textEditForm.SetBorder(true).
		SetTitle(" [::b]Редактирование текста[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorOrange)
}

// openTextList загружает список текстовых записей и переключает на экран списка
func (a *App) openTextList() {
	a.currentTextPage = 1
	a.loadTexts()
	a.pages.SwitchToPage("text_list")
}

// handleRefreshTexts выполняет обновление списка текстовых записей и отображает уведомление
func (a *App) handleRefreshTexts() {
	a.loadTexts()
	if !a.pages.HasPage("error_modal") {
		if a.isTextsCached {
			a.showInfoModal("Данные обновлены из кеша", "text_list")
		} else {
			a.showInfoModal("Данные обновлены", "text_list")
		}
	}
}

// loadTexts выполняет запрос к серверу для получения списка текстовых записей пользователя
func (a *App) loadTexts() {
	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	if a.currentTextPage < 1 {
		a.currentTextPage = 1
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &api.ListRequest{}
	req.SetPage(strconv.Itoa(a.currentTextPage))

	resp, err := a.usersClient.TextList(ctx, req)
	if err != nil {
		// При ошибке сервера пробуем загрузить из кэша
		cached, cacheErr := a.loadCachedTexts(a.currentTextPage)
		if cacheErr == nil && cached != nil {
			a.isTextsCached = true
			a.currentTexts = cached
			a.renderTextsList()
			return
		}

		a.isTextsCached = false
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка получения списка текстов:\n%s", st.Message()), "data_types")
		return
	}

	a.isTextsCached = false
	a.currentTexts = resp.GetText()
	_ = a.saveCachedTexts(a.currentTextPage, a.currentTexts)
	a.renderTextsList()
}

// renderTextsList обновляет визуальный список текстов
func (a *App) renderTextsList() {
	a.textsList.Clear()
	a.textsFlex.SetTitle(a.getTextsListTitle())

	if len(a.currentTexts) == 0 {
		if a.currentTextPage > 1 {
			a.textsList.AddItem("[gray]На этой странице нет записей[-]", "Нажмите кнопку '< Пред.' для возврата", 0, nil)
		} else {
			a.textsList.AddItem("[gray]Нет сохраненных текстов[-]", "Нажмите кнопку 'Добавить' для создания записи", 0, nil)
		}
		return
	}

	for i, item := range a.currentTexts {
		itemCopy := item
		shortcut := rune(0)
		if i < 9 {
			shortcut = rune('1' + i)
		} else if i == 9 {
			shortcut = '0'
		}
		a.textsList.AddItem(
			itemCopy.GetName(),
			"",
			shortcut,
			func() {
				a.openTextView(itemCopy)
			},
		)
	}
}

// openTextView открывает экран подробного просмотра текстовой записи
func (a *App) openTextView(t *pbText.Text) {
	if t == nil {
		return
	}
	a.currentText = t

	cachedMark := ""
	if a.isTextsCached {
		cachedMark = "  [yellow](Локальная копия данных)[-]\n\n"
	}

	info := fmt.Sprintf(
		"\n%s  [yellow]Название:[white] %s\n\n"+
			"  [yellow]Текст:[white]\n  %s\n",
		cachedMark,
		t.GetName(),
		t.GetText(),
	)

	a.textDetailsText.SetText(info)
	a.pages.SwitchToPage("text_view")
}

// openTextEdit открывает экран редактирования текстовой записи
func (a *App) openTextEdit(t *pbText.Text) {
	if t == nil {
		return
	}
	a.currentText = t
	a.editTextIDField.SetText(t.GetId())
	a.editTextNameField.SetText(t.GetName())
	a.editTextValField.SetText(t.GetText())
	a.pages.SwitchToPage("text_edit")
}

// handleTextCreate создает новую текстовую запись
func (a *App) handleTextCreate() {
	name := strings.TrimSpace(a.createTextNameField.GetText())
	textVal := strings.TrimSpace(a.createTextValField.GetText())

	if name == "" {
		a.showErrorModal("Название записи не может быть пустым", "text_create")
		return
	}

	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &pbText.CreateRequest{}
	req.SetName(name)
	req.SetText(textVal)

	_, err := a.usersClient.TextCreate(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка создания записи:\n%s", st.Message()), "text_create")
		return
	}

	a.clearCreateTextFields()
	a.currentTextPage = 1
	a.loadTexts()
	a.pages.SwitchToPage("text_list")
}

// handleTextUpdate обновляет данные выбранной текстовой записи
func (a *App) handleTextUpdate() {
	id := strings.TrimSpace(a.editTextIDField.GetText())
	name := strings.TrimSpace(a.editTextNameField.GetText())
	textVal := strings.TrimSpace(a.editTextValField.GetText())

	if name == "" {
		a.showErrorModal("Название не может быть пустым", "text_edit")
		return
	}

	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &pbText.UpdateRequest{}
	req.SetId(id)
	req.SetName(name)
	req.SetText(textVal)

	resp, err := a.usersClient.TextUpdate(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка обновления записи:\n%s", st.Message()), "text_edit")
		return
	}

	if resp.GetText() != nil {
		a.currentText = resp.GetText()
	} else if a.currentText != nil {
		a.currentText.SetName(name)
		a.currentText.SetText(textVal)
	}

	a.openTextView(a.currentText)
}

// showDeleteTextConfirmModal показывает окно подтверждения удаления текстовой записи
func (a *App) showDeleteTextConfirmModal() {
	if a.currentText == nil {
		return
	}

	name := a.currentText.GetName()
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Вы уверены, что хотите удалить запись '%s'?", name)).
		AddButtons([]string{"Удалить", "Отмена"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("confirm_delete_text_modal")
			if buttonLabel == "Удалить" {
				a.handleTextDelete()
			} else {
				a.pages.SwitchToPage("text_view")
			}
		})

	a.pages.AddPage("confirm_delete_text_modal", modal, true, true)
}

// handleTextDelete отправляет запрос на удаление текстовой записи
func (a *App) handleTextDelete() {
	if a.currentText == nil {
		return
	}

	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &api.IDRequest{}
	req.SetId(a.currentText.GetId())

	_, err := a.usersClient.TextDelete(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка удаления записи:\n%s", st.Message()), "text_view")
		return
	}

	a.currentText = nil
	a.loadTexts()
	if len(a.currentTexts) == 0 && a.currentTextPage > 1 {
		a.currentTextPage--
		a.loadTexts()
	}
	a.pages.SwitchToPage("text_list")
}

// prevTextPage переключает на предыдущую страницу списка текстов
func (a *App) prevTextPage() {
	if a.currentTextPage > 1 {
		a.currentTextPage--
		a.loadTexts()
	}
}

// nextTextPage переключает на следующую страницу списка текстов
func (a *App) nextTextPage() {
	if len(a.currentTexts) > 0 {
		a.currentTextPage++
		a.loadTexts()
	}
}

// GetCurrentTextPage возвращает номер текущей страницы списка текстов
func (a *App) GetCurrentTextPage() int {
	return a.currentTextPage
}

// SetCurrentTextPage устанавливает номер текущей страницы списка текстов
func (a *App) SetCurrentTextPage(p int) {
	if p < 1 {
		p = 1
	}
	a.currentTextPage = p
}

// getTextsListTitle возвращает заголовок для списка записей
func (a *App) getTextsListTitle() string {
	cachedMark := ""
	if a.isTextsCached {
		cachedMark = " [yellow][Локальная копия данных][-]"
	}
	if a.currentTextPage <= 1 {
		return fmt.Sprintf(" [::b]Список текстовых данных (Стр. 1)%s[::-] ", cachedMark)
	}
	return fmt.Sprintf(" [::b]Список текстовых данных (Стр. %d)%s[::-] ", a.currentTextPage, cachedMark)
}

// clearCreateTextFields очищает поля формы создания текстовой записи
func (a *App) clearCreateTextFields() {
	a.createTextNameField.SetText("")
	a.createTextValField.SetText("")
}
