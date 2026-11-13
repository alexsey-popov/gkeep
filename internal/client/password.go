// Клиентское приложение (TUI)
package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	api "github.com/alexsey-popov/gkeep/proto"
	pbPassword "github.com/alexsey-popov/gkeep/proto/password"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"google.golang.org/grpc/status"
)

// initPasswordsListView создает экран со списком записей паролей
func (a *App) initPasswordsListView() {
	a.passwordsList = tview.NewList().
		ShowSecondaryText(true)

	openBtn := tview.NewButton("Открыть").
		SetSelectedFunc(func() {
			idx := a.passwordsList.GetCurrentItem()
			if idx >= 0 && idx < len(a.currentPasswords) {
				a.openPasswordView(a.currentPasswords[idx])
			}
		})

	addBtn := tview.NewButton("Добавить").
		SetSelectedFunc(func() {
			a.clearCreatePasswordFields()
			a.pages.SwitchToPage("password_create")
		})

	a.prevPageBtn = tview.NewButton("< Пред.").
		SetSelectedFunc(a.prevPage)

	a.nextPageBtn = tview.NewButton("След. >").
		SetSelectedFunc(a.nextPage)

	refreshBtn := tview.NewButton("Обновить").
		SetSelectedFunc(a.handleRefreshPasswords)

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
		AddItem(a.prevPageBtn, 10, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(a.nextPageBtn, 10, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(refreshBtn, 11, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(backBtn, 8, 1, false)

	a.passwordsFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.passwordsList, 0, 1, true).
		AddItem(btnFlex, 3, 0, false)

	a.passwordsFlex.SetBorder(true).
		SetTitle(a.getPasswordsListTitle()).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorCadetBlue)
}

// initPasswordDetailView создает экран просмотра отдельной записи пароля
func (a *App) initPasswordDetailView() {
	a.passwordDetailsText = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWrap(true)

	editBtn := tview.NewButton("Изменить").
		SetSelectedFunc(func() {
			if a.currentPassword != nil {
				a.openPasswordEdit(a.currentPassword)
			}
		})

	deleteBtn := tview.NewButton("Удалить").
		SetSelectedFunc(func() {
			if a.currentPassword != nil {
				a.showDeletePasswordConfirmModal()
			}
		})

	backBtn := tview.NewButton("Назад").
		SetSelectedFunc(func() {
			a.pages.SwitchToPage("password_list")
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

	a.passwordViewFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.passwordDetailsText, 0, 1, false).
		AddItem(btnFlex, 3, 0, true)

	a.passwordViewFlex.SetBorder(true).
		SetTitle(" [::b]Просмотр пароля[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkGoldenrod)
}

// initPasswordCreateForm создает форму добавления пароля
func (a *App) initPasswordCreateForm() {
	a.createPassNameField = tview.NewInputField().
		SetLabel("Название: ").
		SetFieldWidth(32)

	a.createPassLoginField = tview.NewInputField().
		SetLabel("Логин: ").
		SetFieldWidth(32)

	a.createPassValField = tview.NewInputField().
		SetLabel("Пароль: ").
		SetFieldWidth(32)

	a.passwordCreateForm = tview.NewForm().
		AddFormItem(a.createPassNameField).
		AddFormItem(a.createPassLoginField).
		AddFormItem(a.createPassValField).
		AddButton("Сохранить", a.handlePasswordCreate).
		AddButton("Назад", func() {
			a.clearCreatePasswordFields()
			a.pages.SwitchToPage("password_list")
		})

	a.passwordCreateForm.SetBorder(true).
		SetTitle(" [::b]Добавление пароля[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorGreen)
}

// initPasswordEditForm создает форму редактирования пароля
func (a *App) initPasswordEditForm() {
	a.editPassIDField = tview.NewInputField().
		SetLabel("ID: ").
		SetFieldWidth(32)
	a.editPassIDField.SetDisabled(true)

	a.editPassNameField = tview.NewInputField().
		SetLabel("Название: ").
		SetFieldWidth(32)

	a.editPassLoginField = tview.NewInputField().
		SetLabel("Логин: ").
		SetFieldWidth(32)

	a.editPassValField = tview.NewInputField().
		SetLabel("Новый пароль: ").
		SetFieldWidth(32)

	a.passwordEditForm = tview.NewForm().
		AddFormItem(a.editPassIDField).
		AddFormItem(a.editPassNameField).
		AddFormItem(a.editPassLoginField).
		AddFormItem(a.editPassValField).
		AddButton("Сохранить", a.handlePasswordUpdate).
		AddButton("Назад", func() {
			a.pages.SwitchToPage("password_view")
		})

	a.passwordEditForm.SetBorder(true).
		SetTitle(" [::b]Редактирование пароля[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorOrange)
}

// openPasswordList загружает список паролей и переключает на экран списка
func (a *App) openPasswordList() {
	a.currentPage = 1
	a.loadPasswords()
	a.pages.SwitchToPage("password_list")
}

// handleRefreshPasswords выполняет обновление списка паролей и отображает уведомление
func (a *App) handleRefreshPasswords() {
	a.loadPasswords()
	if !a.pages.HasPage("error_modal") {
		if a.isPasswordsCached {
			a.showInfoModal("Данные обновлены из кеша", "password_list")
		} else {
			a.showInfoModal("Данные обновлены", "password_list")
		}
	}
}

// loadPasswords выполняет запрос к серверу для получения списка паролей пользователя
func (a *App) loadPasswords() {
	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	if a.currentPage < 1 {
		a.currentPage = 1
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &api.ListRequest{}
	req.SetPage(strconv.Itoa(a.currentPage))

	resp, err := a.usersClient.PasswordList(ctx, req)
	if err != nil {
		// При ошибке сервера пробуем загрузить из кэша
		cached, cacheErr := a.loadCachedPasswords(a.currentPage)
		if cacheErr == nil && cached != nil {
			a.isPasswordsCached = true
			a.currentPasswords = cached
			a.renderPasswordsList()
			return
		}

		a.isPasswordsCached = false
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка получения списка паролей:\n%s", st.Message()), "data_types")
		return
	}

	a.isPasswordsCached = false
	a.currentPasswords = resp.GetPassword()
	_ = a.saveCachedPasswords(a.currentPage, a.currentPasswords)
	a.renderPasswordsList()
}

// renderPasswordsList обновляет визуальный список паролей
func (a *App) renderPasswordsList() {
	a.passwordsList.Clear()
	a.passwordsFlex.SetTitle(a.getPasswordsListTitle())

	if len(a.currentPasswords) == 0 {
		if a.currentPage > 1 {
			a.passwordsList.AddItem("[gray]На этой странице нет записей[-]", "Нажмите кнопку '< Пред.' для возврата", 0, nil)
		} else {
			a.passwordsList.AddItem("[gray]Нет сохраненных паролей[-]", "Нажмите кнопку 'Добавить' для создания записи", 0, nil)
		}
		return
	}

	for i, item := range a.currentPasswords {
		itemCopy := item
		shortcut := rune(0)
		if i < 9 {
			shortcut = rune('1' + i)
		} else if i == 9 {
			shortcut = '0'
		}
		a.passwordsList.AddItem(
			itemCopy.GetName(),
			"",
			shortcut,
			func() {
				a.openPasswordView(itemCopy)
			},
		)
	}
}

// openPasswordView открывает экран подробного просмотра записи пароля
func (a *App) openPasswordView(p *pbPassword.Password) {
	if p == nil {
		return
	}
	a.currentPassword = p

	cachedMark := ""
	if a.isPasswordsCached {
		cachedMark = "  [yellow](Локальная копия данных)[-]\n\n"
	}

	info := fmt.Sprintf(
		"\n%s  [yellow]Название:[white] %s\n\n"+
			"  [yellow]Логин:[white]    %s\n\n"+
			"  [yellow]Пароль:[white]   %s\n",
		cachedMark,
		p.GetName(),
		p.GetLogin(),
		p.GetPassword(),
	)

	a.passwordDetailsText.SetText(info)
	a.pages.SwitchToPage("password_view")
}

// openPasswordEdit открывает экран редактирования пароля
func (a *App) openPasswordEdit(p *pbPassword.Password) {
	if p == nil {
		return
	}
	a.currentPassword = p
	a.editPassIDField.SetText(p.GetId())
	a.editPassNameField.SetText(p.GetName())
	a.editPassLoginField.SetText(p.GetLogin())
	a.editPassValField.SetText(p.GetPassword())
	a.pages.SwitchToPage("password_edit")
}

// handlePasswordCreate создает новую запись пароля
func (a *App) handlePasswordCreate() {
	name := strings.TrimSpace(a.createPassNameField.GetText())
	login := strings.TrimSpace(a.createPassLoginField.GetText())
	passwordVal := a.createPassValField.GetText()

	if name == "" {
		a.showErrorModal("Название записи не может быть пустым", "password_create")
		return
	}

	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &pbPassword.CreateRequest{}
	req.SetName(name)
	req.SetLogin(login)
	req.SetPassword(passwordVal)

	_, err := a.usersClient.PasswordCreate(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка создания записи:\n%s", st.Message()), "password_create")
		return
	}

	a.clearCreatePasswordFields()
	a.currentPage = 1
	a.loadPasswords()
	a.pages.SwitchToPage("password_list")
}

// handlePasswordUpdate обновляет пароль выбранной записи
func (a *App) handlePasswordUpdate() {
	ID := strings.TrimSpace(a.editPassIDField.GetText())
	name := strings.TrimSpace(a.editPassNameField.GetText())
	login := strings.TrimSpace(a.editPassLoginField.GetText())
	passwordVal := a.editPassValField.GetText()

	if login == "" {
		a.showErrorModal("Логин не может быть пустым", "password_edit")
		return
	}

	if name == "" {
		a.showErrorModal("Название не может быть пустым", "password_edit")
		return
	}

	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &pbPassword.UpdateRequest{}
	req.SetId(ID)
	req.SetName(name)
	req.SetLogin(login)
	req.SetPassword(passwordVal)

	resp, err := a.usersClient.PasswordUpdate(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка обновления записи:\n%s", st.Message()), "password_edit")
		return
	}

	if resp.GetPassword() != nil {
		a.currentPassword = resp.GetPassword()
	} else if a.currentPassword != nil {
		a.currentPassword.SetLogin(login)
		a.currentPassword.SetPassword(passwordVal)
	}

	a.openPasswordView(a.currentPassword)
}

// showDeletePasswordConfirmModal показывает окно подтверждения удаления записи
func (a *App) showDeletePasswordConfirmModal() {
	if a.currentPassword == nil {
		return
	}

	name := a.currentPassword.GetName()
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Вы уверены, что хотите удалить запись '%s'?", name)).
		AddButtons([]string{"Удалить", "Отмена"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("confirm_delete_modal")
			if buttonLabel == "Удалить" {
				a.handlePasswordDelete()
			} else {
				a.pages.SwitchToPage("password_view")
			}
		})

	a.pages.AddPage("confirm_delete_modal", modal, true, true)
}

// handlePasswordDelete отправляет запрос на удаление записи
func (a *App) handlePasswordDelete() {
	if a.currentPassword == nil {
		return
	}

	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &api.IDRequest{}
	req.SetId(a.currentPassword.GetId())

	_, err := a.usersClient.PasswordDelete(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка удаления записи:\n%s", st.Message()), "password_view")
		return
	}

	a.currentPassword = nil
	a.loadPasswords()
	if len(a.currentPasswords) == 0 && a.currentPage > 1 {
		a.currentPage--
		a.loadPasswords()
	}
	a.pages.SwitchToPage("password_list")
}

// prevPage переключает на предыдущую страницу списка паролей
func (a *App) prevPage() {
	if a.currentPage > 1 {
		a.currentPage--
		a.loadPasswords()
	}
}

// nextPage переключает на следующую страницу списка паролей
func (a *App) nextPage() {
	if len(a.currentPasswords) > 0 {
		a.currentPage++
		a.loadPasswords()
	}
}

// GetCurrentPage возвращает номер текущей страницы
func (a *App) GetCurrentPage() int {
	return a.currentPage
}

// SetCurrentPage устанавливает номер текущей страницы
func (a *App) SetCurrentPage(p int) {
	if p < 1 {
		p = 1
	}
	a.currentPage = p
}

// getPasswordsListTitle возвращает заголовок для списка паролей
func (a *App) getPasswordsListTitle() string {
	cachedMark := ""
	if a.isPasswordsCached {
		cachedMark = " [yellow][Локальная копия данных][-]"
	}
	if a.currentPage <= 1 {
		return fmt.Sprintf(" [::b]Список паролей (Стр. 1)%s[::-] ", cachedMark)
	}
	return fmt.Sprintf(" [::b]Список паролей (Стр. %d)%s[::-] ", a.currentPage, cachedMark)
}

// clearCreatePasswordFields очищает поля формы создания пароля
func (a *App) clearCreatePasswordFields() {
	a.createPassNameField.SetText("")
	a.createPassLoginField.SetText("")
	a.createPassValField.SetText("")
}
