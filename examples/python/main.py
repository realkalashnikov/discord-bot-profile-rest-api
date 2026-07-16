import base64
import json
import mimetypes
import os
import urllib.request
import urllib.error

# Configurações da API do Discord
API_BASE_URL = "https://discord.com/api/v10"
BOT_TOKEN = os.environ.get("DISCORD_BOT_TOKEN", "SEU_BOT_TOKEN_AQUI")
GUILD_ID = os.environ.get("DISCORD_GUILD_ID", "SEU_GUILD_ID_AQUI")

# Cabeçalhos padrão requeridos pela API do Discord
HEADERS = {
    "Authorization": f"Bot {BOT_TOKEN}",
    "Content-Type": "application/json",
    "User-Agent": "DiscordBot (https://github.com/discord-bot-profile-rest-api, 1.0.0)"
}

def image_data_uri(source: str) -> str:
    """
    Converte um caminho de arquivo local ou uma URL de imagem (HTTP/HTTPS)
    em uma string formatada como Data URI codificada em Base64.
    """
    if not source:
        return None

    # Caso a origem seja uma URL da web
    if source.startswith("http://") or source.startswith("https://"):
        try:
            req = urllib.request.Request(
                source, 
                headers={"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"}
            )
            with urllib.request.urlopen(req) as response:
                content_type = response.headers.get("Content-Type")
                if not content_type or not content_type.startswith("image/"):
                    raise ValueError("A URL informada não contém um tipo de imagem válido.")
                
                image_bytes = response.read()
                
                # Validação de tamanho (máximo 10 MB)
                if len(image_bytes) > 10 * 1024 * 1024:
                    raise ValueError("A imagem baixada excede o tamanho máximo de 10 MB.")
                
                encoded_string = base64.b64encode(image_bytes).decode("utf-8")
                mime_type = content_type.split(";")[0]
                return f"data:{mime_type};base64,{encoded_string}"
        except Exception as e:
            print(f"Erro ao baixar e converter a imagem da URL: {e}")
            raise

    # Caso a origem seja um arquivo local
    if not os.path.exists(source):
        raise FileNotFoundError(f"Arquivo local não encontrado: {source}")
        
    mime_type, _ = mimetypes.guess_type(source)
    if not mime_type:
        mime_type = "application/octet-stream"
        
    with open(source, "rb") as image_file:
        image_bytes = image_file.read()
        encoded_string = base64.b64encode(image_bytes).decode("utf-8")
        
    return f"data:{mime_type};base64,{encoded_string}"

def send_patch_request(url: str, payload: dict):
    """ Envia uma requisição PATCH para a API do Discord. """
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(url, data=data, headers=HEADERS, method="PATCH")
    
    try:
        with urllib.request.urlopen(req) as response:
            response_data = response.read().decode("utf-8")
            return json.loads(response_data)
    except urllib.error.HTTPError as e:
        error_msg = e.read().decode("utf-8")
        print(f"Erro HTTP {e.code}: {e.reason}")
        print(f"Resposta da API: {error_msg}")
        raise
    except urllib.error.URLError as e:
        print(f"Erro de conexão: {e.reason}")
        raise

def update_global_profile(username: str = None, avatar_source: str = None, banner_source: str = None, bio: str = None):
    """ Atualiza o perfil global do bot (PATCH /users/@me). """
    print("Iniciando atualização do perfil global...")
    payload = {}
    
    if username is not None:
        payload["username"] = username
    if avatar_source is not None:
        payload["avatar"] = image_data_uri(avatar_source) if avatar_source else None
    if banner_source is not None:
        payload["banner"] = image_data_uri(banner_source) if banner_source else None
    if bio is not None:
        payload["bio"] = bio

    url = f"{API_BASE_URL}/users/@me"
    result = send_patch_request(url, payload)
    print("Perfil global atualizado com sucesso!")
    return result

def update_guild_profile(guild_id: str, nick: str = None, avatar_source: str = None, banner_source: str = None, bio: str = None):
    """ Atualiza o perfil do bot em um servidor específico (PATCH /guilds/{guild_id}/members/@me). """
    print(f"Iniciando atualização do perfil no servidor {guild_id}...")
    payload = {}
    
    if nick is not None:
        payload["nick"] = nick
    if avatar_source is not None:
        payload["avatar"] = image_data_uri(avatar_source) if avatar_source else None
    if banner_source is not None:
        payload["banner"] = image_data_uri(banner_source) if banner_source else None
    if bio is not None:
        payload["bio"] = bio

    url = f"{API_BASE_URL}/guilds/{guild_id}/members/@me"
    result = send_patch_request(url, payload)
    print("Perfil no servidor atualizado com sucesso!")
    return result

if __name__ == "__main__":
    # Exemplo prático de uso
    try:
        # Exemplo 1: Atualização global usando imagem remota
        # update_global_profile(
        #     username="BotExemploPython",
        #     bio="Desenvolvido com scripts limpos em Python nativo.",
        #     avatar_source="https://i.imgur.com/8Km9t74.png"
        # )

        # Exemplo 2: Atualização do perfil específico no servidor
        if GUILD_ID and GUILD_ID != "SEU_GUILD_ID_AQUI":
            update_guild_profile(
                guild_id=GUILD_ID,
                nick="Ajudante em Python",
                bio="Perfil personalizado para este servidor usando Python.",
                avatar_source="https://i.imgur.com/8Km9t74.png"
            )
        else:
            print("Aviso: Configure a variável DISCORD_GUILD_ID para executar o teste em guilda.")
            
    except Exception as e:
        print(f"Não foi possível completar as alterações: {e}")
