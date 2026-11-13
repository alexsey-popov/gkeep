// Клиентское приложение (TUI)
package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	api "github.com/alexsey-popov/gkeep/proto"
	pbCard "github.com/alexsey-popov/gkeep/proto/card"
	pbFile "github.com/alexsey-popov/gkeep/proto/file"
	pbPassword "github.com/alexsey-popov/gkeep/proto/password"
	pbText "github.com/alexsey-popov/gkeep/proto/text"
	"github.com/rivo/tview"
	"google.golang.org/grpc/metadata"
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
	version     string

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
	cardsFlex          *tview.Flex
	cardsList          *tview.List
	cardViewFlex       *tview.Flex
	cardCreateForm     *tview.Form
	cardEditForm       *tview.Form
	textsFlex          *tview.Flex
	textsList          *tview.List
	textViewFlex       *tview.Flex
	textCreateForm     *tview.Form
	textEditForm       *tview.Form
	filesFlex          *tview.Flex
	filesList          *tview.List
	fileViewFlex       *tview.Flex
	fileUploadForm     *tview.Form
	fileDownloadForm   *tview.Form

	// Поля ввода и компоненты отображения
	loginUserField        *tview.InputField
	loginPassField        *tview.InputField
	regUserField          *tview.InputField
	regPassField          *tview.InputField
	regPassField2         *tview.InputField
	userInfoTextView      *tview.TextView
	dataTypesUserText     *tview.TextView
	passwordDetailsText   *tview.TextView
	cardDetailsText       *tview.TextView
	textDetailsText       *tview.TextView
	fileDetailsText       *tview.TextView
	uploadFileNameField   *tview.InputField
	uploadFilePathField   *tview.InputField
	downloadSavePathField *tview.InputField

	// Поля создания пароля
	createPassNameField  *tview.InputField
	createPassLoginField *tview.InputField
	createPassValField   *tview.InputField

	// Поля редактирования пароля
	editPassIDField    *tview.InputField
	editPassNameField  *tview.InputField
	editPassLoginField *tview.InputField
	editPassValField   *tview.InputField

	// Поля создания карты
	createCardNameField    *tview.InputField
	createCardNumberField  *tview.InputField
	createCardExpiredField *tview.InputField
	createCardCodeField    *tview.InputField

	// Поля редактирования карты
	editCardIDField      *tview.InputField
	editCardNameField    *tview.InputField
	editCardNumberField  *tview.InputField
	editCardExpiredField *tview.InputField
	editCardCodeField    *tview.InputField

	// Поля создания текста
	createTextNameField *tview.InputField
	createTextValField  *tview.InputField

	// Поля редактирования текста
	editTextIDField   *tview.InputField
	editTextNameField *tview.InputField
	editTextValField  *tview.InputField

	// Кнопки пагинации
	prevPageBtn     *tview.Button
	nextPageBtn     *tview.Button
	prevCardPageBtn *tview.Button
	nextCardPageBtn *tview.Button
	prevTextPageBtn *tview.Button
	nextTextPageBtn *tview.Button
	prevFilePageBtn *tview.Button
	nextFilePageBtn *tview.Button

	// Текущее состояние данных
	currentPasswords []*pbPassword.Password
	currentPassword  *pbPassword.Password
	currentPage      int
	currentCards     []*pbCard.Card
	currentCard      *pbCard.Card
	currentCardPage  int
	currentTexts     []*pbText.Text
	currentText      *pbText.Text
	currentTextPage  int
	currentFiles     []*pbFile.File
	currentFile      *pbFile.File
	currentFilePage  int
}

// NewApp создает новый экземпляр терминального приложения
func NewApp(usersClient api.APIClient, version ...string) *App {
	ver := ""
	if len(version) > 0 {
		ver = strings.TrimSpace(version[0])
	}

	a := &App{
		tviewApp:         tview.NewApplication(),
		pages:            tview.NewPages(),
		usersClient:      usersClient,
		version:          ver,
		currentPage:      1,
		currentPasswords: make([]*pbPassword.Password, 0),
		currentCardPage:  1,
		currentCards:     make([]*pbCard.Card, 0),
		currentTextPage:  1,
		currentTexts:     make([]*pbText.Text, 0),
		currentFilePage:  1,
		currentFiles:     make([]*pbFile.File, 0),
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
	a.initCardsListView()
	a.initCardDetailView()
	a.initCardCreateForm()
	a.initCardEditForm()
	a.initTextsListView()
	a.initTextDetailView()
	a.initTextCreateForm()
	a.initTextEditForm()
	a.initFilesListView()
	a.initFileDetailView()
	a.initFileUploadForm()
	a.initFileDownloadForm()

	a.pages.AddPage("login", centerPrimitive(a.loginForm, 54, 15), true, true)
	a.pages.AddPage("register", centerPrimitive(a.registerForm, 54, 18), true, false)
	a.pages.AddPage("data_types", centerPrimitive(a.dataTypesFlex, 64, 18), true, false)
	a.pages.AddPage("user", centerPrimitive(a.userView, 64, 18), true, false)
	a.pages.AddPage("password_list", centerPrimitive(a.passwordsFlex, 70, 22), true, false)
	a.pages.AddPage("password_view", centerPrimitive(a.passwordViewFlex, 64, 18), true, false)
	a.pages.AddPage("password_create", centerPrimitive(a.passwordCreateForm, 60, 18), true, false)
	a.pages.AddPage("password_edit", centerPrimitive(a.passwordEditForm, 60, 18), true, false)
	a.pages.AddPage("card_list", centerPrimitive(a.cardsFlex, 70, 22), true, false)
	a.pages.AddPage("card_view", centerPrimitive(a.cardViewFlex, 64, 20), true, false)
	a.pages.AddPage("card_create", centerPrimitive(a.cardCreateForm, 60, 20), true, false)
	a.pages.AddPage("card_edit", centerPrimitive(a.cardEditForm, 60, 22), true, false)
	a.pages.AddPage("text_list", centerPrimitive(a.textsFlex, 70, 22), true, false)
	a.pages.AddPage("text_view", centerPrimitive(a.textViewFlex, 64, 20), true, false)
	a.pages.AddPage("text_create", centerPrimitive(a.textCreateForm, 60, 18), true, false)
	a.pages.AddPage("text_edit", centerPrimitive(a.textEditForm, 60, 20), true, false)
	a.pages.AddPage("file_list", centerPrimitive(a.filesFlex, 70, 22), true, false)
	a.pages.AddPage("file_view", centerPrimitive(a.fileViewFlex, 64, 20), true, false)
	a.pages.AddPage("file_upload", centerPrimitive(a.fileUploadForm, 60, 18), true, false)
	a.pages.AddPage("file_download", centerPrimitive(a.fileDownloadForm, 60, 18), true, false)

	a.tviewApp.SetRoot(a.pages, true).EnableMouse(true)
}

// Run запускает главный цикл терминального приложения
func (a *App) Run() error {
	return a.tviewApp.Run()
}

// GetSession возвращает текущую сессию
func (a *App) GetSession() *Session {
	return a.session
}

// GetVersion возвращает установленную версию приложения
func (a *App) GetVersion() string {
	return a.version
}

// SetVersion обновляет версию приложения и заголовок формы входа
func (a *App) SetVersion(v string) {
	a.version = strings.TrimSpace(v)
	a.loginForm.SetTitle(a.getLoginTitle())
	a.registerForm.SetTitle(a.getRegisterTitle())
	if a.session != nil {
		a.updateUserInfo()
	}
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

	versionStr := ""
	if a.version != "" {
		versionStr = fmt.Sprintf("   [gray]Версия:[white] %s[-]", a.version)
	}

	info := fmt.Sprintf(
		"[green]Пользователь:[white] %s   [gray][-]%s\n[gray]Выберите тип данных для работы с хранилищем:[-]",
		a.session.Login,
		versionStr,
	)

	a.dataTypesUserText.SetText(info)
}

// showInfoModal показывает информационное модальное окно
func (a *App) showInfoModal(message string, backPage string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("info_modal")
			a.pages.SwitchToPage(backPage)
		})

	a.pages.AddPage("info_modal", modal, true, true)
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
