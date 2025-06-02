## 📦 Projeto: RemoteList RPC com Persistência e Concorrência

Este projeto implementa um sistema distribuído em Go com chamadas RPC síncronas para manipular múltiplas listas de inteiros de forma concorrente e persistente.

### 🚀 Funcionalidades

* Adicionar, consultar, remover e obter tamanho de listas via RPC
* Suporte a múltiplos clientes simultâneos
* Persistência com arquivos de log e snapshot
* Recuperação automática de estado após falhas
* Exclusão mútua com `sync.Mutex` e `sync.RWMutex`

---

### 🗂 Estrutura do Projeto

```
rpc-list/
├── main.go                  # Inicia o servidor RPC
├── client/
│   └── client.go            # Cliente de exemplo
└── server/
    ├── remote_list.go       # Lógica principal do RemoteList
    ├── persistence.go       # Persistência (log/snapshot)
    └── types.go             # Tipos auxiliares
```

---

### ▶️ Como Executar

#### 1. Iniciar o Servidor

```bash
go run main.go
```

Servidor escutará na porta `:1234`.

#### 2. Rodar o Cliente

```bash
go run client/client.go
```

Você verá saídas como:

```
Append: Valor adicionado.
Size: 1
Get index 0: 42
Remove: 42
```

---

### 📁 Persistência

* `snapshot.json`: armazena o estado completo das listas periodicamente.
* `log.jsonl`: armazena todas as operações (append/remove).

No reinício do servidor, ele recupera o estado a partir do snapshot e aplica operações pendentes do log.

---

### 🧠 Discussão: Limitações e Escalabilidade

* **Consistência**: O sistema garante consistência por meio do uso de mecanismos de exclusão mútua, utilizando mutexes específicos para cada lista e um mutex global durante a criação de snapshots. Dessa forma, evita-se condições de corrida e inconsistências nos dados durante acessos simultâneos por múltiplos clientes.

* **Escalabilidade**: A escalabilidade do sistema é limitada, uma vez que se trata de uma arquitetura monolítica com persistência baseada em arquivos locais. Para permitir um crescimento mais eficiente, seria possível adotar estratégias como particionamento das listas (sharding), utilização de sistemas de armazenamento distribuído (como etcd, Redis ou Cassandra) e balanceamento de carga entre instâncias do serviço.

* **Disponibilidade**: Atualmente, o sistema possui um ponto único de falha, já que depende de uma única instância do servidor. Para aumentar a disponibilidade e tolerância a falhas, seria recomendada a implementação de replicação ativa/passiva, permitindo que instâncias secundárias assumam o controle em caso de falha da principal.

* **Falhas**: O sistema realiza a recuperação do estado das listas utilizando snapshots periódicos e registros de log das operações. Essa abordagem garante persistência mesmo após falhas do servidor. No entanto, caso ocorra a perda do arquivo de log após o último snapshot, existe a possibilidade de perda das operações realizadas nesse intervalo. Estratégias de backup e replicação poderiam mitigar esse risco.