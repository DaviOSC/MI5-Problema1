package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Solicitar credenciais
	fmt.Print("Usuário: ")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)

	fmt.Print("Senha: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// Conectar ao servidor
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Erro ao conectar:", err)
		return
	}
	defer conn.Close()

	// Enviar credenciais
	conn.Write([]byte(fmt.Sprintf("%s:%s", user, password)))

	// Receber resposta de autenticação
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil || string(buffer[:n]) != "Autenticado com sucesso!" {
		fmt.Println("Autenticação falhou!")
		return
	}
	fmt.Println("Autenticado com sucesso!")

	// Menu interativo
	for {
		fmt.Print("\nComandos disponíveis:\n")
		fmt.Println("1. Recomendar posto")
		fmt.Println("2. Reservar posto")
		fmt.Println("3. Sair")
		fmt.Print("Escolha: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			conn.Write([]byte("recommend"))
			n, _ := conn.Read(buffer)
			response := string(buffer[:n])
			fmt.Println("\n" + response)

		case "2":
			fmt.Print("ID do posto para reservar: ")
			id, _ := reader.ReadString('\n')
			id = strings.TrimSpace(id)
			conn.Write([]byte("reserve " + id))
			n, _ := conn.Read(buffer)
			fmt.Println("\n" + string(buffer[:n]))

		case "3":
			conn.Write([]byte("exit"))
			return

		default:
			fmt.Println("Opção inválida")
		}
	}
}