FROM golang:1.24.0-alpine AS builder

# Instalando dependências de build
RUN apk add --no-cache tzdata git build-base

# --- CONFIGURAÇÃO ORIGINAL DO OUTRO PROJETO ---
#WORKDIR /go/src/onm-funnel-builder
#COPY ./core /go/src/onm-funnel-builder/core
#COPY ./apis/commons /go/src/onm-funnel-builder/apis/commons
#COPY ./apis/admin /go/src/onm-funnel-builder/apis/admin
#WORKDIR /go/src/onm-funnel-builder/apis/admin
#RUN go clean --modcache && \
#    go get -d && \
#    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o build/main /go/src/onm-funnel-builder/apis/admin
# --- FIM CONFIGURAÇÃO ORIGINAL ---

# --- CONFIGURAÇÃO PARA O PROJETO ATUAL ---
WORKDIR /build
COPY . .
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main main.go
# --- FIM CONFIGURAÇÃO PROJETO ATUAL ---

# Imagem final
FROM alpine:latest

# Configurações de ambiente
ENV TZ=America/Sao_Paulo \
    GOLOG_LOG_LEVEL=info \
    GOTRACEBACK=all 
    
# Instalando dependências necessárias
RUN apk add --no-cache postgresql-client tzdata ca-certificates \
    chromium \
    nss \
    freetype \
    harfbuzz \
    ttf-freefont \
    font-noto \
    font-noto-cjk \
    fontconfig \
    dumb-init && \
    fc-cache -f
# Instalando Chromium para geração de thumbnails

# Criando usuário não-root
RUN adduser -D -u 1000 appuser

# Configurando diretório de trabalho
WORKDIR /app

# --- COPY ORIGINAL DO OUTRO PROJETO ---
#COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
#COPY --from=builder /go/src/onm-funnel-builder/apis/admin/build/main /usr/bin/main
#COPY --from=builder /go/src/onm-funnel-builder/core/errors/errors.json /app/errors/errors.json
#COPY --from=builder /go/src/onm-funnel-builder/core/database/postgres/migration /app/database/migration
# --- FIM COPY ORIGINAL ---

# --- COPY PARA O PROJETO ATUAL ---
COPY --from=builder /build/main /usr/bin/main
# --- FIM COPY PROJETO ATUAL ---

# Ajustando permissões
#RUN chown -R appuser:appuser /app

# Mudando para usuário não-root
#USER appuser

# Expondo a porta
EXPOSE 8081

# Comando para iniciar o aplicativo
CMD ["/usr/bin/main"]
