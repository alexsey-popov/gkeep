package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "github.com/alexsey-popov/gkeep/proto/user"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// initLoginForm создает форму аутентификации
func (a *App) initLoginForm() {
	a.loginUserField = tview.NewInputField().
		SetLabel("Логин: ").
		SetFieldWidth(30)

	a.loginPassField = tview.NewInputField().
		SetLabel("Пароль: ").
		SetFieldWidth(30).
		SetMaskCharacter('*')

	a.loginForm = tview.NewForm().
		AddFormItem(a.loginUserField).
		AddFormItem(a.loginPassField).
		AddButton("Войти", a.handleLogin).
		AddButton("Регистрация", func() {
			a.clearLoginFields()
			a.pages.SwitchToPage("register")
		}).
		AddButton("Закрыть приложение", func() {
			a.tviewApp.Stop()
		})

	a.loginForm.SetBorder(true).
		SetTitle(a.getLoginTitle()).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorCadetBlue)
}

// initRegisterForm создает форму регистрации нового пользователя
func (a *App) initRegisterForm() {
	a.regUserField = tview.NewInputField().
		SetLabel("Логин: ").
		SetFieldWidth(30)

	a.regPassField = tview.NewInputField().
		SetLabel("Пароль: ").
		SetFieldWidth(30).
		SetMaskCharacter('*')

	a.regPassField2 = tview.NewInputField().
		SetLabel("Повтор пароля: ").
		SetFieldWidth(30).
		SetMaskCharacter('*')

	a.registerForm = tview.NewForm().
		AddFormItem(a.regUserField).
		AddFormItem(a.regPassField).
		AddFormItem(a.regPassField2).
		AddButton("Зарегистрироваться", a.handleRegister).
		AddButton("Назад к входу", func() {
			a.clearRegisterFields()
			a.pages.SwitchToPage("login")
		}).
		AddButton("Выход", func() {
			a.tviewApp.Stop()
		})

	a.registerForm.SetBorder(true).
		SetTitle(a.getRegisterTitle()).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkCyan)
}

// initDataTypesView создает экран выбора типа данных
func (a *App) initDataTypesView() {
	a.dataTypesUserText = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	a.dataTypesList = tview.NewList().
		AddItem("1. Пароли", "Хранилище учетных записей (логины, пароли)", '1', func() {
			a.openPasswordList()
		}).
		AddItem("2. Банковские карты", "Хранилище банковских карт (номер, срок, код)", '2', func() {
			a.openCardList()
		}).
		AddItem("3. Текстовые данные", "Хранилище произвольных текстовых заметок", '3', func() {
			a.openTextList()
		}).
		AddItem("4. Файлы", "Хранилище бинарных файлов (загрузка и скачивание)", '4', func() {
			a.openFileList()
		})

	logoutBtn := tview.NewButton("Выйти из аккаунта").
		SetSelectedFunc(a.handleLogout)

	quitBtn := tview.NewButton("Закрыть приложение").
		SetSelectedFunc(func() {
			a.tviewApp.Stop()
		})

	btnFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(logoutBtn, 24, 1, true).
		AddItem(nil, 2, 0, false).
		AddItem(quitBtn, 20, 1, false).
		AddItem(nil, 0, 1, false)

	a.dataTypesFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.dataTypesUserText, 2, 0, false).
		AddItem(a.dataTypesList, 0, 1, true).
		AddItem(btnFlex, 3, 0, false)

	a.dataTypesFlex.SetBorder(true).
		SetTitle(" [::b]Типы данных[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkGreen)

	a.userView = a.dataTypesFlex
	a.userInfoTextView = a.dataTypesUserText
}

// handleLogin выполняет запрос аутентификации пользователя
func (a *App) handleLogin() {
	login := strings.TrimSpace(a.loginUserField.GetText())
	password := a.loginPassField.GetText()

	if login == "" || password == "" {
		a.showErrorModal("Заполните логин и пароль", "login")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var header, trailer metadata.MD
	req := &pb.Request{}
	req.SetLogin(login)
	req.SetPassword(password)

	resp, err := a.usersClient.LoginUser(ctx, req, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка входа:\n%s", st.Message()), "login")
		return
	}

	token := extractTokenFromHeader(header, trailer)
	if token == "" {
		a.showErrorModal("Ошибка аутентификации: сервер не вернул токен авторизации", "login")
		return
	}

	a.session = &Session{
		Token:  token,
		UserID: resp.GetUser().GetId(),
		Login:  resp.GetUser().GetLogin(),
	}

	a.updateUserInfo()
	a.pages.SwitchToPage("data_types")
}

// handleRegister выполняет запрос регистрации нового пользователя
func (a *App) handleRegister() {
	login := strings.TrimSpace(a.regUserField.GetText())
	password := a.regPassField.GetText()
	passwordConfirm := a.regPassField2.GetText()

	if login == "" || password == "" {
		a.showErrorModal("Логин и пароль не могут быть пустыми", "register")
		return
	}

	if password != passwordConfirm {
		a.showErrorModal("Пароли не совпадают", "register")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var header, trailer metadata.MD
	req := &pb.Request{}
	req.SetLogin(login)
	req.SetPassword(password)

	resp, err := a.usersClient.RegisterUser(ctx, req, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка регистрации:\n%s", st.Message()), "register")
		return
	}

	token := extractTokenFromHeader(header, trailer)
	if token == "" {
		a.showErrorModal("Ошибка регистрации: сервер не вернул токен авторизации", "register")
		return
	}

	a.session = &Session{
		Token:  token,
		UserID: resp.GetUser().GetId(),
		Login:  resp.GetUser().GetLogin(),
	}

	a.updateUserInfo()
	a.pages.SwitchToPage("data_types")
}

// handleLogout выполняет выход из текущей сессии
func (a *App) handleLogout() {
	a.session = nil
	a.currentPasswords = nil
	a.currentPassword = nil
	a.currentPage = 1
	a.currentCards = nil
	a.currentCard = nil
	a.currentCardPage = 1
	a.currentTexts = nil
	a.currentText = nil
	a.currentTextPage = 1
	a.currentFiles = nil
	a.currentFile = nil
	a.currentFilePage = 1
	a.clearLoginFields()
	a.clearRegisterFields()
	a.clearCreatePasswordFields()
	a.clearCreateCardFields()
	a.clearCreateTextFields()
	a.clearUploadFileFields()
	a.clearDownloadFileFields()
	a.pages.SwitchToPage("login")
}

func (a *App) getLoginTitle() string {
	return fmt.Sprintf(" [::b]Вход в gKeep[::-] [gray](%s)[-] ", a.version)
}

func (a *App) getRegisterTitle() string {
	return fmt.Sprintf(" [::b]Регистрация[::-] [gray](%s)[-] ", a.version)
}

// clearLoginFields очищает поля формы логина
func (a *App) clearLoginFields() {
	a.loginUserField.SetText("")
	a.loginPassField.SetText("")
}

// clearRegisterFields очищает поля формы регистрации
func (a *App) clearRegisterFields() {
	a.regUserField.SetText("")
	a.regPassField.SetText("")
	a.regPassField2.SetText("")
}
