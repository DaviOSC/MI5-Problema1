## Estado em Memória vs. Comunicação TCP

### 1. **Estado em Memória (Variáveis Globais)**
As variáveis globais (`cars` e `stations`) são usadas para:
- **Acesso rápido**: Evitar a necessidade de ler os arquivos JSON a cada requisição, o que seria lento.
- **Gerenciamento de estado**: Manter o estado atual do sistema (ex: quais postos estão reservados, qual carro está conectado).

#### Por que isso não "quebra" o modelo de requisição-resposta?
- **Separação de responsabilidades**:
  - O **TCP** lida com a comunicação entre cliente e servidor (envio e recebimento de mensagens).
  - As **variáveis globais** gerenciam o estado do sistema (dados dos carros e postos).
- **Requisições e respostas** continuam sendo feitas via TCP. O estado em memória apenas **otimiza o processamento** dessas requisições.

---

### 2. **Persistência em JSON**
Os arquivos `cars.json` e `posts.json` são usados para:
- **Recuperação de estado**: Se o servidor for reiniciado, os dados são carregados dos arquivos JSON.
- **Backup**: Garantir que os dados não sejam perdidos em caso de falha no servidor.

#### Como isso se relaciona com o TCP?
- **Persistência é independente da comunicação**:  
  O TCP lida com a comunicação em tempo real, enquanto os arquivos JSON servem como um **backup** para o estado do sistema.
- **Exemplo**:  
  Se um cliente reserva um posto, o servidor:
  1. Atualiza o estado em memória (`stations` e `cars`).
  2. Persiste a mudança no arquivo JSON.
  3. Responde ao cliente via TCP que a reserva foi feita.

---

### 3. **Requisição e Resposta via TCP**
O TCP é usado para:
- **Enviar comandos**: O cliente envia comandos como `recommend` ou `reserve 1`.
- **Receber respostas**: O servidor processa o comando e envia uma resposta (ex: `Posto recomendado: ID 1, Nome Posto A`).

#### Exemplo de Fluxo:
1. Cliente envia: `"user:password"` (autenticação).
2. Servidor responde: `"Autenticado com sucesso!"`.
3. Cliente envia: `"recommend"`.
4. Servidor responde: `"Posto recomendado: ID 1, Nome Posto A"`.
5. Cliente envia: `"reserve 1"`.
6. Servidor responde: `"Posto 1 reservado com sucesso!"`.

---

### 4. **Por que não usar apenas o TCP para tudo?**
- **Desempenho**: Ler e escrever em arquivos JSON a cada requisição seria muito lento.
- **Complexidade**: O TCP não foi projetado para gerenciar estado persistente. Ele é um protocolo de comunicação, não um banco de dados.
- **Escalabilidade**: Manter o estado em memória permite que o servidor atenda múltiplos clientes de forma eficiente.

---

### 5. **Alternativas**
Se o objetivo fosse evitar completamente o uso de estado em memória, poderíamos:
- **Usar um banco de dados**: Substituir os arquivos JSON e variáveis globais por um banco de dados (ex: SQLite, PostgreSQL).
- **Implementar um sistema stateless**: O cliente enviaria todos os dados necessários em cada requisição, e o servidor não manteria estado.

No entanto, essas abordagens têm trade-offs:
- **Banco de dados**: Adiciona complexidade e overhead.
- **Stateless**: Exige que o cliente envie mais dados a cada requisição, aumentando o tráfego de rede.

---

## Conclusão

O uso de **variáveis globais** e **persistência em JSON** não quebra a proposta do TCP. Pelo contrário, eles complementam a comunicação TCP, permitindo que o servidor:
1. **Mantenha estado** de forma eficiente.
2. **Persista dados** para recuperação após reinicialização.
3. **Responda rapidamente** às requisições dos clientes.

O TCP continua sendo o responsável pela comunicação em tempo real, enquanto o estado em memória e os arquivos JSON garantem a **integridade** e **eficiência** do sistema.