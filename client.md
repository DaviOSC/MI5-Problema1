# Explicação do Código do Cliente (`client.go`)

Este documento explica linha por linha o código do cliente, detalhando o que cada parte faz e por que foi usada.

---

## Estrutura do `client.go`

O código do cliente é dividido em:
1. **Importações** (bibliotecas necessárias)
2. **Função principal** (`main`)

---

## 1. Importações

```go
import (
    "bufio"
    "fmt"
    "net"
    "os"
    "strings"
)
```

- **O que faz**: Importa bibliotecas necessárias para o funcionamento do cliente.
- **Por que foi usada**:
  - `bufio`: Para leitura de entrada do usuário.
  - `fmt`: Para exibição de mensagens no terminal.
  - `net`: Para comunicação TCP com o servidor.
  - `os`: Para interação com o sistema operacional (ex: encerrar o programa).
  - `strings`: Para manipulação de strings.

---

## 2. Função Principal (`main`)

### Inicialização e Autenticação

```go
func main() {
    reader := bufio.NewReader(os.Stdin)

    // Solicitar credenciais
    fmt.Print("Usuário: ")
    user, _ := reader.ReadString('\n')
    user = strings.TrimSpace(user)

    fmt.Print("Senha: ")
    password, _ := reader.ReadString('\n')
    password = strings.TrimSpace(password)
```

- **O que faz**: Solicita ao usuário que insira o nome de usuário e senha.
- **Por que foi usada**: Para autenticar o cliente no servidor.
- **Detalhes**:
  - `reader`: Lê a entrada do usuário.
  - `strings.TrimSpace`: Remove espaços em branco e quebras de linha da entrada.

---

### Conexão com o Servidor

```go
    // Conectar ao servidor
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        fmt.Println("Erro ao conectar:", err)
        return
    }
    defer conn.Close()
```

- **O que faz**: Estabelece uma conexão TCP com o servidor na porta 8080.
- **Por que foi usada**: Para enviar e receber dados do servidor.
- **Detalhes**:
  - `net.Dial`: Conecta ao servidor.
  - `defer conn.Close()`: Garante que a conexão seja fechada ao final da execução.

---

### Envio de Credenciais

```go
    // Enviar credenciais
    conn.Write([]byte(fmt.Sprintf("%s:%s", user, password)))
```

- **O que faz**: Envia as credenciais (usuário e senha) para o servidor.
- **Por que foi usada**: Para autenticar o cliente no servidor.
- **Detalhes**:
  - `conn.Write`: Envia dados para o servidor.
  - `fmt.Sprintf`: Formata a string no formato `user:password`.

---

### Recebimento da Resposta de Autenticação

```go
    // Receber resposta de autenticação
    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil || string(buffer[:n]) != "Autenticado com sucesso!" {
        fmt.Println("Autenticação falhou!")
        return
    }
    fmt.Println("Autenticado com sucesso!")
```

- **O que faz**: Recebe a resposta do servidor após a autenticação.
- **Por que foi usada**: Para verificar se o login foi bem-sucedido.
- **Detalhes**:
  - `buffer`: Armazena a resposta do servidor.
  - `conn.Read`: Lê a resposta do servidor.
  - Se a resposta não for `"Autenticado com sucesso!"`, o cliente é encerrado.

---

### Menu Interativo

```go
    // Menu interativo
    for {
        fmt.Print("\nComandos disponíveis:\n")
        fmt.Println("1. Recomendar posto")
        fmt.Println("2. Reservar posto")
        fmt.Println("3. Sair")
        fmt.Print("Escolha: ")

        choice, _ := reader.ReadString('\n')
        choice = strings.TrimSpace(choice)
```

- **O que faz**: Exibe um menu de opções para o usuário.
- **Por que foi usada**: Para permitir que o usuário interaja com o servidor.
- **Detalhes**:
  - `fmt.Print` e `fmt.Println`: Exibem o menu no terminal.
  - `reader.ReadString`: Lê a escolha do usuário.

---

### Processamento de Comandos

#### Recomendar Posto

```go
        switch choice {
        case "1":
            conn.Write([]byte("recommend"))
            n, _ := conn.Read(buffer)
            response := string(buffer[:n])
            fmt.Println("\n" + response)
```

- **O que faz**: Solicita ao servidor que recomende um posto de recarga.
- **Por que foi usada**: Para obter o posto mais próximo disponível.
- **Detalhes**:
  - `conn.Write`: Envia o comando `"recommend"` ao servidor.
  - `conn.Read`: Recebe a resposta do servidor.
  - `fmt.Println`: Exibe a resposta no terminal.

---

#### Reservar Posto

```go
        case "2":
            fmt.Print("ID do posto para reservar: ")
            id, _ := reader.ReadString('\n')
            id = strings.TrimSpace(id)
            conn.Write([]byte("reserve " + id))
            n, _ := conn.Read(buffer)
            fmt.Println("\n" + string(buffer[:n]))
```

- **O que faz**: Solicita ao servidor que reserve um posto específico.
- **Por que foi usada**: Para garantir que o posto esteja disponível para o carro.
- **Detalhes**:
  - `fmt.Print`: Solicita o ID do posto.
  - `conn.Write`: Envia o comando `"reserve [ID]"` ao servidor.
  - `conn.Read`: Recebe a resposta do servidor.
  - `fmt.Println`: Exibe a resposta no terminal.

---

#### Sair

```go
        case "3":
            conn.Write([]byte("exit"))
            return
```

- **O que faz**: Encerra a conexão com o servidor e fecha o cliente.
- **Por que foi usada**: Para permitir que o usuário saia do programa.
- **Detalhes**:
  - `conn.Write`: Envia o comando `"exit"` ao servidor.
  - `return`: Encerra a função `main`.

---

#### Comando Inválido

```go
        default:
            fmt.Println("Opção inválida")
        }
    }
}
```

- **O que faz**: Lida com entradas inválidas do usuário.
- **Por que foi usada**: Para garantir que o programa não quebre com entradas incorretas.
- **Detalhes**:
  - `fmt.Println`: Exibe uma mensagem de erro.

---

## Resumo

- **Autenticação**: O cliente envia credenciais ao servidor e aguarda a confirmação.
- **Menu Interativo**: Permite ao usuário escolher entre recomendar um posto, reservar um posto ou sair.
- **Comunicação TCP**: Todas as interações com o servidor são feitas via TCP.

---