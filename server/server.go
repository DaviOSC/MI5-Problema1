package main

import (
    "bufio"
    "fmt"
    "net"
)
type CustomConn struct {
    net.Conn
    tipo string
}

func main() {
    endereco := ":8080" // Alterado para escutar em todas as interfaces
    listener, _ := net.Listen("tcp", endereco)
    defer listener.Close()
    fmt.Println("Servidor TCP rodando na porta", endereco,"...")

    for {
        conn, _ := listener.Accept()
        customConn := CustomConn{Conn: conn, tipo: "desconhecido"}
        go handleConnection(customConn)
    }
}

func handleConnection(conn CustomConn) {
    defer conn.Close()
    scanner := bufio.NewScanner(conn)
    if scanner.Scan() {
        conn.tipo = scanner.Text() // Define o tipo com base na primeira mensagem
    }
    fmt.Fprintln(conn, "Conexão estabelecida com sucesso!")
	fmt.Println("Cliente ",conn.tipo,"conectado:", conn.RemoteAddr())
    for scanner.Scan() {
        text := scanner.Text()
		if text == "exit" {
			fmt.Println("Cliente",conn.RemoteAddr(), "desconectado")
			fmt.Fprintln(conn, "Desconectando...")
			break
		}
		if text != "" {
        	fmt.Println("Recebido do cliente",conn.RemoteAddr(),":", text)
        	fmt.Fprintln(conn, "Servidor recebeu:", text)
		}
    }
}