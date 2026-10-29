package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "github.com/alexsey-popov/gkeep/internal/handler/proto"
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
	usersClient pb.UsersClient
	session     *Session

	loginForm    *tview.Form
	registerForm *tview.Form
	userView     *tview.Flex

	// Поля ввода для очистки/доступа
	loginUserField   *tview.InputField
	loginPassField   *tview.InputField
	regUserField     *tview.InputField
	regPassField     *tview.InputField
	regPassField2    *tview.InputField
	userInfoTextView *tview.TextView
}

// NewApp создает новый экземпляр терминального приложения
func NewApp(usersClient pb.UsersClient) *App {
	a := &App{
		tviewApp:    tview.NewApplication(),
		pages:       tview.NewPages(),
		usersClient: usersClient,
	}

	a.initUI()
	return a
}

// initUI инициализирует все экраны и формы приложения
func (a *App) initUI() {
	a.initLoginForm()
	a.initRegisterForm()
	a.initUserView()

	a.pages.AddPage("login", centerPrimitive(a.loginForm, 54, 15), true, true)
	a.pages.AddPage("register", centerPrimitive(a.registerForm, 54, 18), true, false)
	a.pages.AddPage("user", centerPrimitive(a.userView, 64, 18), true, false)

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
		SetTitle(" [::b]Регистрация в gKeep[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkCyan)
}

// initUserView создает экран личного кабинета после успешной аутентификации
func (a *App) initUserView() {
	a.userInfoTextView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetWrap(true)

	logoutBtn := tview.NewButton("Выйти из аккаунта").
		SetSelectedFunc(func() {
			a.session = nil
			a.clearLoginFields()
			a.clearRegisterFields()
			a.pages.SwitchToPage("login")
		})

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

	a.userView = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.userInfoTextView, 0, 1, false).
		AddItem(btnFlex, 3, 1, true)

	a.userView.SetBorder(true).
		SetTitle(" [::b]Личный кабинет gKeep[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkGreen)
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

	var header metadata.MD
	req := &pb.UserRequest{}
	req.SetLogin(login)
	req.SetPassword(password)

	resp, err := a.usersClient.LoginUser(ctx, req, grpc.Header(&header))
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка входа:\n%s", st.Message()), "login")
		return
	}

	token := extractTokenFromHeader(header)
	a.session = &Session{
		Token:  token,
		UserID: resp.GetUser().GetId(),
		Login:  resp.GetUser().GetLogin(),
	}

	a.updateUserInfo()
	a.pages.SwitchToPage("user")
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

	var header metadata.MD
	req := &pb.UserRequest{}
	req.SetLogin(login)
	req.SetPassword(password)

	resp, err := a.usersClient.RegisterUser(ctx, req, grpc.Header(&header))
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка регистрации:\n%s", st.Message()), "register")
		return
	}

	token := extractTokenFromHeader(header)
	a.session = &Session{
		Token:  token,
		UserID: resp.GetUser().GetId(),
		Login:  resp.GetUser().GetLogin(),
	}

	a.updateUserInfo()
	a.pages.SwitchToPage("user")
}

// updateUserInfo обновляет информацию о текущем пользователе на экране кабинета
func (a *App) updateUserInfo() {
	if a.session == nil {
		return
	}

	tokenInfo := "отсутствует"
	if a.session.Token != "" {
		if len(a.session.Token) > 24 {
			tokenInfo = a.session.Token[:10] + "..." + a.session.Token[len(a.session.Token)-10:]
		} else {
			tokenInfo = a.session.Token
		}
	}

	info := fmt.Sprintf(
		"\n[green::b]Успешная аутентификация![::-]\n\n"+
			"[yellow]ID пользователя:[white] %s\n"+
			"[yellow]Логин:[white] %s\n"+
			"[yellow]JWT токен:[white] %s\n\n"+
			"[gray]Вы успешно вошли в систему gKeep.[-]",
		a.session.UserID,
		a.session.Login,
		tokenInfo,
	)

	a.userInfoTextView.SetText(info)
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

// extractTokenFromHeader извлекает токен из полученных заголовков ответа
func extractTokenFromHeader(header metadata.MD) string {
	if vals := header.Get("token"); len(vals) > 0 && vals[0] != "" {
		return vals[0]
	}
	if vals := header.Get("authorization"); len(vals) > 0 && vals[0] != "" {
		val := vals[0]
		if strings.HasPrefix(strings.ToLower(val), "bearer ") {
			return strings.TrimSpace(val[7:])
		}
		return strings.TrimSpace(val)
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

// Run запускает терминальное приложение
func (a *App) Run() error {
	return a.tviewApp.Run()
}

// Stop останавливает приложение
func (a *App) Stop() {
	a.tviewApp.Stop()
}

// GetSession возвращает текущую сессию
func (a *App) GetSession() *Session {
	return a.session
}
