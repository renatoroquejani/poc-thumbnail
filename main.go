package main

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	_ "poc-thumbnail/docs" // Importação correta para o Swagger
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
)

// @title Thumbnail API
// @version 1.0
// @description API para gerar thumbnails de HTML
// @host localhost:8080
// @BasePath /

// ThumbnailRequest representa os parâmetros para geração do thumbnail
// swagger:model
// @Description Parâmetros para geração de thumbnail
type ThumbnailRequest struct {
	HTMLPath  string `json:"htmlPath"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Quality   int    `json:"quality"`
	Timeout   int    `json:"timeout"`
	Headless  bool   `json:"headless"`
	WaitTime  int    `json:"waitTime"`
	AsBase64  bool   `json:"base64"`
	FullPage  bool   `json:"fullPage"` // Capturar a página inteira (true) ou apenas a viewport (false)
}

// ThumbnailResponse representa a resposta da API
// swagger:model
// @Description Resposta da geração de thumbnail
type ThumbnailResponse struct {
	Base64  string `json:"base64,omitempty"`
	Message string `json:"message,omitempty"`
	Size    int    `json:"size,omitempty"`
}

// GenerateThumbnail godoc
// @Summary Gera um thumbnail a partir de um HTML
// @Accept json
// @Produce json
// @Param request body ThumbnailRequest true "Parâmetros"
// @Success 200 {object} ThumbnailResponse
// @Failure 400 {object} ThumbnailResponse
// @Router /thumbnail [post]
func GenerateThumbnail(c *gin.Context) {
    log.Printf("[DEBUG] Início do GenerateThumbnail")
	var req ThumbnailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
        log.Printf("[DEBUG] Erro ao fazer bind do JSON: %v", err)
		c.JSON(http.StatusBadRequest, ThumbnailResponse{Message: "JSON inválido: " + err.Error()})
		return
	}

    log.Printf("[DEBUG] Parâmetros recebidos: %+v", req)
	if req.HTMLPath == "" {
        dir, err := os.Getwd()
        if err != nil {
			c.JSON(http.StatusBadRequest, ThumbnailResponse{Message: "Erro ao obter diretório: " + err.Error()})
			return
		}
        req.HTMLPath = filepath.Join(dir, "index.html")
        log.Printf("[DEBUG] HTMLPath não informado, usando padrão: %s", req.HTMLPath)
    }

    info, err := os.Stat(req.HTMLPath)
    if os.IsNotExist(err) {
        log.Printf("[DEBUG] Arquivo HTML não existe: %s", req.HTMLPath)
		c.JSON(http.StatusBadRequest, ThumbnailResponse{Message: "O arquivo HTML não existe: " + req.HTMLPath})
		return
	}
    if info.IsDir() {
        log.Printf("[DEBUG] Caminho fornecido é um diretório, não um arquivo: %s", req.HTMLPath)
		c.JSON(http.StatusBadRequest, ThumbnailResponse{Message: "O caminho fornecido é um diretório, não um arquivo: " + req.HTMLPath})
		return
	}

    absPath, err := filepath.Abs(req.HTMLPath)
    if err != nil {
        log.Printf("[DEBUG] Erro ao obter caminho absoluto: %v", err)
		c.JSON(http.StatusBadRequest, ThumbnailResponse{Message: "Erro ao obter caminho absoluto: " + err.Error()})
		return
	}

    log.Printf("[DEBUG] Caminho absoluto do HTML: %s", absPath)
    // Parâmetros padrão
	if req.Width == 0 {
		req.Width = 1280
	}
	if req.Height == 0 {
		req.Height = 720
	}
	if req.Quality == 0 {
		req.Quality = 90
	}
	if req.Timeout == 0 {
		req.Timeout = 60
	}
	if req.WaitTime == 0 {
		req.WaitTime = 5
	}
	if !req.FullPage {
		// Se não for enviado, default é true
		req.FullPage = true
	}

    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.Timeout)*time.Second)
    defer cancel()
    log.Printf("[DEBUG] Contexto de timeout criado: %ds", req.Timeout)

    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("headless", req.Headless),
        chromedp.Flag("no-sandbox", true),
        chromedp.Flag("disable-gpu", true),
        chromedp.Flag("disable-dev-shm-usage", true),
        chromedp.Flag("disable-web-security", true),
        chromedp.Flag("allow-running-insecure-content", true),
    )
    log.Printf("[DEBUG] Criando allocator do ChromeDP")
    allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
    defer allocCancel()

    ctx, cancel = chromedp.NewContext(allocCtx)
    defer cancel()
    log.Printf("[DEBUG] Contexto do ChromeDP criado")

    var buf []byte
    log.Printf("[DEBUG] Iniciando navegação e captura do screenshot... (FullPage: %v)", req.FullPage)
    if req.FullPage {
        err = chromedp.Run(ctx,
            chromedp.Navigate("file://"+absPath),
            chromedp.Sleep(time.Duration(req.WaitTime)*time.Second),
            chromedp.FullScreenshot(&buf, req.Quality),
        )
    } else {
        err = chromedp.Run(ctx,
            chromedp.Navigate("file://"+absPath),
            chromedp.Sleep(time.Duration(req.WaitTime)*time.Second),
            chromedp.CaptureScreenshot(&buf),
        )
    }
    if err != nil {
        log.Printf("[DEBUG] Erro ao rodar chromedp: %v", err)
        c.JSON(http.StatusBadRequest, ThumbnailResponse{Message: "Erro ao gerar thumbnail: " + err.Error()})
        return
    }
    log.Printf("[DEBUG] Screenshot capturado com sucesso, tamanho: %d bytes", len(buf))

    if req.AsBase64 {
        b64Image := base64.StdEncoding.EncodeToString(buf)
        log.Printf("[DEBUG] Retornando imagem em base64 (%d bytes)", len(b64Image))
        c.JSON(http.StatusOK, ThumbnailResponse{
            Base64:  b64Image,
            Size:    len(buf),
            Message: "Imagem gerada com sucesso (base64)",
        })
    } else {
        outputPath := "thumbnail.png"
        if err := os.WriteFile(outputPath, buf, 0644); err != nil {
            log.Printf("[DEBUG] Erro ao salvar arquivo: %v", err)
            c.JSON(http.StatusInternalServerError, ThumbnailResponse{Message: "Erro ao salvar arquivo: " + err.Error()})
            return
        }
        log.Printf("[DEBUG] Imagem salva em %s (%d bytes)", outputPath, len(buf))
        c.JSON(http.StatusOK, ThumbnailResponse{
            Message: "Imagem salva em " + outputPath,
            Size:    len(buf),
        })
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

func main() {
	r := gin.Default()

	r.POST("/thumbnail", GenerateThumbnail)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("Servidor iniciado em http://localhost:8080")
	r.Run(":8080")
}
