package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// GlobalProfilePayload define a estrutura para alterar o perfil global do bot.
type GlobalProfilePayload struct {
	Username string  `json:"username,omitempty"`
	Avatar   *string `json:"avatar,omitempty"` // Data URI em base64 ou null
	Banner   *string `json:"banner,omitempty"` // Data URI em base64 ou null
	Bio      string  `json:"bio,omitempty"`
}

// GuildProfilePayload define a estrutura para alterar o perfil no servidor.
type GuildProfilePayload struct {
	Nick   string  `json:"nick,omitempty"`
	Avatar *string `json:"avatar,omitempty"` // Data URI em base64 ou null
	Banner *string `json:"banner,omitempty"` // Data URI em base64 ou null
	Bio    string  `json:"bio,omitempty"`
}

// imageDataURI aceita tanto um arquivo local quanto uma URL e retorna a string no formato Data URI codificada em Base64.
func imageDataURI(source string) (string, error) {
	if source == "" {
		return "", nil
	}

	var data []byte
	var mimeType string

	// Verifica se a origem é uma URL
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		resp, err := http.Get(source)
		if err != nil {
			return "", fmt.Errorf("falha ao baixar a imagem: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("servidor respondeu com código de erro: %d", resp.StatusCode)
		}

		// Valida o cabeçalho content-type
		contentType := resp.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			return "", errors.New("o recurso na URL fornecida não é uma imagem válida")
		}
		mimeType = strings.Split(contentType, ";")[0]

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("falha ao ler o corpo da resposta: %w", err)
		}

		// Limite sugerido de 10 MB
		if len(data) > 10*1024*1024 {
			return "", errors.New("a imagem excede o tamanho limite de 10 MB")
		}
	} else {
		// Tratamento para arquivo local
		var err error
		data, err = os.ReadFile(source)
		if err != nil {
			return "", fmt.Errorf("falha ao ler arquivo local: %w", err)
		}

		switch filepath.Ext(source) {
		case ".png":
			mimeType = "image/png"
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".gif":
			mimeType = "image/gif"
		case ".webp":
			mimeType = "image/webp"
		default:
			mimeType = "image/png" // Padrão
		}
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded), nil
}

// sendPatchRequest executa a chamada PATCH na API REST do Discord.
func sendPatchRequest(url string, token string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao converter payload para json: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("erro ao estruturar requisição HTTP: %w", err)
	}

	// Discord API requer Authorization com prefixo "Bot"
	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("conexão falhou: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("api do discord retornou código %s: %s", resp.Status, string(respBody))
	}

	return nil
}

func main() {
	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	if botToken == "" {
		fmt.Println("Erro: A variável de ambiente DISCORD_BOT_TOKEN não foi configurada.")
		os.Exit(1)
	}

	// Exemplo: URL de imagem remota ou arquivo local
	imageSource := "https://i.imgur.com/8Km9t74.png"
	dataURI, err := imageDataURI(imageSource)
	if err != nil {
		fmt.Printf("Aviso: Falha ao converter imagem de '%s': %v. A imagem não será enviada.\n", imageSource, err)
	}

	// 1. Atualizar Perfil Global
	globalURL := "https://discord.com/api/v10/users/@me"
	globalPayload := GlobalProfilePayload{
		Username: "BotExemploGo",
		Bio:      "Desenvolvido utilizando Go nativo.",
	}
	if err == nil && dataURI != "" {
		globalPayload.Avatar = &dataURI
	}

	fmt.Println("Atualizando perfil global do bot...")
	err = sendPatchRequest(globalURL, botToken, globalPayload)
	if err != nil {
		fmt.Printf("Falha na atualização global: %v\n", err)
	} else {
		fmt.Println("Perfil global atualizado com sucesso!")
	}

	// 2. Atualizar Perfil de Membro no Servidor (Guilda)
	guildID := os.Getenv("DISCORD_GUILD_ID")
	if guildID == "" {
		fmt.Println("Ignorando atualização no servidor porque DISCORD_GUILD_ID não foi definido.")
		return
	}

	guildURL := fmt.Sprintf("https://discord.com/api/v10/guilds/%s/members/@me", guildID)
	guildPayload := GuildProfilePayload{
		Nick: "Ajudante em Go",
		Bio:  "Perfil personalizado para este servidor em Go.",
	}
	if err == nil && dataURI != "" {
		guildPayload.Avatar = &dataURI
	}

	fmt.Printf("Atualizando perfil do bot no servidor %s...\n", guildID)
	err = sendPatchRequest(guildURL, botToken, guildPayload)
	if err != nil {
		fmt.Printf("Falha na atualização do servidor: %v\n", err)
	} else {
		fmt.Println("Perfil no servidor atualizado com sucesso!")
	}
}
