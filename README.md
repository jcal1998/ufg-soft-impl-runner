# Sistema Runner - Plataforma HubSaúde

![Status](https://img.shields.io/badge/Status-Em%20Desenvolvimento-yellow) ![Go](https://img.shields.io/badge/Go-1.25-blue) ![Java](https://img.shields.io/badge/Java-21%20LTS-red)

O **Sistema Runner** é uma plataforma de interoperabilidade de dados em saúde para a Secretaria de Estado de Saúde de Goiás (SES-GO) / UFG. O projeto facilita a execução de lógicas criptográficas (regras de assinatura digital e validação FHIR) via aplicações CLI (**Zero Install**), eliminando a complexidade de gestão de ambiente (ex: instalação do Java) para o utilizador final.

---

## 🏗️ 1. Padrões de Arquitetura (DDD)

A solução foi separada visando o estrito controle de falhas humanas (UX) versus a rigidez técnica das especificações criptográficas da saúde.

* **Open Host Service (OHS) - `assinador.jar` (Java 21):** O motor que hospeda as regras de negócio de assinatura e validação no padrão SES-GO. Ele expõe um contrato estrito (baseado nos Bundles FHIR) e **não** possui tolerância para erros ou problemas de *input/formatação*. Recebe as informações limpas e atua diretamente.
* **Anti-Corruption Layer (ACL) - CLIs (Go 1.25):** Atuam como escudos de UX, interceptando as interações humanas e traduzindo os inputs (seja por *flags* ou interativamente) num formato Data Transfer Object (DTO) válido para o OHS. Qualquer erro retornado pelo `assinador` (como *Stack Traces* e erros complexos FHIR) será intercetado pelo CLI em Go e traduzido numa mensagem semântica focada no utilizador.

---

## 📦 2. Os Três Componentes Entregáveis

### A) Assinador CLI (`assinatura` - Escrito em Go)
Interface principal do utilizador.
* **Zero Install:** Baixa automaticamente um JDK 21 (ex: Eclipse Temurin JRE) limpo, isolando-o na pasta local `~/.hubsaude/jdk/`.
* **Cold Start vs Warm Start:** O CLI orquestra a aplicação Java de acordo com a latência necessária:
  * *Cold Start:* Invocação tradicional e efémera (`java -jar`) para casos pontuais.
  * *Warm Start:* Inicia o assinador.jar em modo Servidor Web / Daemon (`daemon start`). Em seguida, envia os *inputs* via requisições HTTP REST (rotas `/sign` e `/validate`), baixando drasticamente a latência para processamentos contínuos.

### B) Simulador CLI (`simulador` - Escrito em Go) ✅ **[CONCLUÍDO]**
Focado apenas na monitorização e execução do `simulador.jar`.
* **Transferência automática e Dinâmica:** Faz o *download* da última versão pelo GitHub Releases varrendo a API oficial em tempo real, armazenando-a com controle de integridade e versão em Cache local.
* **Zero Install JDK:** Antes de iniciar o simulador, garante que um JDK isolado e portátil está disponível na máquina.
* **Gestão de Ciclo de Vida Avançado (Edge Cases):**
  * Mantém o registo num ficheiro *Lockfile* / *State JSON* único (`~/.hubsaude/state.json`) contendo PIDs e portos guardados.
  * *Health Checks HTTPs:* Verifica a saúde do servidor simulador Java através de requisições web com pulo de verificação SSL em `https://localhost:8080/health`, evitando criar processos "zombie" (PIDs reaproveitados falsamente positivos do S.O.) ou devolver a tela ao usuário antes da inicialização plena.
  * Verificações rigorosas de TCP Ports (portas em uso) antes do arranque prevenindo `PortInUseException` (`simulador start`, `simulador stop`).

### C) Assinador Backend (`assinador.jar` - Escrito em Java)
* O núcleo de Validação FHIR (via biblioteca como HAPI FHIR) baseada rigorosamente nos *Guidelines* de criação de assinatura digital da SES.GO.
* Possui o *parser* que constrói JWTs, gera Base64Url, trata carimbos de tempo em Timestamp UTC, gerando o pacote tipo **Signature**.
* Suporte oficial para lidar com a placa *SunPKCS11*, abstraindo ficheiros .so/.dll como driver no S.O., atuando como ponte direta para o hardware.

---

## 🛡️ 3. O Contrato de Entrada (Regras de Assinatura)

As submissões do CLI em Go para o Java obrigam os seguintes mapeamentos estritos (passados via *flags* POSIX de Terminal):
* **Ficheiro Alvo:** Um ficheiro JSON FHIR Bundle local mapeado em `--bundle` e `--provenance` apontando os caminhos, para evitar sobrecarga de *buffer* em *strings* gigantescas.
* **Autenticação:** As múltiplas estratégias para leitura da Chave Privada:
  * `PEM`: Exige Chave PKCS#8 (+ senha se criptografada).
  * `PKCS#12`: Exibe o path, alias e senha para descriptografia do conteúdo.
  * `Token` ou `Smartcard (PKCS#11)`: Requer o PIN, o Identificador, sendo tratado fisicamente pelo driver JDBC-like do PKCS11.
  * `Remote`: Endereço de URL e credenciais para consumir o Key Vault remoto.
* **Timestamp & Políticas Temporais:** Input natural transformado pelo CLI em carimbos ISO-8601 estritos (`--timestamp`), juntamente com o modelo de tempo `iat` (Instant) ou `tsa` (Timestamping Authority).
* **Política de Assinatura (sigPId):** Fixada pelo parâmetro `--pid`, aguardando `br.go.ses.seguranca|0.0.2`.
* **Certificado (x5c):** Em formato arquivo referenciado do signatário (`--cert`).

---

## ✨ 4. Experiência de Usuário (UX) & Requisitos Non-Funcionais

* **O Fim da "Ditadura das Flags":** Embora as *flags* existam para *scripts* de Pipeline/Automação de Testes (`--output=json`), utilizadores comuns que invocam o comando "vazio" (`assinatura sign`) passarão por um modo **100% interativo** de Q&A de terminal. Menus selecionáveis interativos, abas com setas de opções e caixas de seleção, ocultando a complexidade.
* **Feedback Visual Constante e Spinners:** Em tarefas com processadores bloqueantes como descarregar Java, conferir Hash Sigstore ICP-Brasil e carregamento de *middleware* Driver .dll, *Spinners* de terminal deverão garantir visibilidade absoluta de estado ativo, nunca dando sensação de congelamento.
* **Hardware Security Keys (Efeito UAU):** Previsto ativamente a integração física de Hardware/Chaves USB (Tokens e Smartcards). O componente Go + Java parará ativamente a execução da interface lançando pop-ups visuais no terminal em espera (Ex: *"Tocar na chave USB piscante para confirmar inserção biométrica!"*).
* **Tratamento Seguro de Exceções:** Sem visibilidade técnica da API Java pelo terminal, encapsulados em JSON formatados.

---

## 🚀 5. Segurança da Pipeline (CI/CD)

* **Build Multiplataforma Efetivo:** Distribuindo os executáveis compilados de forma estática para Windows `amd64.exe`, Linux `amd64.AppImage`, macOS `amd64.dmg` / `arm64.dmg`.
* **Assinaturas Criptográficas / Supply Chain:** Artefatos de release CI/CD (Github Actions) obrigatoriamente selados usando **Cosign / Sigstore**, em conjunto com a identidade *OIDC* do próprio assinador, com anexos *.pem e *.sig no Repositório de Software. Os Hashes do `.zip` contendo ICP-Brasil devem ser estritamente controlados em *Trust Stores* e *Caches*.