# Sistema de Gerenciamento de Recarga para Veículos Elétricos

Este projeto visa resolver a falta de comunicação eficiente entre veículos elétricos e estações de recarga, oferecendo um sistema cliente-servidor baseado em TCP/IP com mensagens estruturadas em JSON. O sistema permite registro de usuários, recomendação de estações, reserva automatizada, pagamento e simulação de trajeto.

---

👥 **Integrantes**  
- **Alisson Silva de Pinho**  
- **Davi Oliveira Santana Carvalho**  
- **Sinval Victor Agostinho Mota**  

*Engenharia da Computação - Universidade Estadual de Feira de Santana (UEFS)*

🛠️ **Tecnologias Utilizadas**  
- **Linguagem**: Go (simplicidade, suporte a sockets e Docker).  
- **Comunicação**: Sockets TCP + mensagens JSON.  
- **Virtualização**: Docker para simulação de múltiplos clientes.  
- **Persistência**: Arquivos JSON (`cars.json` e `stations.json`).
  
## 🛠️ Funcionalidades Principais
- **Registro de Carros e Estações**: Persistência em arquivos JSON (`cars.json` e `stations.json`).
- **Recomendação de Estações**: Baseada em distância (Manhattan) e tempo de espera.
- **Reserva Automatizada**: Evita duplicações e gerencia filas de espera.
- **Pagamento**: Simulação via Pix ou cartão de crédito.
- **Simulação de Trajeto**: Atualização em tempo real de coordenadas e nível de bateria.
- **Gerenciamento de Conexões**: Tratamento de desconexões e sincronização de dados.

---

## 🧩 Entidades do Sistema
1. **Servidor Central**:
   - Gerencia conexões TCP.
   - Processa mensagens (registro, reserva, pagamento).
   - Mantém persistência em JSON e sincroniza recursos.

2. **Cliente Carro**:
   - Simula movimento até a estação.
   - Realiza reservas, pagamentos e consulta histórico.

3. **Cliente Estação**:
   - Registra disponibilidade de pontos de recarga.
   - Notifica o servidor sobre status (ocupado/disponível).

---

## 🚀 Como Executar o Projeto (Docker Compose)

### Pré-requisitos
- Docker instalado.

### **Executando com Imagens do Docker Hub**

Os containers do projeto estão disponíveis no Docker Hub e podem ser executados diretamente sem a necessidade de construir as imagens localmente. Siga os passos abaixo:

1. **Baixar e Executar o Servidor**
   ```bash
   docker run -d --name servidor -p 8080:8080 silvaalisson/server-novo
   ```

2. **Baixar e Executar o Cliente Carro**
   ```bash
   docker run -it --name cliente_car --network="host" silvaalisson/car
   ```

3. **Baixar e Executar o Cliente Estação**
   ```bash
   docker run -it --name cliente_station --network="host" silvaalisson/station
   ```

### **Executando com o Makefile**

O projeto também inclui um Makefile para facilitar a construção e execução dos containers localmente. A partir do diretório raiz do repositório, é necessário:

1. **Construir e Executar o Servidor**
   ```bash
   make build_server
   make run_server
   ```

2. **Construir e Executar o Cliente Carro**
   ```bash
   make build_client_car
   make run_client_car
   ```

3. **Construir e Executar o Cliente Estação**
   ```bash
   make build_client_station
   make run_client_station
   ```
---

📌 **Notas Adicionais**  
- **Limitações**: Persistência em JSON pode limitar escalabilidade. Sugere-se migrar para um banco de dados SQL no futuro.  
- **Melhorias Futuras**: Interface gráfica, APIs REST ou integração com serviços em nuvem.