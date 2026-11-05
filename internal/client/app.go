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
	"time"

	api "github.com/alexsey-popov/gkeep/proto"
	pbCard "github.com/alexsey-popov/gkeep/proto/card"
	pbFile "github.com/alexsey-popov/gkeep/proto/file"
	pbPassword "github.com/alexsey-popov/gkeep/proto/password"
	pbText "github.com/alexsey-popov/gkeep/proto/text"
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
		SetSelectedFunc(func() {
			a.loadPasswords()
		})

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

// initCardsListView создает экран со списком банковских карт
func (a *App) initCardsListView() {
	a.cardsList = tview.NewList().
		ShowSecondaryText(true)

	openBtn := tview.NewButton("Открыть").
		SetSelectedFunc(func() {
			idx := a.cardsList.GetCurrentItem()
			if idx >= 0 && idx < len(a.currentCards) {
				a.openCardView(a.currentCards[idx])
			}
		})

	addBtn := tview.NewButton("Добавить").
		SetSelectedFunc(func() {
			a.clearCreateCardFields()
			a.pages.SwitchToPage("card_create")
		})

	a.prevCardPageBtn = tview.NewButton("< Пред.").
		SetSelectedFunc(a.prevCardPage)

	a.nextCardPageBtn = tview.NewButton("След. >").
		SetSelectedFunc(a.nextCardPage)

	refreshBtn := tview.NewButton("Обновить").
		SetSelectedFunc(func() {
			a.loadCards()
		})

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
		AddItem(a.prevCardPageBtn, 10, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(a.nextCardPageBtn, 10, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(refreshBtn, 11, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(backBtn, 8, 1, false)

	a.cardsFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.cardsList, 0, 1, true).
		AddItem(btnFlex, 3, 0, false)

	a.cardsFlex.SetBorder(true).
		SetTitle(a.getCardsListTitle()).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorCadetBlue)
}

// initCardDetailView создает экран просмотра отдельной банковской карты
func (a *App) initCardDetailView() {
	a.cardDetailsText = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWrap(true)

	editBtn := tview.NewButton("Изменить").
		SetSelectedFunc(func() {
			if a.currentCard != nil {
				a.openCardEdit(a.currentCard)
			}
		})

	deleteBtn := tview.NewButton("Удалить").
		SetSelectedFunc(func() {
			if a.currentCard != nil {
				a.showDeleteCardConfirmModal()
			}
		})

	backBtn := tview.NewButton("Назад").
		SetSelectedFunc(func() {
			a.pages.SwitchToPage("card_list")
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

	a.cardViewFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.cardDetailsText, 0, 1, false).
		AddItem(btnFlex, 3, 0, true)

	a.cardViewFlex.SetBorder(true).
		SetTitle(" [::b]Просмотр карты[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkGoldenrod)
}

// initCardCreateForm создает форму добавления банковской карты
func (a *App) initCardCreateForm() {
	a.createCardNameField = tview.NewInputField().
		SetLabel("Название: ").
		SetFieldWidth(32)

	a.createCardNumberField = tview.NewInputField().
		SetLabel("Номер карты: ").
		SetFieldWidth(32)

	a.createCardExpiredField = tview.NewInputField().
		SetLabel("Срок действия: ").
		SetFieldWidth(32)

	a.createCardCodeField = tview.NewInputField().
		SetLabel("Код (CVV): ").
		SetFieldWidth(32)

	a.cardCreateForm = tview.NewForm().
		AddFormItem(a.createCardNameField).
		AddFormItem(a.createCardNumberField).
		AddFormItem(a.createCardExpiredField).
		AddFormItem(a.createCardCodeField).
		AddButton("Сохранить", a.handleCardCreate).
		AddButton("Назад", func() {
			a.clearCreateCardFields()
			a.pages.SwitchToPage("card_list")
		})

	a.cardCreateForm.SetBorder(true).
		SetTitle(" [::b]Добавление карты[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorGreen)
}

// initCardEditForm создает форму редактирования банковской карты
func (a *App) initCardEditForm() {
	a.editCardIDField = tview.NewInputField().
		SetLabel("ID: ").
		SetFieldWidth(32)
	a.editCardIDField.SetDisabled(true)

	a.editCardNameField = tview.NewInputField().
		SetLabel("Название: ").
		SetFieldWidth(32)

	a.editCardNumberField = tview.NewInputField().
		SetLabel("Номер карты: ").
		SetFieldWidth(32)

	a.editCardExpiredField = tview.NewInputField().
		SetLabel("Срок действия: ").
		SetFieldWidth(32)

	a.editCardCodeField = tview.NewInputField().
		SetLabel("Код (CVV): ").
		SetFieldWidth(32)

	a.cardEditForm = tview.NewForm().
		AddFormItem(a.editCardIDField).
		AddFormItem(a.editCardNameField).
		AddFormItem(a.editCardNumberField).
		AddFormItem(a.editCardExpiredField).
		AddFormItem(a.editCardCodeField).
		AddButton("Сохранить", a.handleCardUpdate).
		AddButton("Назад", func() {
			a.pages.SwitchToPage("card_view")
		})

	a.cardEditForm.SetBorder(true).
		SetTitle(" [::b]Редактирование карты[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorOrange)
}

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
		SetSelectedFunc(func() {
			a.loadTexts()
		})

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

// openPasswordList загружает список паролей и переключает на экран списка
func (a *App) openPasswordList() {
	a.currentPage = 1
	a.loadPasswords()
	a.pages.SwitchToPage("password_list")
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
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка получения списка паролей:\n%s", st.Message()), "data_types")
		return
	}

	a.currentPasswords = resp.GetPassword()
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

// openCardList загружает список банковских карт и переключает на экран списка
func (a *App) openCardList() {
	a.currentCardPage = 1
	a.loadCards()
	a.pages.SwitchToPage("card_list")
}

// loadCards выполняет запрос к серверу для получения списка банковских карт пользователя
func (a *App) loadCards() {
	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	if a.currentCardPage < 1 {
		a.currentCardPage = 1
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &api.ListRequest{}
	req.SetPage(strconv.Itoa(a.currentCardPage))

	resp, err := a.usersClient.CardList(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка получения списка карт:\n%s", st.Message()), "data_types")
		return
	}

	a.currentCards = resp.GetCard()
	a.cardsList.Clear()
	a.cardsFlex.SetTitle(a.getCardsListTitle())

	if len(a.currentCards) == 0 {
		if a.currentCardPage > 1 {
			a.cardsList.AddItem("[gray]На этой странице нет записей[-]", "Нажмите кнопку '< Пред.' для возврата", 0, nil)
		} else {
			a.cardsList.AddItem("[gray]Нет сохраненных карт[-]", "Нажмите кнопку 'Добавить' для создания записи", 0, nil)
		}
		return
	}

	for i, item := range a.currentCards {
		itemCopy := item
		shortcut := rune(0)
		if i < 9 {
			shortcut = rune('1' + i)
		} else if i == 9 {
			shortcut = '0'
		}
		a.cardsList.AddItem(
			itemCopy.GetName(),
			"",
			shortcut,
			func() {
				a.openCardView(itemCopy)
			},
		)
	}
}

// openCardView открывает экран подробного просмотра банковской карты
func (a *App) openCardView(c *pbCard.Card) {
	if c == nil {
		return
	}
	a.currentCard = c

	info := fmt.Sprintf(
		"\n  [yellow]Название:[white]      %s\n\n"+
			"  [yellow]Номер карты:[white]   %s\n\n"+
			"  [yellow]Срок действия:[white] %s\n\n"+
			"  [yellow]Код (CVV):[white]     %s\n",
		c.GetName(),
		c.GetNumber(),
		c.GetExpired(),
		c.GetCode(),
	)

	a.cardDetailsText.SetText(info)
	a.pages.SwitchToPage("card_view")
}

// openCardEdit открывает экран редактирования банковской карты
func (a *App) openCardEdit(c *pbCard.Card) {
	if c == nil {
		return
	}
	a.currentCard = c
	a.editCardIDField.SetText(c.GetId())
	a.editCardNameField.SetText(c.GetName())
	a.editCardNumberField.SetText(c.GetNumber())
	a.editCardExpiredField.SetText(c.GetExpired())
	a.editCardCodeField.SetText(c.GetCode())
	a.pages.SwitchToPage("card_edit")
}

// handleCardCreate создает новую запись банковской карты
func (a *App) handleCardCreate() {
	name := strings.TrimSpace(a.createCardNameField.GetText())
	number := strings.TrimSpace(a.createCardNumberField.GetText())
	expired := strings.TrimSpace(a.createCardExpiredField.GetText())
	code := strings.TrimSpace(a.createCardCodeField.GetText())

	if name == "" {
		a.showErrorModal("Название записи не может быть пустым", "card_create")
		return
	}

	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &pbCard.CreateRequest{}
	req.SetName(name)
	req.SetNumber(number)
	req.SetExpired(expired)
	req.SetCode(code)

	_, err := a.usersClient.CardCreate(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка создания записи:\n%s", st.Message()), "card_create")
		return
	}

	a.clearCreateCardFields()
	a.currentCardPage = 1
	a.loadCards()
	a.pages.SwitchToPage("card_list")
}

// handleCardUpdate обновляет данные выбранной банковской карты
func (a *App) handleCardUpdate() {
	id := strings.TrimSpace(a.editCardIDField.GetText())
	name := strings.TrimSpace(a.editCardNameField.GetText())
	number := strings.TrimSpace(a.editCardNumberField.GetText())
	expired := strings.TrimSpace(a.editCardExpiredField.GetText())
	code := strings.TrimSpace(a.editCardCodeField.GetText())

	if name == "" {
		a.showErrorModal("Название не может быть пустым", "card_edit")
		return
	}

	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &pbCard.UpdateRequest{}
	req.SetId(id)
	req.SetName(name)
	req.SetNumber(number)
	req.SetExpired(expired)
	req.SetCode(code)

	resp, err := a.usersClient.CardUpdate(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка обновления записи:\n%s", st.Message()), "card_edit")
		return
	}

	if resp.GetCard() != nil {
		a.currentCard = resp.GetCard()
	} else if a.currentCard != nil {
		a.currentCard.SetName(name)
		a.currentCard.SetNumber(number)
		a.currentCard.SetExpired(expired)
		a.currentCard.SetCode(code)
	}

	a.openCardView(a.currentCard)
}

// showDeleteCardConfirmModal показывает окно подтверждения удаления банковской карты
func (a *App) showDeleteCardConfirmModal() {
	if a.currentCard == nil {
		return
	}

	name := a.currentCard.GetName()
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Вы уверены, что хотите удалить карту '%s'?", name)).
		AddButtons([]string{"Удалить", "Отмена"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("confirm_delete_card_modal")
			if buttonLabel == "Удалить" {
				a.handleCardDelete()
			} else {
				a.pages.SwitchToPage("card_view")
			}
		})

	a.pages.AddPage("confirm_delete_card_modal", modal, true, true)
}

// handleCardDelete отправляет запрос на удаление банковской карты
func (a *App) handleCardDelete() {
	if a.currentCard == nil {
		return
	}

	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &api.IDRequest{}
	req.SetId(a.currentCard.GetId())

	_, err := a.usersClient.CardDelete(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка удаления записи:\n%s", st.Message()), "card_view")
		return
	}

	a.currentCard = nil
	a.loadCards()
	if len(a.currentCards) == 0 && a.currentCardPage > 1 {
		a.currentCardPage--
		a.loadCards()
	}
	a.pages.SwitchToPage("card_list")
}

// prevCardPage переключает на предыдущую страницу списка карт
func (a *App) prevCardPage() {
	if a.currentCardPage > 1 {
		a.currentCardPage--
		a.loadCards()
	}
}

// nextCardPage переключает на следующую страницу списка карт
func (a *App) nextCardPage() {
	if len(a.currentCards) > 0 {
		a.currentCardPage++
		a.loadCards()
	}
}

// GetCurrentCardPage возвращает номер текущей страницы списка карт
func (a *App) GetCurrentCardPage() int {
	return a.currentCardPage
}

// SetCurrentCardPage устанавливает номер текущей страницы списка карт
func (a *App) SetCurrentCardPage(p int) {
	if p < 1 {
		p = 1
	}
	a.currentCardPage = p
}

func (a *App) getCardsListTitle() string {
	if a.currentCardPage <= 1 {
		return " [::b]Список банковских карт (Стр. 1)[::-] "
	}
	return fmt.Sprintf(" [::b]Список банковских карт (Стр. %d)[::-] ", a.currentCardPage)
}

// clearCreateCardFields очищает поля формы создания карты
func (a *App) clearCreateCardFields() {
	a.createCardNameField.SetText("")
	a.createCardNumberField.SetText("")
	a.createCardExpiredField.SetText("")
	a.createCardCodeField.SetText("")
}

// openTextList загружает список текстовых записей и переключает на экран списка
func (a *App) openTextList() {
	a.currentTextPage = 1
	a.loadTexts()
	a.pages.SwitchToPage("text_list")
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
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка получения списка текстов:\n%s", st.Message()), "data_types")
		return
	}

	a.currentTexts = resp.GetText()
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

	info := fmt.Sprintf(
		"\n  [yellow]Название:[white] %s\n\n"+
			"  [yellow]Текст:[white]\n  %s\n",
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

func (a *App) getTextsListTitle() string {
	if a.currentTextPage <= 1 {
		return " [::b]Список текстовых данных (Стр. 1)[::-] "
	}
	return fmt.Sprintf(" [::b]Список текстовых данных (Стр. %d)[::-] ", a.currentTextPage)
}

// clearCreateTextFields очищает поля формы создания текстовой записи
func (a *App) clearCreateTextFields() {
	a.createTextNameField.SetText("")
	a.createTextValField.SetText("")
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
				st := status.Convert(sendErr)
				a.showErrorModal(fmt.Sprintf("Ошибка отправки чанка файла:\n%s", st.Message()), "file_upload")
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

func (a *App) getLoginTitle() string {
	return fmt.Sprintf(" [::b]Вход в gKeep[::-] [gray](%s)[-] ", a.version)
}

func (a *App) getRegisterTitle() string {
	return fmt.Sprintf(" [::b]Регистрация[::-] [gray](%s)[-] ", a.version)
}

func (a *App) getPasswordsListTitle() string {
	if a.currentPage <= 1 {
		return " [::b]Список паролей (Стр. 1)[::-] "
	}
	return fmt.Sprintf(" [::b]Список паролей (Стр. %d)[::-] ", a.currentPage)
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
