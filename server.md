# Explicação Detalhada do Servidor (`server.go`)

Este documento explica a lógica central do servidor, focando nos componentes críticos e na comunicação com o cliente.

---

## Variáveis Globais e Controle de Concorrência

### Declaração de Variáveis
```go
var (
    cars     []Car       // Armazena todos os carros do arquivo cars.json
    stations []Station   // Armazena todos os postos do arquivo posts.json
    mu       sync.Mutex  // Mutex para controle de acesso concorrente
)
```

#### Por que foram usadas?
- **`cars` e `stations`**: São *slices* que armazenam os dados em memória para acesso rápido.  
- **`mu` (Mutex)**: Garante que apenas uma goroutine (conexão de cliente) acesse/modifique os dados por vez, evitando **condições de corrida**.

---

## Uso do `sync.Mutex`

O `sync.Mutex` é usado em operações críticas onde os dados globais (`cars` e `stations`) são lidos ou modificados.

### Exemplo 1: Autenticação
```go
mu.Lock() // Bloqueia o acesso às variáveis
var authCar *Car
for i := range cars {
    if cars[i].User == creds[0] && cars[i].Password == creds[1] {
        authCar = &cars[i]
        break
    }
}
mu.Unlock() // Libera o acesso
```
- **Protege**: A busca pelo carro durante a autenticação.
- **Motivo**: Evita que outro cliente modifique a lista `cars` durante a verificação.

### Exemplo 2: Reserva de Posto
```go
mu.Lock()
target.IsReserved = true      // Modifica o posto
authCar.ReservedStation = stationID // Modifica o carro
saveStations()                // Salva no JSON
saveCars()                    // Salva no JSON
mu.Unlock()
```
- **Protege**: A modificação simultânea de `stations` e `cars`.
- **Motivo**: Garante que reservas sejam atômicas (ou acontecem completamente ou não acontecem).

---

## Lógica de Comunicação Cliente-Servidor

O servidor segue um modelo **request-response** via TCP, onde:
1. O cliente envia um comando (ex: `recommend` ou `reserve 3`).
2. O servidor processa o comando.
3. O servidor envia uma resposta (ex: `Posto recomendado: ID 1, Nome Posto A`).

### Fluxo de Processamento
1. **Conexão Aceita**: Cada cliente é atendido em uma goroutine separada (`go handleConnection(conn)`).
2. **Autenticação**:  
   - Cliente envia `user:password`.  
   - Servidor verifica em `cars` e responde com sucesso ou falha.
3. **Comandos**:  
   - `recommend`: Calcula o posto mais próximo usando coordenadas do carro.  
   - `reserve [ID]`: Marca o posto como reservado e atualiza o carro.  
   - `exit`: Encerra a conexão.

---

## Estrutura do Servidor

### 1. Carregamento de Dados (`loadData()`)
- **O que faz**: Lê `cars.json` e `posts.json` para as variáveis `cars` e `stations`.  
- **Quando é chamado**: Na inicialização do servidor (`main()`).

### 2. Salvamento de Dados (`saveCars()` e `saveStations()`)
- **O que fazem**: Convertem `cars` e `stations` para JSON e salvam nos arquivos.  
- **Quando são chamadas**: Após qualquer modificação (ex: reserva de posto).

### 3. Lógica de Recomendação (`findNearestStation()`)
```go
dist := calculateDistance(car.CoordX, car.CoordY, station.CoordX, station.CoordY)
```
- **Algoritmo**: Distância euclidiana entre o carro e cada posto.  
- **Filtro**: Ignora postos já reservados.

---

## Diagrama Simplificado da Comunicação

```
Cliente (TCP)               Servidor
|--------connect----------->|
| "user:password"          ->| (Autentica)
| "recommend"              ->| (Calcula posto)
| "Posto recomendado: ..." <-| 
| "reserve 1"              ->| (Reserva posto)
| "Posto 1 reservado!"     <-|
|--------exit--------------->|
```

---

## Por que Essas Escolhas?

1. **Mutex (`sync.Mutex`)**:  
   - Garante **consistência** dos dados quando múltiplos clientes operam simultaneamente.  
   - Evita corrupção de dados em operações de leitura/escrita.

2. **Variáveis Globais**:  
   - Mantêm o estado do sistema em memória para acesso rápido.  
   - Persistem em JSON para recuperação após reinicialização.

3. **TCP**:  
   - Protocolo confiável para comunicação cliente-servidor.  
   - Garante que comandos e respostas cheguem em ordem.

---

Este design permite que o servidor atenda múltiplos clientes de forma segura e eficiente, garantindo integridade dos dados. 🛠️🔒