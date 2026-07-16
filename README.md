# Discord Bot Profile REST API Guide

Este repositório é um guia prático que explica como alterar a identidade visual de um bot do Discord utilizando diretamente a API REST oficial (versão 10). O diferencial deste material é demonstrar o processo sem depender de wrappers ou bibliotecas de alto nível, permitindo que a lógica seja facilmente portável para qualquer linguagem de programação.

---

## Índice
1. [Conceitos e Endpoints](#conceitos-e-endpoints)
2. [Autenticação](#autenticação)
3. [Codificação de Imagens (Base64 Data URI)](#codificação-de-imagens-base64-data-uri)
   - [Usando arquivos locais](#usando-arquivos-locais)
   - [Usando URLs externas (imagens da web)](#usando-urls-externas-imagens-da-web)
4. [Estrutura dos Payloads](#estrutura-dos-payloads)
5. [Exemplos por Linguagem](#exemplos-por-linguagem)
6. [Observações sobre Rate Limits](#observações-sobre-rate-limits)

---

## Conceitos e Endpoints

O Discord oferece suporte a duas camadas de personalização para bots:
1. **Perfil Global:** A identidade padrão do bot, visível em todas as mensagens diretas e servidores onde ele não tenha personalização local.
2. **Perfil por Servidor (Guilda):** Permite alterar o apelido, biografia, avatar e banner especificamente para um único servidor de sua escolha.

A tabela abaixo resume as rotas utilizadas para cada ação:

| Camada do Perfil | Método HTTP | Endpoint | Campos Suportados |
| :--- | :--- | :--- | :--- |
| **Global** | `PATCH` | `https://discord.com/api/v10/users/@me` | `username`, `avatar`, `banner`, `bio` |
| **Servidor (Guilda)** | `PATCH` | `https://discord.com/api/v10/guilds/{guild_id}/members/@me` | `nick`, `avatar`, `banner`, `bio` |

*Nota: Para modificar o perfil do servidor, o bot necessita da permissão **Alterar Apelido** (`CHANGE_NICKNAME`) no servidor de destino.*

---

## Autenticação

Para se autenticar na API REST, todas as chamadas precisam incluir o cabeçalho `Authorization` com o prefixo `Bot` seguido pelo token da sua aplicação.

```http
Authorization: Bot SEU_TOKEN_AQUI
Content-Type: application/json
```

---

## Codificação de Imagens (Base64 Data URI)

A API do Discord não aceita uploads convencionais (como `multipart/form-data`) ou URLs diretas de imagens nos endpoints de perfil. Qualquer imagem fornecida (seja avatar ou banner) precisa ser enviada no corpo do JSON como uma string formatada no esquema **Base64 Data URI**.

O formato padrão esperado é:
```text
data:<mime_type>;base64,<dados_codificados_em_base64>
```

Os tipos de mídia (MIME types) aceitos pelo Discord incluem `image/jpeg`, `image/png`, `image/gif` e `image/webp`. Para remover uma imagem já definida (redefinindo para o padrão), basta passar o valor `null` no campo.

### Usando arquivos locais
Quando a imagem está salva no mesmo ambiente de execução, o arquivo é lido e codificado diretamente para Base64:
- Identifica-se a extensão do arquivo (ex: `.png` vira `image/png`).
- Lê-se o arquivo binário.
- Os bytes do arquivo são convertidos em uma string Base64.

### Usando URLs externas (imagens da web)
Quando a imagem é carregada da internet:
- O bot realiza uma requisição HTTP GET para baixar a imagem.
- É verificado se o cabeçalho de resposta `Content-Type` é uma imagem válida.
- O corpo da resposta (array buffer de bytes) é lido e convertido em Base64.
### Como gerar a string Base64 manualmente

Caso queira testar a requisição usando um cliente HTTP como Postman, Insomnia ou Bruno, você pode codificar uma imagem local para Base64 diretamente pelo terminal do seu sistema operacional:

- **Linux / macOS (Bash):**
  ```bash
  base64 -w 0 imagem.png
  ```

- **Windows (PowerShell):**
  ```powershell
  [Convert]::ToBase64String([IO.File]::ReadAllBytes("imagem.png"))
  ```

- **Ferramentas Online:**
  Você também pode utilizar sites como `base64-image.de` ou qualquer conversor confiável de imagem para Base64. Lembre-se de concatenar o prefixo `data:image/png;base64,` antes do conteúdo gerado.

---

## Estrutura dos Payloads

Você pode enviar no corpo do JSON apenas os campos específicos que deseja atualizar:

### Payload Global (`PATCH /users/@me`)
```json
{
  "username": "NomeDoBot",
  "bio": "Biografia global padrão.",
  "avatar": "data:image/png;base64,iVBORw0KG...",
  "banner": "data:image/jpeg;base64,/9j/4AAQ..."
}
```

### Payload de Servidor (`PATCH /guilds/{guild_id}/members/@me`)
```json
{
  "nick": "Apelido no Servidor",
  "bio": "Biografia específica exibida neste servidor.",
  "avatar": "data:image/png;base64,iVBORw0KG...",
  "banner": null
}
```

---

## Exemplos por Linguagem

O repositório contém exemplos completos e autoexplicativos em várias linguagens com suporte a conversão de arquivos locais e download de links web:

*   [**cURL (Shell Script)**](./examples/curl/update.sh) - Execução direta no terminal Unix.
*   [**JavaScript (Node.js)**](./examples/javascript/index.js) - Utilizando `fetch` nativo (Node 18+).
*   [**Python**](./examples/python/main.py) - Utilizando a biblioteca padrão `urllib` (sem dependências).
*   [**Go**](./examples/go/main.go) - Utilizando `net/http` e structs.
*   [**Rust**](./examples/rust/src/main.rs) - Utilizando `reqwest` assíncrono e `tokio`.
*   [**C# (.NET)**](./examples/csharp/Program.cs) - Utilizando `HttpClient`.

Para testar as implementações locais, defina as variáveis de ambiente necessárias em seu sistema:
```bash
export DISCORD_TOKEN="seu_token_aqui"
export DISCORD_GUILD_ID="id_do_servidor_aqui"
```

---

## Observações sobre Rate Limits

Alterações de perfil na API do Discord são monitoradas de perto pelos limites de taxa da plataforma:
- Modificações excessivas em um curto período no perfil global (como username ou avatar) causarão erros HTTP `429 Too Many Requests`.
- Desenvolva mecanismos de tratamento que verifiquem o cabeçalho `Retry-After` nas respostas para evitar que a chave da aplicação seja temporariamente bloqueada.
