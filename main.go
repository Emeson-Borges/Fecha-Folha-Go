package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var (
	password = "S551bp7fRs4qRCWx2M5y"
	user     = "postgres"
	port     = "5432"
	login    = "admin"
	senha    = "1234"
)

func main() {
	os.Setenv("FYNE_RENDER", "software")

	myApp := app.NewWithID("itarget.folha")
	mostrarLogin(myApp)
}

func mostrarLogin(myApp fyne.App) {
	loginWindow := myApp.NewWindow("Login - Fechamento de Folha")
	loginWindow.Resize(fyne.NewSize(300, 200))

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("Usuário")

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Senha")

	loginBtn := widget.NewButton("Entrar", func() {
		if userEntry.Text == login && passEntry.Text == senha {
			loginWindow.Hide()
			abrirAppPrincipal(myApp, loginWindow)
		} else {
			dialog.ShowError(fmt.Errorf("Login ou senha incorretos"), loginWindow)
		}
	})

	form := container.NewVBox(
		widget.NewLabelWithStyle("Acesso Restrito", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		userEntry,
		passEntry,
		loginBtn,
	)

	loginWindow.SetContent(container.NewCenter(form))
	// loginWindow.Show()
	loginWindow.ShowAndRun()
}

func abrirAppPrincipal(myApp fyne.App, loginWindow fyne.Window) {
	win := myApp.NewWindow("Fechamento de Folha")
	win.Resize(fyne.NewSize(500, 400))

	title := widget.NewLabelWithStyle("Fechamento de Folha", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.Wrapping = fyne.TextWrapBreak

	servidores := []string{
		"ipg04.aws.itarget.com.br",
		"ipg04-13.aws.itarget.com.br",
	}

	bancoEntry := widget.NewEntry()
	bancoEntry.SetPlaceHolder("Digite para filtrar bancos...")

	bancoSelect := widget.NewSelect([]string{}, nil)
	bancoSelect.PlaceHolder = "Selecione um banco"

	idEntry := widget.NewEntry()
	idEntry.SetPlaceHolder("ID da folha (ex: 1,2,3)")

	servidorSelect := widget.NewSelect(servidores, func(server string) {
		bancos, err := listarBancos(server)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Erro ao listar bancos: %v", err), win)
			return
		}
		bancoSelect.Options = bancos
		bancoSelect.Refresh()

		bancoEntry.OnChanged = func(text string) {
			var filtrados []string
			for _, banco := range bancos {
				if strings.Contains(strings.ToLower(banco), strings.ToLower(text)) {
					filtrados = append(filtrados, banco)
				}
			}
			bancoSelect.Options = filtrados
			bancoSelect.Refresh()
		}
	})
	servidorSelect.PlaceHolder = "Selecione o servidor"

	relatorioLabel := widget.NewLabel("")

	button := widget.NewButton("Fechar Folha", func() {
		servidor := servidorSelect.Selected
		banco := bancoSelect.Selected
		ids := strings.TrimSpace(idEntry.Text)

		if servidor == "" || banco == "" || ids == "" {
			dialog.ShowError(fmt.Errorf("Preencha todos os campos."), win)
			return
		}

		err := fecharFolha(servidor, banco, ids)
		if err != nil {
			dialog.ShowError(err, win)
		} else {
			gerarRelatorioCSV(servidor, banco, ids)
			dialog.ShowInformation("Sucesso", "Folha(s) fechada(s) com sucesso!", win)
			relatorioLabel.SetText("Relatório salvo com sucesso.")
		}
	})
	button.Importance = widget.HighImportance

	form := container.NewVBox(
		title,
		widget.NewSeparator(),
		widget.NewLabel("Servidor:"),
		servidorSelect,
		widget.NewLabel("Banco de dados:"),
		bancoEntry,
		bancoSelect,
		widget.NewLabel("ID da Folha:"),
		idEntry,
		button,
		relatorioLabel,
	)

	scroll := container.NewVScroll(form)
	scroll.SetMinSize(fyne.NewSize(480, 360))

	win.SetContent(container.NewCenter(scroll))
	win.Show()
}

func listarBancos(servidor string) ([]string, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable", servidor, port, user, password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT datname FROM pg_database WHERE datistemplate = false;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bancos []string
	for rows.Next() {
		var nome string
		if err := rows.Scan(&nome); err != nil {
			continue
		}
		bancos = append(bancos, nome)
	}

	return bancos, nil
}

func fecharFolha(servidor, banco, ids string) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", servidor, port, user, password, banco)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("Erro de conexão: %v", err)
	}
	defer db.Close()

	queries := []string{
		fmt.Sprintf("UPDATE folhas SET status = 0 WHERE id IN (%s);", ids),
		fmt.Sprintf("UPDATE orgaos_folhas SET status = 0 WHERE folha_id IN (%s);", ids),
		fmt.Sprintf("UPDATE folhas SET status = 1 WHERE id IN (%s);", ids),
		fmt.Sprintf("UPDATE orgaos_folhas SET status = 1 WHERE folha_id IN (%s);", ids),
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("Erro na transação: %v", err)
	}

	for _, query := range queries {
		if _, err := tx.Exec(query); err != nil {
			tx.Rollback()
			return fmt.Errorf("Erro ao executar: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Erro ao finalizar transação: %v", err)
	}

	return nil
}

func gerarRelatorioCSV(servidor, banco, ids string) {
	_ = os.MkdirAll("relatorios", os.ModePerm)
	nomeArquivo := filepath.Join("relatorios", fmt.Sprintf("fechamento_%s.csv", time.Now().Format("2006-01-02_15-04-05")))
	file, err := os.Create(nomeArquivo)
	if err != nil {
		return
	}
	defer file.Close()

	escritor := csv.NewWriter(file)
	defer escritor.Flush()

	escritor.Write([]string{"Data", "Servidor", "Banco", "IDs Fechados"})
	escritor.Write([]string{time.Now().Format("2006-01-02 15:04:05"), servidor, banco, ids})
}
