package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	api "github.com/alexsey-popov/gkeep/proto"
	pbCard "github.com/alexsey-popov/gkeep/proto/card"
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

	// Поля ввода и компоненты отображения
	loginUserField      *tview.InputField
	loginPassField      *tview.InputField
	regUserField        *tview.InputField
	regPassField        *tview.InputField
	regPassField2       *tview.InputField
	userInfoTextView    *tview.TextView
	dataTypesUserText   *tview.TextView
	passwordDetailsText *tview.TextView
	cardDetailsText     *tview.TextView
	textDetailsText     *tview.TextView

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
	a.currentPage = 1
	a.currentCards = nil
	a.currentCard = nil
	a.currentCardPage = 1
	a.currentTexts = nil
	a.currentText = nil
	a.currentTextPage = 1
	a.clearLoginFields()
	a.clearRegisterFields()
	a.clearCreatePasswordFields()
	a.clearCreateCardFields()
	a.clearCreateTextFields()
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

	req := &pbPassword.PasswordListRequest{}
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

	req := &pbPassword.PasswordUpdateRequest{}
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

	req := &pbCard.CardListRequest{}
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

	req := &pbCard.CardCreateRequest{}
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

	req := &pbCard.CardUpdateRequest{}
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

	req := &pbCard.CardDeleteRequest{}
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

	req := &pbText.TextListRequest{}
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

	req := &pbText.TextCreateRequest{}
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

	req := &pbText.TextUpdateRequest{}
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

	req := &pbText.TextDeleteRequest{}
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
