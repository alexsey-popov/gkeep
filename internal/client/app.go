package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	api "github.com/alexsey-popov/gkeep/proto"
	pbPassword "github.com/alexsey-popov/gkeep/proto/password"
	pb "github.com/alexsey-popov/gkeep/proto/user"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Session хранит данные текущего аутентифицированного пользователя
type Session struct {
	Token  string
	UserID string
	Login  string
}

// App представляет TUI-приложение клиента
type App struct {
	tviewApp    *tview.Application
	pages       *tview.Pages
	usersClient api.APIClient
	session     *Session

	// Экраны приложения
	loginForm          *tview.Form
	registerForm       *tview.Form
	userView           *tview.Flex // Сохранено для совместимости
	dataTypesFlex      *tview.Flex
	dataTypesList      *tview.List
	passwordsFlex      *tview.Flex
	passwordsList      *tview.List
	passwordViewFlex   *tview.Flex
	passwordCreateForm *tview.Form
	passwordEditForm   *tview.Form

	// Поля ввода и компоненты отображения
	loginUserField      *tview.InputField
	loginPassField      *tview.InputField
	regUserField        *tview.InputField
	regPassField        *tview.InputField
	regPassField2       *tview.InputField
	userInfoTextView    *tview.TextView
	dataTypesUserText   *tview.TextView
	passwordDetailsText *tview.TextView

	// Поля создания пароля
	createPassNameField  *tview.InputField
	createPassLoginField *tview.InputField
	createPassValField   *tview.InputField

	// Поля редактирования пароля
	editPassNameField  *tview.InputField
	editPassLoginField *tview.InputField
	editPassValField   *tview.InputField

	// Текущее состояние данных
	currentPasswords []*pbPassword.Password
	currentPassword  *pbPassword.Password
}

// NewApp создает новый экземпляр терминального приложения
func NewApp(usersClient api.APIClient) *App {
	a := &App{
		tviewApp:         tview.NewApplication(),
		pages:            tview.NewPages(),
		usersClient:      usersClient,
		currentPasswords: make([]*pbPassword.Password, 0),
	}

	a.initUI()
	return a
}

// initUI инициализирует все экраны и формы приложения
func (a *App) initUI() {
	a.initLoginForm()
	a.initRegisterForm()
	a.initDataTypesView()
	a.initPasswordsListView()
	a.initPasswordDetailView()
	a.initPasswordCreateForm()
	a.initPasswordEditForm()

	a.pages.AddPage("login", centerPrimitive(a.loginForm, 54, 15), true, true)
	a.pages.AddPage("register", centerPrimitive(a.registerForm, 54, 18), true, false)
	a.pages.AddPage("data_types", centerPrimitive(a.dataTypesFlex, 64, 18), true, false)
	a.pages.AddPage("user", centerPrimitive(a.userView, 64, 18), true, false)
	a.pages.AddPage("password_list", centerPrimitive(a.passwordsFlex, 70, 22), true, false)
	a.pages.AddPage("password_view", centerPrimitive(a.passwordViewFlex, 64, 18), true, false)
	a.pages.AddPage("password_create", centerPrimitive(a.passwordCreateForm, 60, 18), true, false)
	a.pages.AddPage("password_edit", centerPrimitive(a.passwordEditForm, 60, 18), true, false)

	a.tviewApp.SetRoot(a.pages, true).EnableMouse(true)
}

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
		SetTitle(" [::b]Вход в gKeep[::-] ").
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
		SetTitle(" [::b]Регистрация[::-] ").
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

	addBtn := tview.NewButton("Добавить пароль").
		SetSelectedFunc(func() {
			a.clearCreatePasswordFields()
			a.pages.SwitchToPage("password_create")
		})

	refreshBtn := tview.NewButton("Обновить").
		SetSelectedFunc(func() {
			a.loadPasswords()
		})

	backBtn := tview.NewButton("Назад").
		SetSelectedFunc(func() {
			a.pages.SwitchToPage("data_types")
		})

	btnFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(openBtn, 12, 1, true).
		AddItem(nil, 1, 0, false).
		AddItem(addBtn, 18, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(refreshBtn, 12, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(backBtn, 10, 1, false)

	a.passwordsFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.passwordsList, 0, 1, true).
		AddItem(btnFlex, 3, 0, false)

	a.passwordsFlex.SetBorder(true).
		SetTitle(" [::b]Список паролей[::-] ").
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
	a.editPassNameField = tview.NewInputField().
		SetLabel("Название: ").
		SetFieldWidth(32)
	a.editPassNameField.SetDisabled(true)

	a.editPassLoginField = tview.NewInputField().
		SetLabel("Логин: ").
		SetFieldWidth(32)

	a.editPassValField = tview.NewInputField().
		SetLabel("Новый пароль: ").
		SetFieldWidth(32)

	a.passwordEditForm = tview.NewForm().
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
	req := &pb.UserRequest{}
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
	req := &pb.UserRequest{}
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
	a.clearLoginFields()
	a.clearRegisterFields()
	a.clearCreatePasswordFields()
	a.pages.SwitchToPage("login")
}

// openPasswordList загружает список паролей и переключает на экран списка
func (a *App) openPasswordList() {
	a.loadPasswords()
	a.pages.SwitchToPage("password_list")
}

// loadPasswords выполняет запрос к серверу для получения списка паролей пользователя
func (a *App) loadPasswords() {
	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &pbPassword.PasswordListRequest{}
	resp, err := a.usersClient.PasswordList(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка получения списка паролей:\n%s", st.Message()), "data_types")
		return
	}

	a.currentPasswords = resp.GetPassword()
	a.passwordsList.Clear()

	if len(a.currentPasswords) == 0 {
		a.passwordsList.AddItem("[gray]Нет сохраненных паролей[-]", "Нажмите кнопку 'Добавить пароль' для создания записи", 0, nil)
		return
	}

	for i, item := range a.currentPasswords {
		itemCopy := item
		a.passwordsList.AddItem(
			itemCopy.GetName(),
			"",
			rune('0'+i+1),
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

	info := fmt.Sprintf(
		"\n  [yellow]Название:[white] %s\n\n"+
			"  [yellow]Логин:[white]    %s\n\n"+
			"  [yellow]Пароль:[white]   %s\n",
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

	req := &pbPassword.PasswordCreateRequest{}
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
	a.loadPasswords()
	a.pages.SwitchToPage("password_list")
}

// handlePasswordUpdate обновляет пароль выбранной записи
func (a *App) handlePasswordUpdate() {
	login := strings.TrimSpace(a.editPassLoginField.GetText())
	passwordVal := a.editPassValField.GetText()

	if login == "" {
		a.showErrorModal("Логин не может быть пустым", "password_edit")
		return
	}

	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &pbPassword.PasswordUpdateRequest{}
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

	req := &pbPassword.PasswordDeleteRequest{}
	req.SetId(a.currentPassword.GetId())

	_, err := a.usersClient.PasswordDelete(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка удаления записи:\n%s", st.Message()), "password_view")
		return
	}

	a.currentPassword = nil
	a.loadPasswords()
	a.pages.SwitchToPage("password_list")
}

// authCtx создает контекст с метаданными токена аутентификации текущей сессии
func (a *App) authCtx(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	if a.session != nil && a.session.Token != "" {
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(
			"authorization", "Bearer "+a.session.Token,
			"token", a.session.Token,
		))
	}
	return ctx, cancel
}

// updateUserInfo обновляет информацию о текущем пользователе на экране кабинета/типов данных
func (a *App) updateUserInfo() {
	if a.session == nil {
		return
	}

	info := fmt.Sprintf(
		"[green]Пользователь:[white] %s   [gray](ID: %s)[-]\n[gray]Выберите тип данных для работы с хранилищем:[-]",
		a.session.Login,
		a.session.UserID,
	)

	a.dataTypesUserText.SetText(info)
}

// showErrorModal показывает модальное окно с сообщением об ошибке
func (a *App) showErrorModal(message string, backPage string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("error_modal")
			a.pages.SwitchToPage(backPage)
		})

	a.pages.AddPage("error_modal", modal, true, true)
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

// clearCreatePasswordFields очищает поля формы создания пароля
func (a *App) clearCreatePasswordFields() {
	a.createPassNameField.SetText("")
	a.createPassLoginField.SetText("")
	a.createPassValField.SetText("")
}

// extractTokenFromHeader извлекает токен из полученных заголовков или трейлеров ответа
func extractTokenFromHeader(mds ...metadata.MD) string {
	for _, md := range mds {
		if md == nil {
			continue
		}
		if vals := md.Get("token"); len(vals) > 0 && vals[0] != "" {
			return strings.TrimSpace(vals[0])
		}
		if vals := md.Get("authorization"); len(vals) > 0 && vals[0] != "" {
			val := vals[0]
			if strings.HasPrefix(strings.ToLower(val), "bearer ") {
				return strings.TrimSpace(val[7:])
			}
			return strings.TrimSpace(val)
		}
		if vals := md.Get("jwt"); len(vals) > 0 && vals[0] != "" {
			return strings.TrimSpace(vals[0])
		}
	}
	return ""
}

// centerPrimitive центрирует примитив по экрану с заданными размерами
func centerPrimitive(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}

// Run запускает главный цикл терминального приложения
func (a *App) Run() error {
	return a.tviewApp.Run()
}

// GetSession возвращает текущую сессию
func (a *App) GetSession() *Session {
	return a.session
}
