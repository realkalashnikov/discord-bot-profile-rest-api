import fs from "fs/promises";
import path from "path";

const token = process.env.DISCORD_TOKEN;
const guildId = process.env.DISCORD_GUILD_ID;
const API_BASE = "https://discord.com/api/v10";

/**
 * Converte um arquivo local ou uma URL HTTP/HTTPS em um Data URI codificado em Base64.
 * O Discord espera este formato para upload de imagens (avatar e banner).
 * Formato retornado: data:<mime_type>;base64,<dados_em_base64>
 * 
 * @param {string} source - Caminho do arquivo local ou URL da imagem
 * @returns {Promise<string>} Data URI em string
 */
async function imageDataUri(source) {
  if (!source) return null;

  // Verifica se a origem é uma URL
  if (source.startsWith("http://") || source.startsWith("https://")) {
    try {
      const response = await fetch(source, { redirect: "follow" });
      if (!response.ok) {
        throw new Error(`Não foi possível baixar a imagem da URL (${response.status}).`);
      }

      const mimeType = response.headers.get("content-type")?.split(";")[0];
      if (!mimeType || !mimeType.startsWith("image/")) {
        throw new Error("A URL fornecida não retornou uma imagem válida.");
      }

      const arrayBuffer = await response.arrayBuffer();
      const buffer = Buffer.from(arrayBuffer);

      // Limite sugerido de 10 MB para evitar payloads excessivamente grandes
      if (buffer.byteLength > 10 * 1024 * 1024) {
        throw new Error("A imagem ultrapassa o limite permitido de 10 MB.");
      }

      return `data:${mimeType};base64,${buffer.toString("base64")}`;
    } catch (error) {
      console.error(`Erro ao converter a URL da imagem: ${error.message}`);
      throw error;
    }
  }

  // Tratamento para arquivo local
  try {
    const data = await fs.readFile(source);
    const ext = path.extname(source).toLowerCase();
    
    let mimeType = "image/png";
    if (ext === ".jpg" || ext === ".jpeg") mimeType = "image/jpeg";
    else if (ext === ".gif") mimeType = "image/gif";
    else if (ext === ".webp") mimeType = "image/webp";
    
    return `data:${mimeType};base64,${data.toString("base64")}`;
  } catch (error) {
    console.error(`Erro ao carregar e converter o arquivo local em ${source}: ${error.message}`);
    throw error;
  }
}

/**
 * Atualiza o perfil global do bot.
 */
async function updateGlobalProfile({ username, avatarUrlOrPath, bannerUrlOrPath, bio }) {
  console.log("Iniciando atualização do perfil global...");
  const body = {};

  if (username !== undefined) body.username = username;
  if (bio !== undefined) body.bio = bio;
  
  if (avatarUrlOrPath !== undefined) {
    body.avatar = avatarUrlOrPath ? await imageDataUri(avatarUrlOrPath) : null;
  }
  if (bannerUrlOrPath !== undefined) {
    body.banner = bannerUrlOrPath ? await imageDataUri(bannerUrlOrPath) : null;
  }

  const response = await fetch(`${API_BASE}/users/@me`, {
    method: "PATCH",
    headers: {
      "Authorization": `Bot ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Falha na atualização global (Status ${response.status}): ${errorText}`);
  }

  const data = await response.json();
  console.log("Perfil global atualizado com sucesso.");
  console.log(data);
}

/**
 * Atualiza o perfil de membro do bot em um servidor (guilda) específico.
 */
async function updateGuildProfile(targetGuildId, { nick, avatarUrlOrPath, bannerUrlOrPath, bio }) {
  if (!targetGuildId) {
    throw new Error("É necessário informar o ID do servidor (guildId) para esta operação.");
  }
  
  console.log(`Iniciando atualização do perfil no servidor: ${targetGuildId}...`);
  const body = {};

  if (nick !== undefined) body.nick = nick;
  if (bio !== undefined) body.bio = bio;
  
  if (avatarUrlOrPath !== undefined) {
    body.avatar = avatarUrlOrPath ? await imageDataUri(avatarUrlOrPath) : null;
  }
  if (bannerUrlOrPath !== undefined) {
    body.banner = bannerUrlOrPath ? await imageDataUri(bannerUrlOrPath) : null;
  }

  const response = await fetch(`${API_BASE}/guilds/${targetGuildId}/members/@me`, {
    method: "PATCH",
    headers: {
      "Authorization": `Bot ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Falha na atualização do perfil no servidor (Status ${response.status}): ${errorText}`);
  }

  const data = await response.json();
  console.log("Perfil do servidor atualizado com sucesso.");
  console.log(data);
}

// Fluxo principal de execução
async function main() {
  if (!token) {
    console.error("Erro: A variável de ambiente DISCORD_TOKEN não foi configurada.");
    process.exit(1);
  }

  try {
    // Exemplo 1: Atualização global usando imagem da internet
    // await updateGlobalProfile({
    //   username: "MeuBotExemplo",
    //   bio: "Desenvolvido diretamente por requisições REST da API do Discord.",
    //   avatarUrlOrPath: "https://i.imgur.com/8Km9t74.png" // URL da internet
    // });

    // Exemplo 2: Atualização do perfil dentro de um servidor específico
    if (guildId) {
      await updateGuildProfile(guildId, {
        nick: "Ajudante Especial",
        bio: "Focado em gerenciar este servidor.",
        avatarUrlOrPath: "https://i.imgur.com/8Km9t74.png"
      });
    } else {
      console.log("Aviso: DISCORD_GUILD_ID não está configurado. O exemplo de guilda foi ignorado.");
    }

  } catch (error) {
    console.error("Ocorreu um erro ao processar as requisições:", error.message);
  }
}

if (import.meta.url === `file://${process.argv[1]}`) {
  main();
}
