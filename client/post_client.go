package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
)

func main() {
    // Conectar ao servidor pelo endereço localhost
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        fmt.Println("Erro ao conectar:", err)
        return
    }
    defer conn.Close()

    // Lendo a resposta do servidor
    mensagem, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Println("Servidor respondeu:", mensagem)

	for {
        // Enviar mensagem para testar
        fmt.Print("Digite algo para enviar ao servidor: ")
        input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
        fmt.Fprintln(conn, input)

        resposta, _ := bufio.NewReader(conn).ReadString('\n')
        fmt.Println("Resposta do servidor:", resposta)

        if input == "exit\n" {
            break
        }
    }
}