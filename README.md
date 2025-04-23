# HTML2Image Renderer

Um utilitário em Go para renderizar páginas HTML em imagens, com suporte especial para conteúdo dinâmico e iframes.

## Sobre o Projeto

Este projeto consiste em uma ferramenta de linha de comando que permite renderizar arquivos HTML em imagens (PNG/JPEG), utilizando o Chrome/Chromium em modo headless via ChromeDP. É especialmente útil para:

- Gerar thumbnails de páginas HTML
- Criar capturas de tela automatizadas
- Renderizar templates HTML em imagens para compartilhamento
- Capturar previews de páginas com conteúdo dinâmico, incluindo iframes

## Requisitos

- [Go](https://golang.org/dl/) 1.16 ou superior
- [Chrome/Chromium](https://www.google.com/chrome/) instalado no sistema
- Sistema operacional: Windows, Linux ou macOS

## Instalação

1. Clone o repositório ou copie os arquivos do projeto para sua máquina local
2. Navegue até o diretório do projeto
3. Instale as dependências necessárias:

```bash
go get -u github.com/chromedp/chromedp
```

4. Compile o projeto:

```bash
go build
```

## Uso

### Comando Básico

```bash
.\poc-thumbnail.exe -html=caminho/para/seu/arquivo.html -output=imagem.png
```

### Opções Disponíveis

| Opção | Descrição | Padrão |
|-------|-----------|--------|
| `-html` | Caminho para o arquivo HTML a ser renderizado | `index.html` no diretório atual |
| `-output` | Caminho para o arquivo de saída da imagem | `thumbnail.png` |
| `-width` | Largura do viewport em pixels | 1280 |
| `-height` | Altura do viewport em pixels | 720 |
| `-quality` | Qualidade da imagem (0-100) | 90 |
| `-full` | Capturar a página inteira (true) ou apenas a viewport (false) | true |
| `-timeout` | Timeout máximo em segundos | 60 |
| `-headless` | Executar em modo headless (sem interface gráfica) | true |
| `-wait` | Tempo de espera em segundos para carregamento de recursos | 5 |
| `-base64` | Retornar a imagem como string base64 em vez de salvar no arquivo | false |

### Exemplos

#### Capturar um arquivo HTML específico

```bash
.\poc-thumbnail.exe -html="C:\caminho\completo\para\arquivo.html" -output=resultado.png
```

#### Capturar com tamanho de viewport personalizado

```bash
.\poc-thumbnail.exe -html=pagina.html -width=1920 -height=1080 -output=tela-larga.png
```

#### Aguardar mais tempo para recursos carregarem (útil para páginas complexas)

```bash
.\poc-thumbnail.exe -html=pagina-complexa.html -wait=10 -output=resultado.png
```

#### Gerar imagem como string base64 (útil para integração com outros sistemas)

```bash
.\poc-thumbnail.exe -html=pagina.html -base64=true > imagem-base64.txt
```

#### Desativar modo headless para visualizar o processo (útil para depuração)

```bash
.\poc-thumbnail.exe -html=pagina.html -headless=false -output=resultado.png
```

## Lidando com Conteúdo Dinâmico e iFrames

Para páginas HTML com conteúdo dinâmico, especialmente aquelas com iframes de vídeo (como YouTube), recomendamos:

1. Aumentar o tempo de espera usando o parâmetro `-wait`
2. Executar em modo não-headless com `-headless=false` (pode ajudar com certos iframes)
3. Ajustar o timeout global com `-timeout` para páginas muito complexas

## Limitações

- Conteúdo de vídeo em iframes (como YouTube) pode não ser renderizado completamente em modo headless
- Alguns recursos externos podem ser bloqueados por políticas de segurança do navegador
- A renderização pode variar ligeiramente dependendo do sistema operacional e da versão do Chrome

## Integrando em seu Projeto

Você pode chamar este utilitário a partir de outras aplicações para automatizar a geração de imagens:

```go
import (
    "os/exec"
    "log"
)

func GerarThumbnail(htmlPath, outputPath string) error {
    cmd := exec.Command("poc-thumbnail.exe", 
        "-html=" + htmlPath, 
        "-output=" + outputPath,
        "-wait=5")
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("Erro ao gerar thumbnail: %v\nOutput: %s", err, output)
        return err
    }
    
    return nil
}
```

## Licença

[MIT](LICENSE)