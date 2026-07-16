using System;
using System.IO;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;

namespace DiscordBotProfileExample
{
    class Program
    {
        private static readonly string ApiBaseUrl = "https://discord.com/api/v10";
        private static readonly string BotToken = Environment.GetEnvironmentVariable("DISCORD_BOT_TOKEN") ?? "SEU_BOT_TOKEN_AQUI";
        private static readonly string GuildId = Environment.GetEnvironmentVariable("DISCORD_GUILD_ID") ?? "SEU_GUILD_ID_AQUI";

        static async Task Main(string[] args)
        {
            using var httpClient = new HttpClient();
            
            // Configuração dos cabeçalhos obrigatórios exigidos pela API do Discord
            httpClient.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bot", BotToken);
            httpClient.DefaultRequestHeaders.UserAgent.ParseAdd("DiscordBot (https://github.com/discord-bot-profile-rest-api, 1.0.0)");

            try
            {
                // URL da imagem de exemplo ou caminho do arquivo local
                string imageSource = "https://i.imgur.com/8Km9t74.png";
                string? dataUri = null;

                try
                {
                    dataUri = await ImageDataUriAsync(httpClient, imageSource);
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"Aviso: Não foi possível obter a imagem em '{imageSource}': {ex.Message}. Prosseguindo sem imagem.");
                }

                // 1. Atualizar Perfil Global (PATCH /users/@me)
                Console.WriteLine("Atualizando perfil global...");
                var globalPayload = new
                {
                    username = "BotExemploCS",
                    avatar = dataUri,
                    banner = dataUri,
                    bio = "Desenvolvido utilizando C# com HttpClient nativo."
                };
                
                var globalResult = await SendPatchAsync(httpClient, $"{ApiBaseUrl}/users/@me", globalPayload);
                Console.WriteLine("Perfil global atualizado com sucesso!");
                Console.WriteLine($"Resposta da API: {globalResult}");

                // 2. Atualizar Perfil no Servidor (PATCH /guilds/{guild_id}/members/@me)
                if (!string.IsNullOrEmpty(GuildId) && GuildId != "SEU_GUILD_ID_AQUI")
                {
                    Console.WriteLine($"\nAtualizando perfil no servidor: {GuildId}...");
                    var guildPayload = new
                    {
                        nick = "Ajudante em C#",
                        avatar = dataUri,
                        banner = (string?)null, // Exemplo de redefinição/descarte do banner
                        bio = "Perfil personalizado para este servidor em C#."
                    };
                    
                    var guildResult = await SendPatchAsync(httpClient, $"{ApiBaseUrl}/guilds/{GuildId}/members/@me", guildPayload);
                    Console.WriteLine("Perfil no servidor atualizado com sucesso!");
                    Console.WriteLine($"Resposta da API: {guildResult}");
                }
                else
                {
                    Console.WriteLine("\n[INFO] Configure a variável de ambiente DISCORD_GUILD_ID para executar o teste de guilda.");
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Ocorreu um erro no processamento: {ex.Message}");
            }
        }

        /// <summary>
        /// Converte um caminho de arquivo local ou uma URL (HTTP/HTTPS) em uma string Base64 Data URI.
        /// </summary>
        private static async Task<string> ImageDataUriAsync(HttpClient client, string source)
        {
            if (string.IsNullOrEmpty(source))
            {
                throw new ArgumentException("A origem da imagem não pode ser nula ou vazia.");
            }

            byte[] fileBytes;
            string mimeType;

            // Se for uma URL da internet
            if (source.StartsWith("http://") || source.StartsWith("https://"))
            {
                using var response = await client.GetAsync(source);
                if (!response.IsSuccessStatusCode)
                {
                    throw new HttpRequestException($"Erro ao baixar a imagem, status HTTP: {response.StatusCode}");
                }

                var contentType = response.Content.Headers.ContentType?.MediaType;
                if (contentType == null || !contentType.StartsWith("image/"))
                {
                    throw new InvalidDataException("O recurso fornecido na URL não possui um Content-Type de imagem válido.");
                }

                fileBytes = await response.Content.ReadAsByteArrayAsync();
                
                // Validação de tamanho (máximo 10 MB)
                if (fileBytes.Length > 10 * 1024 * 1024)
                {
                    throw new InvalidDataException("A imagem excede o tamanho limite permitido de 10 MB.");
                }

                mimeType = contentType;
            }
            else
            {
                // Se for arquivo local
                if (!File.Exists(source))
                {
                    throw new FileNotFoundException($"Arquivo local não localizado em: {source}");
                }

                fileBytes = await File.ReadAllBytesAsync(source);
                string extension = Path.GetExtension(source).ToLowerInvariant();

                mimeType = extension switch
                {
                    ".png" => "image/png",
                    ".jpg" or ".jpeg" => "image/jpeg",
                    ".gif" => "image/gif",
                    ".webp" => "image/webp",
                    _ => "application/octet-stream"
                };
            }

            string base64String = Convert.ToBase64String(fileBytes);
            return $"data:{mimeType};base64,{base64String}";
        }

        /// <summary>
        /// Realiza a chamada PATCH contendo o corpo JSON correspondente.
        /// </summary>
        private static async Task<string> SendPatchAsync(HttpClient client, string url, object payload)
        {
            string json = JsonSerializer.Serialize(payload);
            using var content = new StringContent(json, Encoding.UTF8, "application/json");
            
            using var response = await client.PatchAsync(url, content);
            string responseBody = await response.Content.ReadAsStringAsync();

            if (!response.IsSuccessStatusCode)
            {
                throw new HttpRequestException($"Erro HTTP {(int)response.StatusCode} ({response.ReasonPhrase}): {responseBody}");
            }

            return responseBody;
        }
    }
}
