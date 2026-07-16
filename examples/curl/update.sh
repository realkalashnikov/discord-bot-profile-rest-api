#!/bin/bash

# Configuração de credenciais via variáveis de ambiente
TOKEN="${DISCORD_TOKEN}"
GUILD_ID="${DISCORD_GUILD_ID}"

# Validação do token de acesso
if [ -z "$TOKEN" ]; then
    echo "Erro: A variável DISCORD_TOKEN não foi configurada."
    exit 1
fi

API_BASE="https://discord.com/api/v10"

# Função utilitária para converter arquivo local ou URL em Base64 Data URI
get_image_data_uri() {
    local source="$1"
    if [ -z "$source" ]; then
        echo "null"
        return
    fi
    
    local base64_data=""
    local mime_type="image/png" # Tipo padrão de fallback

    # Se for uma URL (HTTP/HTTPS)
    if [[ "$source" =~ ^https?:// ]]; then
        # Efetua download silencioso seguindo redirecionamentos
        local temp_file
        temp_file=$(mktemp)
        
        if ! curl -s -L -o "$temp_file" "$source"; then
            echo "Erro ao baixar a imagem da URL." >&2
            rm -f "$temp_file"
            echo "null"
            return 1
        fi
        
        # Tenta identificar o mime-type usando 'file' (Linux/macOS)
        if command -v file >/dev/null 2>&1; then
            local file_mime
            file_mime=$(file --mime-type -b "$temp_file")
            if [[ "$file_mime" =~ ^image/ ]]; then
                mime_type="$file_mime"
            fi
        fi
        
        # Realiza a codificação
        if command -v base64 >/dev/null 2>&1; then
            base64_data=$(base64 < "$temp_file" | tr -d '\n')
        else
            base64_data=$(openssl base64 -in "$temp_file" | tr -d '\n')
        fi
        rm -f "$temp_file"
        
    else
        # Se for arquivo local
        if [ ! -f "$source" ]; then
            echo "null"
            return
        fi
        
        if [[ "$source" == *.jpg || "$source" == *.jpeg ]]; then
            mime_type="image/jpeg"
        elif [[ "$source" == *.gif ]]; then
            mime_type="image/gif"
        elif [[ "$source" == *.webp ]]; then
            mime_type="image/webp"
        fi
        
        if command -v base64 >/dev/null 2>&1; then
            base64_data=$(base64 < "$source" | tr -d '\n')
        else
            base64_data=$(openssl base64 -in "$source" | tr -d '\n')
        fi
    fi
    
    echo "data:${mime_type};base64,${base64_data}"
}

# ==========================================
# 1. Atualizar Perfil Global
# ==========================================
update_global() {
    echo "Iniciando atualização do perfil global..."
    
    # Exemplo utilizando uma URL de imagem remota
    local image_url="https://i.imgur.com/8Km9t74.png"
    local avatar_uri
    avatar_uri=$(get_image_data_uri "$image_url")
    
    # Estruturação do JSON
    local payload
    payload=$(cat <<EOF
{
  "username": "BotExemploCurl",
  "bio": "Configurado inteiramente via chamadas de terminal com cURL.",
  "avatar": "${avatar_uri}"
}
EOF
)

    curl -X PATCH "${API_BASE}/users/@me" \
         -H "Authorization: Bot ${TOKEN}" \
         -H "Content-Type: application/json" \
         -H "User-Agent: DiscordBot (https://github.com/discord-bot-profile-rest-api, 1.0.0)" \
         -d "$payload" \
         -i
}

# ==========================================
# 2. Atualizar Perfil no Servidor (Guilda)
# ==========================================
update_guild() {
    if [ -z "$GUILD_ID" ]; then
        echo "Aviso: A variável DISCORD_GUILD_ID não está definida. O teste no servidor foi ignorado."
        return
    fi
    
    echo "Iniciando atualização do perfil no servidor ${GUILD_ID}..."
    
    local image_url="https://i.imgur.com/8Km9t74.png"
    local avatar_uri
    avatar_uri=$(get_image_data_uri "$image_url")
    
    local payload
    payload=$(cat <<EOF
{
  "nick": "Ajudante em cURL",
  "bio": "Perfil personalizado para este servidor.",
  "avatar": "${avatar_uri}"
}
EOF
)

    curl -X PATCH "${API_BASE}/guilds/${GUILD_ID}/members/@me" \
         -H "Authorization: Bot ${TOKEN}" \
         -H "Content-Type: application/json" \
         -H "User-Agent: DiscordBot (https://github.com/discord-bot-profile-rest-api, 1.0.0)" \
         -d "$payload" \
         -i
}

# Chamada das funções
update_global
update_guild
