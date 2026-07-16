use base64::{engine::general_purpose::STANDARD, Engine as _};
use serde::Serialize;
use std::env;
use std::fs;
use std::path::Path;

// GlobalProfilePayload define a estrutura para alterar o perfil global do bot.
#[derive(Serialize)]
struct GlobalProfilePayload {
    #[serde(skip_serializing_if = "Option::is_none")]
    username: Option<String>,
    
    // Usando Option<Option<T>> para suportar a semântica do PATCH:
    // - None: Campo omitido (não altera)
    // - Some(None): Campo definido como null (remove avatar/banner)
    // - Some(Some(val)): Campo definido com a string base64
    #[serde(skip_serializing_if = "Option::is_none")]
    avatar: Option<Option<String>>,
    
    #[serde(skip_serializing_if = "Option::is_none")]
    banner: Option<Option<String>>,
    
    #[serde(skip_serializing_if = "Option::is_none")]
    bio: Option<String>,
}

// GuildProfilePayload define a estrutura para alterar o perfil no servidor.
#[derive(Serialize)]
struct GuildProfilePayload {
    #[serde(skip_serializing_if = "Option::is_none")]
    nick: Option<String>,
    
    #[serde(skip_serializing_if = "Option::is_none")]
    avatar: Option<Option<String>>,
    
    #[serde(skip_serializing_if = "Option::is_none")]
    banner: Option<Option<String>>,
    
    #[serde(skip_serializing_if = "Option::is_none")]
    bio: Option<String>,
}

/// Converte um caminho de arquivo local ou uma URL (HTTP/HTTPS) em uma string Base64 Data URI.
async fn image_data_uri(
    client: &reqwest::Client,
    source: &str,
) -> Result<String, Box<dyn std::error::Error>> {
    if source.starts_with("http://") || source.starts_with("https://") {
        let response = client.get(source).send().await?;
        if !response.status().is_success() {
            return Err(format!(
                "Falha ao baixar imagem, status HTTP: {}",
                response.status()
            )
            .into());
        }

        // Validação do content-type
        let content_type = response
            .headers()
            .get("content-type")
            .and_then(|val| val.to_str().ok())
            .unwrap_or("image/png");

        if !content_type.starts_with("image/") {
            return Err("O recurso na URL especificada não é uma imagem válida.".into());
        }

        let bytes = response.bytes().await?;
        
        // Limite padrão de 10 MB
        if bytes.len() > 10 * 1024 * 1024 {
            return Err("A imagem baixada excede o limite máximo de 10 MB.".into());
        }

        let mime_type = content_type.split(';').next().unwrap_or("image/png");
        let base64_data = STANDARD.encode(&bytes);
        Ok(format!("data:{};base64,{}", mime_type, base64_data))
    } else {
        // Tratamento para arquivo local
        let path = Path::new(source);
        let bytes = fs::read(path)?;

        let mime_type = match path.extension().and_then(|ext| ext.to_str()) {
            Some("png") => "image/png",
            Some("jpg") | Some("jpeg") => "image/jpeg",
            Some("gif") => "image/gif",
            Some("webp") => "image/webp",
            _ => "image/png",
        };

        let base64_data = STANDARD.encode(&bytes);
        Ok(format!("data:{};base64,{}", mime_type, base64_data))
    }
}

/// Helper para envio de chamadas PATCH na API do Discord.
async fn send_patch_request<T: Serialize>(
    client: &reqwest::Client,
    url: &str,
    token: &str,
    payload: &T,
) -> Result<(), Box<dyn std::error::Error>> {
    let response = client
        .patch(url)
        .header("Authorization", format!("Bot {}", token))
        .json(payload)
        .send()
        .await?;

    let status = response.status();
    if !status.is_success() {
        let text = response.text().await?;
        return Err(format!("A API respondeu com status de erro {}: {}", status, text).into());
    }

    Ok(())
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let bot_token = match env::var("DISCORD_BOT_TOKEN") {
        Ok(token) => token,
        Err(_) => {
            eprintln!("Erro: A variável de ambiente DISCORD_BOT_TOKEN não foi configurada.");
            std::process::exit(1);
        }
    };

    let client = reqwest::Client::new();
    let image_source = "https://i.imgur.com/8Km9t74.png";

    // Converte a origem da imagem para Base64
    let data_uri = match image_data_uri(&client, image_source).await {
        Ok(uri) => Some(Some(uri)),
        Err(e) => {
            println!(
                "Aviso: Não foi possível carregar a imagem em '{}': {}. O perfil será alterado sem imagem.",
                image_source, e
            );
            None
        }
    };

    // 1. Atualizar Perfil Global
    let global_url = "https://discord.com/api/v10/users/@me";
    let global_payload = GlobalProfilePayload {
        username: Some("BotExemploRust".to_string()),
        avatar: data_uri.clone(),
        banner: data_uri.clone(),
        bio: Some("Desenvolvido de forma performática com Rust.".to_string()),
    };

    println!("Atualizando perfil global do bot...");
    match send_patch_request(&client, global_url, &bot_token, &global_payload).await {
        Ok(_) => println!("Perfil global atualizado com sucesso!"),
        Err(e) => eprintln!("Erro ao atualizar perfil global: {}", e),
    }

    // 2. Atualizar Perfil no Servidor (Guilda)
    let guild_id = match env::var("DISCORD_GUILD_ID") {
        Ok(id) => id,
        Err(_) => {
            println!("Ignorando a atualização em servidor porque DISCORD_GUILD_ID não está definido.");
            return Ok(());
        }
    };

    let guild_url = format!("https://discord.com/api/v10/guilds/{}/members/@me", guild_id);
    let guild_payload = GuildProfilePayload {
        nick: Some("Ajudante em Rust".to_string()),
        avatar: data_uri,
        banner: None,
        bio: Some("Perfil personalizado para este servidor em Rust.".to_string()),
    };

    println!("Atualizando perfil do bot no servidor {}...", guild_id);
    match send_patch_request(&client, &guild_url, &bot_token, &guild_payload).await {
        Ok(_) => println!("Perfil no servidor atualizado com sucesso!"),
        Err(e) => eprintln!("Erro ao atualizar perfil no servidor: {}", e),
    }

    Ok(())
}
