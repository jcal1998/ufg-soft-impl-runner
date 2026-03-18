# Especificação Detalhada: Assinador CLI (`assinatura`)

## 1. Visão Geral

O **Assinador CLI** (distribuído como o executável `assinatura`) é o principal ponto de contato do usuário com o subsistema de assinaturas digitais do Sistema Runner. Ele atua como um orquestrador leve e multiplataforma, responsável por abstrair a complexidade do ambiente Java e facilitar a interação com as regras de negócio criptográficas contidas no `assinador.jar`.

## 2. Objetivos

* **Abstração de Ambiente:** Garantir que o usuário final não precise saber instalar, configurar ou gerenciar variáveis de ambiente do Java (JDK/JRE).
* **Usabilidade:** Fornecer uma interface semântica baseada em comandos e *flags* (padrão POSIX) para operações de criação e validação de assinaturas.
* **Orquestração de Processos:** Gerenciar o ciclo de vida do backend Java (`assinador.jar`), decidindo entre invocações efêmeras (*Cold Start*) ou contínuas (*Warm Start* / Modo Servidor).
* **Integração:** Fornecer saídas de dados estruturadas e previsíveis (como JSON) para facilitar a automação por outros sistemas ou scripts de teste.

## 3. Requisitos Funcionais (Histórias de Usuário)

### US-01: Invocar Assinador via CLI
**Como** usuário do Sistema Runner  
**Quero** executar comandos de assinatura digital através de uma interface de terminal semântica  
**Para que** eu possa criar e validar assinaturas sem conhecer os comandos ou parâmetros internos da JVM.

**Critérios de Aceitação:**
- [ ] O CLI deve possuir os subcomandos explícitos `criar` e `validar`.
- [ ] O CLI deve aceitar o caminho do arquivo de entrada via flag (ex: `--documento`).
- [ ] O CLI deve capturar a saída de execução do `assinador.jar` e exibi-la formatada no terminal.
- [ ] O CLI deve prover ajuda contextual via flag `--help` global e por subcomando.

### US-02: Provisionar JDK Isolado Automaticamente
**Como** usuário do Sistema Runner  
**Quero** que o CLI baixe e isole um JDK compatível apenas quando necessário  
**Para que** eu não precise instalar o Java manualmente, nem corra o risco do CLI alterar ou conflitar com a versão do Java instalada no meu sistema operacional.

**Critérios de Aceitação:**
- [ ] O CLI deve detectar se o Java já está disponível na versão exigida.
- [ ] Se ausente, deve realizar o download de uma distribuição confiável (ex: Eclipse Temurin JRE/JDK) em *background*.
- [ ] O Java baixado deve ser armazenado em um diretório oculto exclusivo da ferramenta (ex: `~/.runner/jdk/`).
- [ ] O CLI deve injetar a variável `JAVA_HOME` apontando para este diretório isolado **apenas** no contexto da invocação do `assinador.jar`.

### US-03: Gerenciar o Modo Servidor (Warm Start)
**Como** sistema integrador com necessidade de alta performance  
**Quero** poder iniciar e parar o `assinador.jar` em modo *daemon* (servidor HTTP em *background*)  
**Para que** o CLI possa processar múltiplas requisições de assinatura com baixa latência, evitando o custo de inicialização repetida da JVM (*cold start*).

**Critérios de Aceitação:**
- [ ] O CLI deve possuir subcomandos de ciclo de vida: `daemon start`, `daemon stop` e `daemon status`.
- [ ] Ao executar `assinatura criar` ou `assinatura validar`, o CLI deve verificar se o daemon local está ativo (via ping em porta pré-definida, ex: `8080`).
- [ ] Se o daemon estiver ativo, o CLI deve enviar os parâmetros via requisição HTTP (REST/JSON).
- [ ] Se o daemon não estiver ativo, o CLI deve realizar a invocação direta tradicional via linha de comando (`java -jar`).

### US-04: Saída Estruturada e Tratamento de Erros
**Como** engenheiro de automação / desenvolvedor de CI  
**Quero** que o CLI retorne resultados em formatos estruturados e trate erros de forma elegante  
**Para que** eu possa integrar a ferramenta facilmente em *pipelines* e *scripts* de validação.

**Critérios de Aceitação:**
- [ ] Suporte à flag `--output=json` para forçar a saída do terminal em JSON válido.
- [ ] Exceções originadas no `assinador.jar` não devem vazar como *Stack Traces* crus para o usuário, devendo ser parseadas em mensagens concisas (ex: "Erro de Validação FHIR: Campo X ausente").
- [ ] Retornar *Exit Codes* apropriados (ex: `0` para sucesso, `1` para erro de validação, `2` para erro de ambiente).

## 4. Requisitos Não Funcionais (Qualidade - ISO 25010)

1. **Portabilidade:** Os binários devem ser compilados de forma estática, sem dependências dinâmicas complexas, suportando nativamente:
   - Windows (`amd64`)
   - Linux (`amd64`)
   - macOS (`amd64` e `arm64` / Apple Silicon).
2. **Desempenho (Latência):**
   - No modo *Warm Start* (via HTTP), o *overhead* introduzido pelo roteamento do CLI não deve ultrapassar **50ms** somados ao tempo de resposta do Java.
3. **Segurança (Supply Chain):** Todos os artefatos de release do CLI devem ser assinados criptograficamente utilizando **Cosign/Sigstore**, e a verificação deve ser parte integrante da documentação de instalação.

## 5. Design da Interface de Linha de Comando (CLI)

A estrutura de comandos proposta para a aplicação `assinatura`:

```bash
# Ajuda e informações da ferramenta
$assinatura --help$ assinatura version

# ---------------------------------------------
# 1. Operações Core de Assinatura
# ---------------------------------------------
# Cria uma assinatura simulada para um documento
$ assinatura criar --documento ./receita.json --token true --output json

# Valida uma assinatura existente
$ assinatura validar --documento ./receita_assinada.json

# ---------------------------------------------
# 2. Gerenciamento do Backend (Modo Servidor)
# ---------------------------------------------
# Inicia o assinador.jar em background aguardando requisições HTTP
$ assinatura daemon start --port 8080

# Verifica se o daemon está rodando e seu consumo de memória
$ assinatura daemon status

# Encerra o processo do daemon
$ assinatura daemon stop

# ---------------------------------------------
# 3. Diagnóstico e Ambiente
# ---------------------------------------------
# Exibe informações sobre o JDK isolado e os caminhos dos binários
$ assinatura env info

# Força a atualização do JDK ou do assinador.jar associado
$ assinatura env update