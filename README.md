# 🧾 Software de Fechamento de Folha - Go

Este projeto é um utilitário de **fechamento de folha de pagamento**, desenvolvido com foco em performance, praticidade e segurança. Ele permite ao usuário **selecionar um servidor PostgreSQL**, informar o **ID da folha**, e **executar o fechamento** com um clique, atualizando o status da folha diretamente no banco de dados.

---

## 💡 Funcionalidades

- Conexão com dois servidores PostgreSQL via interface gráfica
- Fechamento automático da folha de pagamento (update no status da folha)
- Geração de relatório CSV com status das folhas
- Interface com tema azul, botão verde e suporte a **modo escuro**
- Autenticação com **login e senha**
- Empacotamento como `.exe` com ícone personalizado para Windows

---

## 🛠 Tecnologias Utilizadas

- **Go (Golang)** – linguagem principal
- **Fyne** – biblioteca para a interface gráfica em Go
- **PostgreSQL** – banco de dados
- **SQL** – execução de scripts para atualização
- **CSV** – exportação de relatórios

---

## 📦 Como Executar

### 1. Clone o repositório:

```bash
git clone https://github.com/seuusuario/fechamento-folha-go.git
cd fechamento-folha-go
