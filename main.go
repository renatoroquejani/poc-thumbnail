package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func main() {
	// Define flags para configuração
	htmlPath := flag.String("html", "", "Caminho para o arquivo HTML")
	outputPath := flag.String("output", "thumbnail.png", "Caminho para a imagem de saída (ignorado se base64=true)")
	width := flag.Int("width", 1280, "Largura do viewport em pixels")
	height := flag.Int("height", 720, "Altura do viewport em pixels")
	quality := flag.Int("quality", 90, "Qualidade da imagem (0-100, sendo 100 a melhor qualidade)")
	fullPage := flag.Bool("full", true, "Capturar a página inteira (true) ou apenas a viewport (false)")
	timeout := flag.Int("timeout", 60, "Timeout máximo em segundos")
	headless := flag.Bool("headless", true, "Executar em modo headless (sem interface gráfica)")
	waitTime := flag.Int("wait", 5, "Tempo de espera em segundos para carregamento de recursos")
	asBase64 := flag.Bool("base64", true, "Retornar a imagem como string base64 em vez de salvar no arquivo")
	flag.Parse()

	// Verificar se o caminho do HTML foi fornecido
	if *htmlPath == "" {
		// Se não foi fornecido um caminho específico, assumir que está na pasta do projeto
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		*htmlPath = filepath.Join(dir, "index.html")
	}

	// Verificar se o arquivo HTML existe
	info, err := os.Stat(*htmlPath)
	if os.IsNotExist(err) {
		log.Fatalf("O arquivo HTML não existe: %s", *htmlPath)
	}
	if info.IsDir() {
		log.Fatalf("O caminho fornecido é um diretório, não um arquivo: %s", *htmlPath)
	}

	// Converter o caminho do arquivo para URI de arquivo
	absPath, err := filepath.Abs(*htmlPath)
	if err != nil {
		log.Fatal(err)
	}
	fileURL := fmt.Sprintf("file://%s", filepath.ToSlash(absPath))

	// Configurar as opções do navegador
	opts := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", *headless),
		chromedp.Flag("disable-web-security", true), // Desativa restrições de segurança para carregar iframes
		chromedp.Flag("allow-running-insecure-content", true),
		chromedp.Flag("autoplay-policy", "no-user-gesture-required"), // Permite reprodução automática
		chromedp.Flag("enable-automation", false),                    // Tentar esconder que é automação
		chromedp.WindowSize(*width, *height),
	}

	// Adicionar opções padrão
	opts = append(chromedp.DefaultExecAllocatorOptions[:], opts...)

	// Criar um novo contexto de alocação com as opções definidas
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Criar um contexto com o contexto de alocação
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Adicionar um timeout para evitar bloqueios
	ctx, cancel = context.WithTimeout(ctx, time.Duration(*timeout)*time.Second)
	defer cancel()

	// Executar as tarefas do chromedp
	log.Printf("Renderizando o arquivo HTML: %s", fileURL)
	log.Printf("Configurações: largura=%d, altura=%d, qualidade=%d, páginaCompleta=%v",
		*width, *height, *quality, *fullPage)

	// Navegar para o HTML e aguardar carregamento
	if err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(fileURL),
		// Esperar que o DOM esteja completamente carregado
		chromedp.WaitReady("body", chromedp.ByQuery),
		// Esperar que os iframes sejam carregados (se houver)
		chromedp.EvaluateAsDevTools(`
			Array.from(document.querySelectorAll('iframe')).forEach(iframe => {
				iframe.addEventListener('load', () => {
					iframe.setAttribute('data-loaded', 'true');
				});
				// Para iframes já carregados
				if (iframe.contentDocument && iframe.contentDocument.readyState === 'complete') {
					iframe.setAttribute('data-loaded', 'true');
				}
			});
		`, nil),
		// Dar tempo para iframes carregarem
		chromedp.Sleep(time.Duration(*waitTime) * time.Second),
	}); err != nil {
		log.Fatalf("Erro ao navegar para a página: %v", err)
	}

	// Capturar a screenshot
	var buf []byte

	if *fullPage {
		// Usar função personalizada para screenshots de página completa
		dimensions, err := getPageDimensions(ctx)
		if err != nil {
			log.Fatalf("Erro ao obter dimensões da página: %v", err)
		}

		width := dimensions[0]
		height := dimensions[1]

		// Definir viewport para capturar a página inteira
		err = chromedp.Run(ctx, emulation.SetDeviceMetricsOverride(width, height, 1, false))
		if err != nil {
			log.Fatalf("Erro ao definir dimensões do dispositivo: %v", err)
		}

		// Capturar screenshot
		err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			var captureErr error
			buf, captureErr = page.CaptureScreenshot().
				WithQuality(int64(*quality)).
				WithClip(&page.Viewport{
					X:      0,
					Y:      0,
					Width:  float64(width),
					Height: float64(height),
					Scale:  1,
				}).Do(ctx)
			return captureErr
		}))

		if err != nil {
			log.Fatalf("Erro ao capturar screenshot de página completa: %v", err)
		}
	} else {
		// Capturar apenas a viewport visível
		err = chromedp.Run(ctx, chromedp.CaptureScreenshot(&buf))
		if err != nil {
			log.Fatalf("Erro ao capturar screenshot: %v", err)
		}
	}

	// Processar o resultado
	if *asBase64 {
		// Converter para base64
		b64Image := base64.StdEncoding.EncodeToString(buf)

		// Imprimir string base64 na saída padrão
		fmt.Println(b64Image)

		// Opcionalmente, imprimir tag de imagem HTML completa
		// Determinar o tipo MIME correto
		mimeType := "image/png"
		if *quality < 100 {
			mimeType = "image/jpeg"
		}

		log.Printf("Tamanho da string base64: %d bytes", len(b64Image))
		log.Printf("Para usar em HTML: <img src=\"data:%s;base64,%s\" />", mimeType, b64Image[:20]+"...")
	} else {
		// Salvar a imagem como arquivo
		log.Printf("Salvando a imagem em: %s", *outputPath)
		if err := os.WriteFile(*outputPath, buf, 0644); err != nil {
			log.Fatal(err)
		}

		log.Printf("Imagem gerada com sucesso! Tamanho: %d bytes", len(buf))
	}
}

// getPageDimensions retorna as dimensões da página inteira
func getPageDimensions(ctx context.Context) ([]int64, error) {
	var dimensions []int64
	err := chromedp.Run(ctx, chromedp.EvaluateAsDevTools(`
		[Math.max(
			document.body.scrollWidth,
			document.documentElement.scrollWidth,
			document.body.offsetWidth,
			document.documentElement.offsetWidth,
			document.body.clientWidth,
			document.documentElement.clientWidth
		), Math.max(
			document.body.scrollHeight,
			document.documentElement.scrollHeight,
			document.body.offsetHeight,
			document.documentElement.offsetHeight,
			document.body.clientHeight,
			document.documentElement.clientHeight
		)]
	`, &dimensions))

	if err != nil {
		return nil, err
	}

	return dimensions, nil
}
