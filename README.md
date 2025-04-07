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
- Docker e Docker Compose instalados.

### Passos:
1. **Clone o repositório**:
   ```bash
   git clone [URL_DO_REPOSITÓRIO]
   cd [NOME_DO_DIRETÓRIO]

📌 **Notas Adicionais**  
- **Limitações**: Persistência em JSON pode limitar escalabilidade. Sugere-se migrar para um banco de dados SQL no futuro.  
- **Melhorias Futuras**: Interface gráfica, APIs REST ou integração com serviços em nuvem.  
